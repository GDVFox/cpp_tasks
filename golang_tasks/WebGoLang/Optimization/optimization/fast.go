package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	// "log"
)

type User struct {
	Browsers []string
	Company  string
	Country  string
	Email    string
	Job      string
	Name     string
	Phone    string
}

const filePath string = "./data/users.txt"

func FastSearch(out io.Writer) {
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		panic(fileErr)
	}
	defer file.Close()

	writer := bytes.Buffer{}
	fileScanner := bufio.NewScanner(file)
	seenBrowsers := make(map[string]struct{}, 0)
	intBuff := make([]byte, 0, 8)

	writer.WriteString("found users:\n")
	var i int64
	user := &User{}
	for fileScanner.Scan() {
		err := user.UnmarshalJSON(fileScanner.Bytes())
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				seenBrowsers[browser] = struct{}{}
			} else if strings.Contains(browser, "MSIE") {
				isMSIE = true
				seenBrowsers[browser] = struct{}{}
			}
		}

		if !(isAndroid && isMSIE) {
			i++
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		indx := strings.Index(user.Email, "@")
		writer.WriteByte('[')
		intBuff = intBuff[:0]
		writer.Write(strconv.AppendInt(intBuff, i, 10))
		writer.WriteString("] ")
		writer.WriteString(user.Name)
		writer.WriteString(" <")
		writer.WriteString(user.Email[:indx])
		writer.WriteString(" [at] ")
		writer.WriteString(user.Email[indx+1:])
		writer.WriteString(">\n")
		writer.WriteTo(out)
		i++
	}

	writer.WriteString("\nTotal unique browsers ")
	intBuff = intBuff[:0]
	writer.Write(strconv.AppendInt(intBuff, int64(len(seenBrowsers)), 10))
	writer.WriteByte('\n')
	writer.WriteTo(out)
}

func main() {
	FastSearch(os.Stdout)
}
