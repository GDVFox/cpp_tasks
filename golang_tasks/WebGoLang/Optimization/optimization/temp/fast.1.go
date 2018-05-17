package main

import (
	"bufio"
	"fmt"
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

	seenBrowsers := []string{}
	uniqueBrowsers := 0

	fmt.Fprintln(out, "found users:")
	for i := 0; ; i++ {
		line, _, inputErr := fileReader.ReadLine()
		if inputErr == io.EOF {
			break
		} else if inputErr != nil {
			panic(inputErr)
		}

		user := &User{}
		// fmt.Printf("%v %v\n", err, line)
		err := user.UnmarshalJSON([]byte(line))
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			if strings.Contains(browser, "Android") {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
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
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			i++
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		indx := strings.Index(user.Email, "@")
		fmt.Fprintln(out, "[", strconv.Itoa(i), "] ",
			user.Name, " <", user.Email[:indx], " [at] ", user.Email[indx+1:], ">")
		i++
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}

// func main() {
// 	SlowSearch(ioutil.Discard)
// }
