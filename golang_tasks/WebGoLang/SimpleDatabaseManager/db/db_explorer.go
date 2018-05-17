package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//MemesRouter роутер, такой же тупой, как и его название
type MemesRouter struct {
	simpleRoutes   map[string]func(http.ResponseWriter, *http.Request)
	advancedRoutes map[string]func(http.ResponseWriter, *http.Request,
		map[string]string)

	advancedRoutesTemplates []*AdvancedRoute
}

//AdvancedRoute вспомогательная стркутура для MemesRouter
type AdvancedRoute struct {
	URL    string
	Method string
	RegExp *regexp.Regexp
}

//NewMemesRouter создаёт MemesRouter
func NewMemesRouter() *MemesRouter {
	return &MemesRouter{
		simpleRoutes: make(map[string]func(http.ResponseWriter, *http.Request)),
		advancedRoutes: make(map[string]func(http.ResponseWriter, *http.Request,
			map[string]string), 0),

		advancedRoutesTemplates: make([]*AdvancedRoute, 0),
	}
}

//Вешает просто хэндлер без параметров
func (ms *MemesRouter) addSimpleHandler(url string, method string,
	handler func(http.ResponseWriter, *http.Request)) error {
	newURLName := url + "_" + method
	if _, ok := ms.simpleRoutes[newURLName]; ok {
		return fmt.Errorf("Route already exist")
	}
	ms.simpleRoutes[newURLName] = handler
	return nil
}

//Будет парсить параметры в {} в структуру
func (ms *MemesRouter) addAdvancedHandler(url string, method string,
	handler func(http.ResponseWriter, *http.Request, map[string]string)) error {
	braceIndex := strings.Index(url, "{")
	if braceIndex == -1 {
		return fmt.Errorf("Use addSimpleHandler() for simple routes")
	}

	newURLName := url + "_" + method
	if _, ok := ms.advancedRoutes[newURLName]; ok {
		return fmt.Errorf("Route already exist")
	}

	ms.advancedRoutes[newURLName] = handler
	newAdvancedRoute := &AdvancedRoute{
		URL:    newURLName,
		Method: method,
	}

	url = strings.Replace(url, "{", "(?P<", -1)
	url = strings.Replace(url, "}", ">[[:alnum:]_]+)", -1)
	newAdvancedRoute.RegExp = regexp.MustCompile("^" + url + "$")

	ms.advancedRoutesTemplates = append(ms.advancedRoutesTemplates, newAdvancedRoute)
	return nil
}

//Handle отправляет r на нужную функцию
func (ms *MemesRouter) Handle(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	if call, ok := ms.simpleRoutes[urlPath+"_"+r.Method]; ok {
		call(w, r)
		return
	}

	for _, tmpl := range ms.advancedRoutesTemplates {
		if tmpl.Method == r.Method && tmpl.RegExp.MatchString(urlPath) {
			log.Printf("MemesRouter: handle METHOD: %s;\tURL: %s\tto\t%s",
				r.Method, urlPath, tmpl.URL)
			params := make(map[string]string)
			match := tmpl.RegExp.FindStringSubmatch(urlPath)
			for i, name := range tmpl.RegExp.SubexpNames() {
				if i > 0 && i <= len(match) {
					params[name] = match[i]
				}
			}

			ms.advancedRoutes[tmpl.URL](w, r, params)
			return
		}
	}

	log.Printf("MemesRouter: handle METHOD: %s;\tURL: %s\t NOT FOUND",
		r.Method, urlPath)
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("<h1>404: Not Found!</h1>"))
}

//DBExplorer simple MySQL Database manager
type DBExplorer struct {
	DB     *sql.DB
	router *MemesRouter

	tablesInfo map[string][]*Column
}

//Column метаданные по некоторому столбцу
type Column struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default string
	Extra   string
}

