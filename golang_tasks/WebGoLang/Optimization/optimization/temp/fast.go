// package main

// // suppress unused package warning
// import (
// 	"bufio"
// 	"bytes"
// 	"io"
// 	"io/ioutil"
// 	"os"
// 	"strconv"
// 	"strings"

// 	"github.com/mailru/easyjson"
// 	// "log"
// )

// type User struct {
// 	Browsers []string
// 	Company  string
// 	Country  string
// 	Email    string
// 	Job      string
// 	Name     string
// 	Phone    string
// }

// const filePath string = "./data/users.txt"

// func FastSearch(out io.Writer) {
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		panic(err)
// 	}

// 	seenBrowsers := make([]string, 0, 114)
// 	fileReader := bufio.NewReader(file)
// 	buf := bytes.Buffer{}
// 	users := make([]*User, 0, 1000)
// 	for {
// 		line, _, inputErr := fileReader.ReadLine()

// 		// log.Println(string(line))
// 		if inputErr == io.EOF {
// 			break
// 		} else if inputErr != nil {
// 			panic(inputErr)
// 		}

// 		user := &User{}
// 		err := easyjson.Unmarshal(line, user)
// 		// fmt.Printf("%v %v\n", err, line)
// 		if err != nil {
// 			panic(err)
// 		}

// 		users = append(users, user)
// 	}

// 	buf.WriteString("found users:\n")
// 	var isAndroid, isMSIE, notSeenBefore bool
// 	for i, user := range users {
// 		isAndroid = false
// 		isMSIE = false

// 		for _, browser := range user.Browsers {
// 			if strings.Contains(browser, "Android") {
// 				isAndroid = true
// 				notSeenBefore = true
// 				for _, item := range seenBrowsers {
// 					if item == browser {
// 						notSeenBefore = false
// 						break
// 					}
// 				}
// 				if notSeenBefore {
// 					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
// 					seenBrowsers = append(seenBrowsers, browser)
// 				}
// 			} else if strings.Contains(browser, "MSIE") {
// 				isMSIE = true
// 				notSeenBefore = true
// 				for _, item := range seenBrowsers {
// 					if item == browser {
// 						notSeenBefore = false
// 						break
// 					}
// 				}
// 				if notSeenBefore {
// 					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
// 					seenBrowsers = append(seenBrowsers, browser)
// 				}
// 			}
// 		}

// 		if !(isAndroid && isMSIE) {
// 			continue
// 		}

// 		// log.Println("Android and MSIE user:", user["name"], user["email"])
// 		buf.WriteByte('[')
// 		buf.WriteString(strconv.Itoa(i))
// 		buf.WriteByte(']')
// 		buf.WriteByte(' ')
// 		buf.WriteString(user.Name)
// 		buf.WriteByte(' ')
// 		buf.WriteByte('<')
// 		buf.WriteString(strings.Replace(user.Email, "@", " [at] ", 1))
// 		buf.WriteByte('>')
// 		buf.WriteByte('\n')
// 		// fmt.Fprintf(out, "[%d] %s <%s>\n", i, user.Name, email)
// 	}

// 	buf.WriteString("\nTotal unique browsers ")
// 	buf.WriteString(strconv.Itoa(len(seenBrowsers)))
// 	buf.WriteByte('\n')
// 	buf.WriteTo(out)
// 	// fmt.Fprintln(out, "\nTotal unique browsers", uniqueBrowsers)
// }

// func main() {
// 	FastSearch(ioutil.Discard)
// }
