package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"

	"encoding/json"
	"net/http"

	"golang.org/x/net/html"
)

const permissionBits = fs.ModePerm

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		//fmt.Printf("Path:%s does NOT exist\n", path)
		return false
	}
	return true
}

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
		log.Fatalf("Error Code:%d\tStatus:%s\n", res.StatusCode, res.Status)
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
				scrapeContent(n, text)
				bufferOut = text
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var data map[string]interface{}
	if jsonError := json.Unmarshal(bufferOut.Bytes(), &data); jsonError != nil {
		fmt.Println("Error unmarshalling JSON data")
		fmt.Println(jsonError)
		return
	}

	var findQuestionData func(map[string]interface{})
	findQuestionData = func(m map[string]interface{}) {
		for _, v := range m {
			switch vtyped := v.(type) {
			case string:
				//fmt.Println(k, "is string", vtyped)
			case bool:
				//fmt.Println(k, "is bool", vtyped)
			case float64:
				//fmt.Println(k, "is float64", vtyped)
			case []interface{}: // JSON Array
				for _, u := range vtyped {
					val, exists := u.(map[string]interface{})
					if exists {
						findQuestionData(val)
					}
					//fmt.Printf("i:%d, u:%s\n", i, u)
				}
			case map[string]interface{}: // JSON Object
				if val, exists := vtyped["title"]; exists {
					titleChan <- val.(string)
				} else if val, exists := vtyped["questionFrontendId"]; exists {
					questionIdChan <- val.(string)
				}
				findQuestionData(vtyped)
			default:
				//fmt.Printf("unknown value type for key:%s\n", k)
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
	for i := 0; i < 2; i++ {
		select {
		case title = <-titleChan:
			title = strings.ReplaceAll(title, " ", "_")
		case question = <-questionIdChan:
		}
	}

	fmt.Printf("Got Title:%s\tGot QuestionId:%s\n", title, question)
	folderTitle := question + "_" + title
	createFolder(folderTitle)

	filePath := folderTitle + string(os.PathSeparator) + folderTitle + "." + fileExtension
	createFile(filePath, fileExtension)

	fmt.Printf("\n\tvim %s\n\n", filePath)
	cmd := exec.Command("vim", filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if vimErr := cmd.Run(); vimErr != nil {
		fmt.Printf("Error opening editor:%s\n", vimErr)
	}
	os.Exit(0)
}

func createFolder(folderTitle string) {
	if exists(folderTitle + string(os.PathSeparator)) {
		fmt.Println("Folder ", folderTitle, " already exists")
		return
	}
	folderErr := os.Mkdir(folderTitle+string(os.PathSeparator), permissionBits)
	if folderErr != nil {
		log.Fatalf("Error creating folder:%s\n", folderErr)
	}
}

func createFile(filePath string, fileExtension string) {
	if exists(filePath) {
		fmt.Println("File ", filePath, " already exists")
		return
	}
	fileErr := os.WriteFile(filePath, []byte(""), permissionBits)
	if fileErr != nil {
		log.Fatalf("Error creating file:%s\n", fileErr)
	}
}