//NewDbExplorer creates new DBExplorer
func NewDbExplorer(db *sql.DB) (*DBExplorer, error) {
	dbex := &DBExplorer{
		DB:         db,
		tablesInfo: make(map[string][]*Column),
		router:     NewMemesRouter(),
	}

	rows, err := dbex.DB.Query("SHOW TABLES")
	if err != nil {
		log.Fatalf("tables open error: %v", err)
	}

	defer rows.Close()
	var name string
	for rows.Next() {
		rows.Scan(&name)
		dbex.tablesInfo[name] = make([]*Column, 0)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("db error: %v:", err)
	}

	for tableName := range dbex.tablesInfo {
		rows, err := dbex.DB.Query(fmt.Sprintf("SHOW COLUMNS FROM %s", tableName))
		if err != nil {
			log.Fatalf("columns read form %s error: %v", tableName, err)
		}

		defer rows.Close()
		for rows.Next() {
			colInfo := &Column{}
			rows.Scan(&colInfo.Field, &colInfo.Type, &colInfo.Null,
				&colInfo.Key, &colInfo.Default, &colInfo.Extra)

			dbex.tablesInfo[tableName] = append(dbex.tablesInfo[tableName], colInfo)
		}
		if err := rows.Err(); err != nil {
			log.Fatalf("db error: %v:", err)
		}
	}

	dbex.router.addSimpleHandler("/", "GET", dbex.tableList)
	dbex.router.addAdvancedHandler("/{table}", "GET", dbex.getListFrom)
	dbex.router.addAdvancedHandler("/{table}/{id}", "GET", dbex.getRecord)
	dbex.router.addAdvancedHandler("/{table}/", "PUT", dbex.createRecord)
	dbex.router.addAdvancedHandler("/{table}/{id}", "POST", dbex.updateRecord)
	dbex.router.addAdvancedHandler("/{table}/{id}", "DELETE", dbex.deleteRecord)
	return dbex, nil
}

func (dbex *DBExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbex.router.Handle(w, r)
}

func (dbex *DBExplorer) readDBData(rows *sql.Rows, tableInfo []*Column) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	//Укладываем полученные записи в tableData
	size := len(cols)
	tableData := make([]map[string]interface{}, 0)
	records := make([]interface{}, size)
	recordsPtr := make([]interface{}, size)
	for rows.Next() {
		for i := 0; i < size; i++ {
			recordsPtr[i] = &records[i]
		}
		//Данные попадут в records через recordsPtr
		rows.Scan(recordsPtr...)
		entry := make(map[string]interface{})
		for i, column := range cols {
			var v interface{}
			rec := records[i]
			if b, ok := rec.([]byte); ok {
				tmp := string(b)
				if strings.HasPrefix(tableInfo[i].Type, "int") {
					v, err = strconv.Atoi(tmp)
					if err != nil {
						return nil, err
					}
				} else if strings.HasPrefix(tableInfo[i].Type, "float") {
					v, err = strconv.ParseFloat(tmp, 64)
					if err != nil {
						return nil, err
					}
				} else {
					v = tmp
				}
			} else {
				v = rec
			}

			entry[column] = v
		}
		tableData = append(tableData, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tableData, nil
}

func (dbex *DBExplorer) validateParametrs(field interface{}, info *Column) (interface{}, error) {
	switch x := field.(type) {
	case string:
		if info.Type != "varchar(255)" && info.Type != "text" {
			return nil, fmt.Errorf("field " + info.Field + " have invalid type")
		}
		return x, nil
	case float64:
		if math.Trunc(x) == x {
			if !strings.HasPrefix(info.Type, "int") {
				return nil, fmt.Errorf("field " + info.Field + " have invalid type")
			}
			return int(x), nil
		} else {
			if !strings.HasPrefix(info.Type, "float") {
				return nil, fmt.Errorf("field " + info.Field + " have invalid type")
			}
			return x, nil
		}
	default:
		if info.Null != "YES" {
			return nil, fmt.Errorf("field " + info.Field + " have invalid type")
		}
		return x, nil
	}
}

func (dbex *DBExplorer) tableList(w http.ResponseWriter,
	req *http.Request) {
	tables := make([]string, 0, len(dbex.tablesInfo))
	for tableName := range dbex.tablesInfo {
		tables = append(tables, tableName)
	}

	//Будем выдавать в лексикографическом порядке
	sort.Strings(tables)
	jsonRes, _ := json.Marshal(map[string]interface{}{
		"response": map[string]interface{}{
			"tables": tables}})
	w.Write(jsonRes)
}

func (dbex *DBExplorer) getListFrom(w http.ResponseWriter, req *http.Request,
	params map[string]string) {
	tableName := params["table"]
	tableInfo, ok := dbex.tablesInfo[tableName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		jsonRes, _ := json.Marshal(map[string]interface{}{
			"error": "unknown table"})
		w.Write(jsonRes)
		return
	}

	query := req.URL.Query()
	limit := "5"
	offset := "0"
	if lim := query.Get("limit"); lim != "" {
		if _, err := strconv.Atoi(lim); err == nil {
			limit = lim
		}
	}

	if off := query.Get("offset"); off != "" {
		if _, err := strconv.Atoi(off); err == nil {
			offset = off
		}
	}

	rows, err := dbex.DB.Query(fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?",
		tableName), limit, offset)
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}
	defer rows.Close()

	tableData, err := dbex.readDBData(rows, tableInfo)
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	jsonRes, _ := json.Marshal(map[string]interface{}{
		"response": map[string]interface{}{
			"records": tableData}})
	w.Write(jsonRes)
}

