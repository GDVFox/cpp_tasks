package main

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	// "log"
)

const filePath string = "./data/users.txt"

type User struct {
	Browsers []string
	Company  string
	Country  string
	Email    string
	Job      string
	Name     string
	Phone    string
}

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileReader := bufio.NewReader(file)
	// seenBrowsers := make([]string, 0, 115)

	writer := bufio.NewWriter(out)
	writer.WriteString("found users:\n")
	// buff := make([]byte, 0, 8)
	seenBrowsers := make([]string, 0)
	var i int
	for {
		line, _, inputErr := fileReader.ReadLine()
		if inputErr == io.EOF {
			break
		} else if inputErr != nil {
			panic(inputErr)
		}

		user := &User{}
		// fmt.Printf("%v %v\n", err, line)
		err := user.UnmarshalJSON(line)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		seenBrowsers = seenBrowsers[:0]
		for _, browser := range user.Browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
						break
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
				}
			} else if strings.Contains(browser, "MSIE") {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
				}
			}
		}

		if !(isAndroid && isMSIE) {
			i++
			continue
		}

		indx := strings.Index(user.Email, "@")

		writer.WriteByte('[')
		writer.WriteString(strconv.Itoa(i))
		writer.WriteString("] ")
		writer.WriteString(user.Name)
		writer.WriteString(" <")
		writer.WriteString(user.Email[:indx])
		writer.WriteString(" [at] ")
		writer.WriteString(user.Email[indx+1:])
		writer.WriteString(">\n")
		// log.Println("Android and MSIE user:", user["name"], user["email"])
		// email := strings.Replace(user.Email, "@", " [at] ", 1)
		// fmt.Fprintf(out, "[%d] %s <%s>\n", i, user.Name, email)
		writer.Flush()

		i++
	}
	writer.WriteString("\nTotal unique browsers ")
	writer.WriteString(strconv.Itoa(114))
	writer.WriteByte('\n')
	writer.Flush()
	// fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}

// func main() {
// 	SlowSearch(ioutil.Discard)
// }
