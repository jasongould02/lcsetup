package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	//"io"
	"encoding/json"
	"net/http"

	"golang.org/x/net/html"
)

func scrapeContent(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		scrapeContent(c, buf)
	}
}

func scrapeProblemData(url string) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("Error Code: %d \t Status:%s\n", res.StatusCode, res.Status)
	}
	doc, e1 := html.Parse(res.Body)
	if e1 != nil {
		log.Fatalf("Error:%s\n", e1)
	}

	var bufferOut *bytes.Buffer
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			if n.Attr != nil && len(n.Attr) > 0 && n.Attr[0].Val == "__NEXT_DATA__" {
				text := &bytes.Buffer{}
				//collectText(n, text)
				scrapeContent(n, text)
				bufferOut = text
			}
		} else {
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var data map[string]interface{}
	if jsonError := json.Unmarshal(bufferOut.Bytes(), &data); jsonError != nil {
		fmt.Println("Error unmarshalling")
		fmt.Println(jsonError)
	}

	var findQuestionData func(map[string]interface{})
	findQuestionData = func(m map[string]interface{}) {
		for k, v := range m {
			switch vtyped := v.(type) {
			case string:
				//fmt.Println(k, "is string", vtyped)
			case bool:
				//fmt.Println(k, "is bool", vtyped)
			case float64:
				//fmt.Println(k, "is float64", vtyped)
			case []interface{}:
				fmt.Println(k, "is an array:")
				for _, u := range vtyped {
					val, exists := u.(map[string]interface{})
					if exists {
						findQuestionData(val)
					}
					//fmt.Printf("i:%d, u:%s\n", i, u)
				}
				//findQuestionData(v.(map[string]interface{}))
			case map[string]interface{}:
				if val, exists := vtyped["title"]; exists {
					fmt.Printf("Found titleSlug:%s\n", val)
					titleChan <- val.(string)
				} else if val, exists := vtyped["questionFrontendId"]; exists {
					fmt.Printf("Found questionFrontendID:%s\n", val)
					questionIdChan <- val.(string)
				}
				findQuestionData(v.(map[string]interface{}))
			default:
				fmt.Printf("unknown value type for key:%s\n", k)
			}
		}
	}

	findQuestionData(data)
}

var titleChan chan string = make(chan string)
var questionIdChan chan string = make(chan string)

func main() {
	args := os.Args[1:]
	fmt.Println(args)
	url := args[0]
	fileExtension := args[1]

	go scrapeProblemData(url)

	var title string
	var question string
	var questionId int
	for i := 0; i < 2; i++ {
		select {
		case title = <-titleChan:
			title = strings.ReplaceAll(title, " ", "_")
			fmt.Printf("Got Title:%s\n", title)
		case question = <-questionIdChan:
			fmt.Printf("Got QuestionId:%s\n", question)
			conv, err := strconv.Atoi(question)
			if err != nil {
				fmt.Println("Error converting questionId to an integer:", err)
			}
			questionId = conv
			fmt.Printf("Question ID converted:%d\n", conv)
		}
	}

	fmt.Printf("\n\nGot title:%s\nGot questionID:%d\n", title, questionId)
	folderTitle := question + "_" + title
	fmt.Printf("Folder title:%s\n", folderTitle)

	if exists(folderTitle + string(os.PathSeparator)) {
		fmt.Println("Folder ", folderTitle, " already exists")
		return
	}
	folderErr := os.Mkdir(folderTitle+string(os.PathSeparator), 0777)
	if folderErr != nil {
		log.Fatalf("Error creating folder:%s\n", folderErr)
	}

	if exists(folderTitle + string(os.PathSeparator) + folderTitle + "." + fileExtension) {
		fmt.Println("File ", folderTitle, " already exists")
		return
	}
	fileErr := os.WriteFile(folderTitle+string(os.PathSeparator)+folderTitle+"."+fileExtension, []byte(""), 0777)
	if fileErr != nil {
		log.Fatalf("Error creating file:%s\n", fileErr)
	}
}

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Path:%s does NOT exist\n", path)
		return false
	}
	return true
}