func (dbex *DBExplorer) getRecord(w http.ResponseWriter, req *http.Request,
	params map[string]string) {
	tableName := params["table"]
	tableInfo, ok := dbex.tablesInfo[tableName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		jsonRes, _ := json.Marshal(map[string]interface{}{
			"error": "unknown table"})
		w.Write(jsonRes)
		return
	}

	var priName string
	for _, info := range tableInfo {
		if info.Key == "PRI" {
			priName = info.Field
			break
		}
	}

	rows, err := dbex.DB.Query(fmt.Sprintf("SELECT * FROM %s WHERE %s= ?",
		tableName, priName), params["id"])
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}
	defer rows.Close()

	tableData, err := dbex.readDBData(rows, tableInfo)
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	if len(tableData) == 0 {
		w.WriteHeader(http.StatusNotFound)
		jsonRes, _ := json.Marshal(map[string]interface{}{
			"error": "record not found"})
		w.Write(jsonRes)
		return
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	jsonRes, _ := json.Marshal(map[string]interface{}{
		"response": map[string]interface{}{
			"record": tableData[0]}})
	w.Write(jsonRes)
}

func (dbex *DBExplorer) createRecord(w http.ResponseWriter, r *http.Request,
	params map[string]string) {
	tableName := params["table"]
	tableInfo, ok := dbex.tablesInfo[tableName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		jsonRes, _ := json.Marshal(map[string]interface{}{
			"error": "unknown table"})
		w.Write(jsonRes)
		return
	}

	bodyStrct := make(map[string]interface{})
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	err = json.Unmarshal(b, &bodyStrct)
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	values := make([]interface{}, 0)
	keys := make([]string, 0)
	var priName string
	for _, info := range tableInfo {
		newField, ok := bodyStrct[info.Field]
		if info.Key == "PRI" {
			priName = info.Field
			continue
		}

		//Валидация + подготовка INSERT запроса
		if !ok {
			if info.Null != "YES" {
				if info.Default == "" {
					w.WriteHeader(http.StatusBadRequest)
					jsonRes, _ := json.Marshal(map[string]interface{}{
						"error": "field " + info.Field + " is not nullable"})
					w.Write(jsonRes)
					return
				} else {
					newField = info.Default
				}
			}
			continue
		}

		v, err := dbex.validateParametrs(newField, info)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			jsonRes, _ := json.Marshal(map[string]interface{}{
				"error": err.Error()})
			w.Write(jsonRes)
			return
		}

		values = append(values, v)
		keys = append(keys, info.Field)
	}

	insertReq := bytes.Buffer{}
	insertReq.WriteString("INSERT INTO ")
	insertReq.WriteString(tableName)
	insertReq.WriteString(" (")
	for _, info := range keys[:len(keys)-1] {
		insertReq.WriteString("`" + info + "`, ")
	}
	insertReq.WriteString("`" + keys[len(keys)-1] + "`) ")
	insertReq.WriteString("VALUES (")
	for i := 0; i < len(keys)-1; i++ {
		insertReq.WriteString("?, ")
	}
	insertReq.WriteString("?)")

	result, err := dbex.DB.Exec(
		insertReq.String(),
		values...,
	)
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	jsonRes, _ := json.Marshal(map[string]interface{}{
		"response": map[string]interface{}{
			priName: lastID}})
	w.Write(jsonRes)
}

