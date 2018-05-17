package main

import (
	"bufio"
	json "encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
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

const filePath string = "../data/users.txt"

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fileReader := bufio.NewReader(file)
	users := make([]User, 0)
	for i := 0; i < 10; i++ {
		line, _, inputErr := fileReader.ReadLine()

		// log.Println(string(line))
		if inputErr == io.EOF {
			break
		} else if inputErr != nil {
			panic(inputErr)
		}

		user := &User{}
		// fmt.Printf("%v %v\n", err, line)
		err := json.Unmarshal(line, user)
		log.Printf("%v", user)
		if err != nil {
			panic(err)
		}
		users = append(users, *user)
	}
}

func main() {
	FastSearch(ioutil.Discard)
}


for _, browser := range user.Browsers {
	if strings.Contains(browser, "Android") {
		isAndroid = true
		if seenBefore := seenBrowsers[browser]; !seenBefore {
			// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
			seenBrowsers[browser] = true
			uniqueBrowsers++
		}
	} else if strings.Contains(browser, "MSIE") {
		isMSIE = true
		if seenBefore := seenBrowsers[browser]; !seenBefore {
			// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
			seenBrowsers[browser] = true
			uniqueBrowsers++
		}
	}
}