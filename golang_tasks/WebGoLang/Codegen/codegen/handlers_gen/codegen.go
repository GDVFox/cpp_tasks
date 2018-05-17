package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

var globalAuthKey = `"100500"`

type ApiValidateStruct struct {
	Name   string
	Fields []*StructField
}

type StructField struct {
	Type         string
	CodeName     string
	ParamName    string
	Required     bool
	Enum         []string
	DefaultValue string
	Min          int
	Max          int
}

type MethodConfig struct {
	Name           string
	ReceiverName   string
	ValidateStruct *ApiValidateStruct
	URL            string
	Method         string
	AuthKey        string
	Auth           bool
}

type ApiStruct struct {
	Name    string
	Methods map[string]*MethodConfig
}

var (
	serveHTTPTpl = template.Must(template.New("serveHTTPTpl").Parse(`
func (h *{{.Name}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	{{range .Methods}}case "{{.URL}}":
		h.handle{{.Name}}(w, r)
	{{end}}default:
		w.WriteHeader(http.StatusNotFound)
		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : "unknown method"})
		w.Write(jsonRes)
	}
}

`))
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ./codegen <file_for_parsing.go> <result_file.go>")
		return
	}

	node, err := parser.ParseFile(token.NewFileSet(), os.Args[1],
		nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("parser error: %v", err)
	}

	resultFile, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatalf("result file error: %v", err)
	}

	//Сначала нужно распарсить входные данные
	apiStructs := make(map[string]*ApiStruct)                 //Сюда складываем структуры, для которых нужно генерировать методы
	apiValidateStructs := make(map[string]*ApiValidateStruct) //Сюда складываем структуры для валидации
	for _, decl := range node.Decls {
		if fDecl, ok := decl.(*ast.FuncDecl); !ok {
			log.Printf("skip: is not *ast.FuncDecl\n")
		} else { //Парсим метод
			if fDecl.Doc == nil {
				log.Printf("skip: method %s does not have comments\n",
					fDecl.Name.Name)
				continue
			}

			var methodConf *MethodConfig
			for _, comment := range fDecl.Doc.List {
				if strings.HasPrefix(comment.Text, "// apigen:api") {
					methodConf = &MethodConfig{}
					if err := json.Unmarshal([]byte(strings.TrimPrefix(comment.Text, "// apigen:api")),
						methodConf); err != nil {
						log.Fatalf("unmarshal error in %s: %v", comment.Text, err)
					}
				}
			}

			if methodConf == nil {
				log.Printf("skip: method %s does not have apigen:api mark",
					fDecl.Name.Name)
				continue
			}

			if fDecl.Recv == nil {
				log.Printf("skip: method %s is function!",
					fDecl.Name.Name)
				continue
			}

			methodConf.AuthKey = globalAuthKey
			methodConf.Name = fDecl.Name.Name
			methodConf.ReceiverName = fDecl.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
			//немного харкод(верю в то, что структура для валидации всегда 2я)
			//Достаём стрктуру обработчик
			//Метод может быть раньше объявления структуры обработчика
			if strct, ok := apiStructs[methodConf.ReceiverName]; !ok {
				newApiStruct := &ApiStruct{
					Name:    methodConf.ReceiverName,
					Methods: make(map[string]*MethodConfig),
				}
				newApiStruct.Methods[methodConf.Name] = methodConf
				apiStructs[newApiStruct.Name] = newApiStruct
			} else {
				strct.Methods[methodConf.Name] = methodConf
			}
			//Достаём стркутуру валидатор
			validatorName := fDecl.Type.Params.List[1].Type.(*ast.Ident).Name
			//Метод может быть раньше объявления структуры валидации
			if validator, ok := apiValidateStructs[validatorName]; !ok {
				newApiValidateStruct := &ApiValidateStruct{
					Name:   methodConf.ReceiverName,
					Fields: make([]*StructField, 0),
				}
				methodConf.ValidateStruct = newApiValidateStruct
				apiValidateStructs[newApiValidateStruct.Name] = newApiValidateStruct
			} else {
				methodConf.ValidateStruct = validator
			}
			continue
		}

		if sDecl, ok := decl.(*ast.GenDecl); !ok {
			log.Printf("skip: is not *ast.GenDecl\n")
		} else { //Объявление структуры
			for _, spec := range sDecl.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					log.Printf("skip: is not *ast.TypeSpec\n")
					continue
				}

				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					if !ok {
						log.Printf("skip: is not *ast.StructType\n")
						continue
					}
				}

				for _, field := range currStruct.Fields.List {
					if field.Tag != nil {
						tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
						var newApiValidateStruct *ApiValidateStruct
						if tags, ok := tag.Lookup("apivalidator"); ok { //Будет использоватся для валидации
							if newApiValidateStruct == nil { // первый заход на нужное поле
								//Могли уже создать, когда делали метод
								if strct, ok := apiValidateStructs[currType.Name.Name]; !ok {
									newApiValidateStruct = &ApiValidateStruct{
										Name:   currType.Name.Name,
										Fields: make([]*StructField, 0),
									}
									apiValidateStructs[newApiValidateStruct.Name] = newApiValidateStruct
								} else {
									newApiValidateStruct = strct
								}
							}
							//Парсим тег структуры в конфиг для валидатора
							newField := &StructField{
								CodeName: field.Names[0].Name,
								Type:     field.Type.(*ast.Ident).Name,
								Min:      -1,
								Max:      -1,
							}

							newApiValidateStruct.Fields = append(newApiValidateStruct.Fields, newField)
							args := strings.Split(tags, ",")
							wasParamName := false
							for _, arg := range args {
								if arg == "required" {
									newField.Required = true
								} else if strings.HasPrefix(arg, "paramname") {
									wasParamName = true
									newField.ParamName = arg[strings.Index(arg, "=")+1:]
								} else if strings.HasPrefix(arg, "enum") {
									newField.Enum = strings.Split(arg[strings.Index(arg, "=")+1:], "|")
								} else if strings.HasPrefix(arg, "default") {
									newField.DefaultValue = arg[strings.Index(arg, "=")+1:]
								} else if strings.HasPrefix(arg, "min") {
									val, err := strconv.Atoi(arg[strings.Index(arg, "=")+1:])
									if err != nil {
										log.Fatalf("in %s min convert error %v",
											newField.CodeName, err)
									}
									newField.Min = val
								} else if strings.HasPrefix(arg, "max") {
									val, err := strconv.Atoi(arg[strings.Index(arg, "=")+1:])
									if err != nil {
										log.Fatalf("in %s max convert error %v",
											newField.CodeName, err)
									}
									newField.Max = val
								}
							}
							if !wasParamName {
								newField.ParamName = strings.ToLower(newField.CodeName)
							}
						}
					}
				}
			}
		}
	}

	//Генерируем файл
	fmt.Fprintf(resultFile, "package %s\n", node.Name.Name)
	fmt.Fprintln(resultFile)
	fmt.Fprintln(resultFile, `import "net/http"`)
	fmt.Fprintln(resultFile, `import "strconv"`)
	fmt.Fprintln(resultFile, `import "encoding/json"`)
	fmt.Fprintln(resultFile)
	for _, strct := range apiStructs {
		log.Printf("STRUCT %s:\n", strct.Name)
		log.Println("METHODS")
		for name, method := range strct.Methods {
			log.Printf("\t%s -> %v\n", name, method)
			log.Printf("\tVALIDATE STRUCT %s:\n", method.ValidateStruct.Name)
			for _, field := range method.ValidateStruct.Fields {
				log.Printf("\t\t%s -> %v\n", field.CodeName, field)
			}
		}

		if _, ok := strct.Methods["ServeHTTP"]; ok {
			log.Fatalf("%s: ServeHTTP method already exsist", strct.Name)
		}
		//Строим ServeHTTP связку через шаблон
		serveHTTPTpl.Execute(resultFile, strct)

		//А handler будем собирать по кусочкам
		for _, method := range strct.Methods {
			fmt.Fprintf(resultFile, "func (h *%s) handle%s(w http.ResponseWriter, r *http.Request) {\n",
				method.ReceiverName, method.Name)
			
			if method.Method != "" {
				fmt.Fprintf(resultFile, `	if r.Method != "%s" {`+"\n",
					method.Method)
				fmt.Fprintln(resultFile,
					"\t\tw.WriteHeader(http.StatusNotAcceptable)")
				fmt.Fprintln(resultFile, `		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : "bad method"})`)
				fmt.Fprintf(resultFile, "\t\tw.Write(jsonRes)\n")
				fmt.Fprintf(resultFile, "\t\treturn\n\t}\n")
			}

			if method.Auth {
				fmt.Fprintf(resultFile, `	if r.Header.Get("X-Auth") != %s {`+"\n",
					method.AuthKey)
				fmt.Fprintln(resultFile,
					"\t\tw.WriteHeader(http.StatusForbidden)")
				fmt.Fprintln(resultFile, `		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : "unauthorized"})`)
				fmt.Fprintf(resultFile, "\t\tw.Write(jsonRes)\n")
				fmt.Fprintf(resultFile, "\t\treturn\n\t}\n")
			}
			fmt.Fprintln(resultFile, "\tr.ParseForm()")
			fmt.Fprintf(resultFile, "\tvalidateStuct := %s{}\n",
				method.ValidateStruct.Name)
			for _, field := range method.ValidateStruct.Fields {
				if field.Type == "int" {
					fmt.Fprintf(resultFile, `	val, err := strconv.Atoi(r.Form.Get("%s"))`+"\n", field.ParamName)
					fmt.Fprintf(resultFile, "\tif err != nil {\n")
					fmt.Fprintf(resultFile, "\t\tw.WriteHeader(http.StatusBadRequest)\n")
					fmt.Fprintf(resultFile, `		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : "%s must be int"})`+"\n",
						field.ParamName)
					fmt.Fprintf(resultFile, "\t\tw.Write(jsonRes)\n")
					fmt.Fprintf(resultFile, "\t\treturn\n\t}\n")
					fmt.Fprintf(resultFile, "\tvalidateStuct.%s = val\n", field.CodeName)
				} else if field.Type == "string" {
					fmt.Fprintf(resultFile, `	validateStuct.%s = r.Form.Get("%s")`+"\n",
						field.CodeName, field.ParamName)
				}
			}

			for _, field := range method.ValidateStruct.Fields {
				if field.Required {
					fmt.Fprintf(resultFile, `	if validateStuct.%s == "" {`+"\n",
						field.CodeName)
					fmt.Fprintf(resultFile, "\t\tw.WriteHeader(http.StatusBadRequest)\n")
					fmt.Fprintf(resultFile, `		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : "%s must me not empty"})`+"\n",
						field.ParamName)
					fmt.Fprintf(resultFile, "\t\tw.Write(jsonRes)\n")
					fmt.Fprintf(resultFile, "\t\treturn\n\t}\n")
				}

				if field.DefaultValue != "" {
					fmt.Fprintf(resultFile, `	if validateStuct.%s == "" {`+"\n",
						field.CodeName)
					fmt.Fprintf(resultFile, `		validateStuct.%s = "%s"`+"\n\t}\n",
						field.CodeName, field.DefaultValue)
				}

				if len(field.Enum) != 0 {
					fmt.Fprint(resultFile, "\tif !(")
					for indx, prot := range field.Enum {
						fmt.Fprintf(resultFile, `validateStuct.%s == "%s"`,
							field.CodeName, prot)
						if indx == len(field.Enum)-1 {
							fmt.Fprintf(resultFile, ") {\n")
						} else {
							fmt.Fprintf(resultFile, " || ")
						}
					}
					fmt.Fprintf(resultFile, "\t\tw.WriteHeader(http.StatusBadRequest)\n")
					fmt.Fprintf(resultFile, `		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : "%s must be one of [%s]"})`+"\n",
						field.ParamName, strings.Join(field.Enum, ", "))
					fmt.Fprintf(resultFile, "\t\tw.Write(jsonRes)\n")
					fmt.Fprintf(resultFile, "\t\treturn\n\t}\n")
				}

				if field.Min != -1 {
					fmt.Fprint(resultFile, "\tif ")
					if field.Type == "int" {
						fmt.Fprintf(resultFile, "validateStuct.%s < %d {\n",
							field.CodeName, field.Min)
					} else if field.Type == "string" {
						fmt.Fprintf(resultFile, "len(validateStuct.%s) < %d {\n",
							field.CodeName, field.Min)
					}
					fmt.Fprintf(resultFile, "\t\tw.WriteHeader(http.StatusBadRequest)\n")
					if field.Type == "int" {
						fmt.Fprintf(resultFile, `		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : "%s must be >= %d"})`+"\n",
							field.ParamName, field.Min)
					} else if field.Type == "string" {
						fmt.Fprintf(resultFile, `		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : "%s len must be >= %d"})`+"\n",
							field.ParamName, field.Min)
					}

					fmt.Fprintf(resultFile, "\t\tw.Write(jsonRes)\n")
					fmt.Fprintf(resultFile, "\t\treturn\n\t}\n")
				}

				if field.Max != -1 {
					fmt.Fprint(resultFile, "\tif ")
					if field.Type == "int" {
						fmt.Fprintf(resultFile, "validateStuct.%s > %d {\n",
							field.CodeName, field.Max)
					} else if field.Type == "string" {
						fmt.Fprintf(resultFile, "len(validateStuct.%s) > %d {\n",
							field.CodeName, field.Max)
					}
					fmt.Fprintf(resultFile, "\t\tw.WriteHeader(http.StatusBadRequest)\n")
					if field.Type == "int" {
						fmt.Fprintf(resultFile, `		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : "%s must be <= %d"})`+"\n",
							field.ParamName, field.Max)
					} else if field.Type == "string" {
						fmt.Fprintf(resultFile, `		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : "%s len must be <= %d"})`+"\n",
							field.ParamName, field.Max)
					}

					fmt.Fprintf(resultFile, "\t\tw.Write(jsonRes)\n")
					fmt.Fprintf(resultFile, "\t\treturn\n\t}\n")
				}
			}

			fmt.Fprintf(resultFile, "\tres, err := h.%s(r.Context(), validateStuct)\n",
				method.Name)
			fmt.Fprintln(resultFile, "\tif err != nil {")
			fmt.Fprintf(resultFile, "\t\tif apiError, ok := err.(ApiError); !ok {\n")
			fmt.Fprintf(resultFile, "\t\t\tw.WriteHeader(http.StatusInternalServerError)\n")
			fmt.Fprintf(resultFile, "\t\t} else {\n")
			fmt.Fprintf(resultFile, "\t\t\tw.WriteHeader(apiError.HTTPStatus)\n\t\t}\n")
			fmt.Fprintln(resultFile, `		jsonRes, _ := json.Marshal(map[string]interface{}{"error" : err.Error()})`)
			fmt.Fprintln(resultFile, "\t\tw.Write(jsonRes)\n\t\treturn\n\t}")
			fmt.Fprintln(resultFile, `	jsonRes, err := json.Marshal(map[string]interface{}{"error" : "", "response" : res})`)
			fmt.Fprintln(resultFile, "\tif err != nil {")
			fmt.Fprintln(resultFile, "\t\tw.WriteHeader(http.StatusInternalServerError)")
			fmt.Fprintln(resultFile, "\t\tw.Write([]byte{1})\n\t}")
			fmt.Fprintln(resultFile, "\tw.Write(jsonRes)")
			fmt.Fprintf(resultFile, "\n}\n\n")
		}
	}
}