func (dbex *DBExplorer) updateRecord(w http.ResponseWriter, r *http.Request,
	params map[string]string) {
	tableName := params["table"]
	tableInfo, ok := dbex.tablesInfo[tableName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		jsonRes, _ := json.Marshal(map[string]interface{}{
			"error": "unknown table"})
		w.Write(jsonRes)
		return
	}

	bodyStrct := make(map[string]interface{})
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	err = json.Unmarshal(b, &bodyStrct)
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	values := make([]interface{}, 0)
	keys := make([]string, 0)
	var priName string
	for _, info := range tableInfo {
		if info.Key == "PRI" {
			priName = info.Field
		}

		newField, ok := bodyStrct[info.Field]
		//Валидация + подготовка UPDATE запроса
		if !ok {
			continue
		}

		// primary key нельзя обновлять у существующей записи
		if info.Key == "PRI" {
			w.WriteHeader(http.StatusBadRequest)
			jsonRes, _ := json.Marshal(map[string]interface{}{
				"error": "field " + info.Field + " have invalid type"})
			w.Write(jsonRes)
			return
		}

		v, err := dbex.validateParametrs(newField, info)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			jsonRes, _ := json.Marshal(map[string]interface{}{
				"error": err.Error()})
			w.Write(jsonRes)
			return
		}

		values = append(values, v)
		keys = append(keys, info.Field)
	}

	insertReq := bytes.Buffer{}
	insertReq.WriteString("UPDATE ")
	insertReq.WriteString(tableName)
	insertReq.WriteString(" SET ")
	for _, info := range keys[:len(keys)-1] {
		insertReq.WriteString("`" + info + "` = ?,")
	}
	insertReq.WriteString("`" + keys[len(keys)-1] + "` = ? ")
	insertReq.WriteString("WHERE ")
	insertReq.WriteString(priName)
	insertReq.WriteString(" = ?")
	values = append(values, params["id"])

	result, err := dbex.DB.Exec(insertReq.String(), values...)
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	updated, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	jsonRes, _ := json.Marshal(map[string]interface{}{
		"response": map[string]interface{}{
			"updated": updated}})
	w.Write(jsonRes)

}

func (dbex *DBExplorer) deleteRecord(w http.ResponseWriter, r *http.Request,
	params map[string]string) {

	tableName := params["table"]
	_, ok := dbex.tablesInfo[tableName]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		jsonRes, _ := json.Marshal(map[string]interface{}{
			"error": "unknown table"})
		w.Write(jsonRes)
		return
	}

	result, err := dbex.DB.Exec(fmt.Sprintf(
		"DELETE FROM %s WHERE id = ?", tableName), params["id"])
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "500", http.StatusInternalServerError)
	}

	jsonRes, _ := json.Marshal(map[string]interface{}{
		"response": map[string]interface{}{
			"deleted": deleted}})
	w.Write(jsonRes)
}
