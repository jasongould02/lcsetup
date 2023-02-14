package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"

	//"time"

	"encoding/json"
	"net/http"
)

const permissionBits = fs.ModePerm

var titleChan chan string = make(chan string)
var questionIdChan chan string = make(chan string)

func queryQuestion(titleSlug string) {
	//https://leetcode.com/graphql?query={question(titleSlug:%22isomorphic-strings%22)%20{questionFrontendId%20titleSlug}}
	res, err := http.Get("https://leetcode.com/graphql?query={question(titleSlug: " + "\"" + titleSlug + "\") {questionFrontendId title}}")
	if err != nil {
		log.Fatalf("Error with get request:%s\n", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("Error Code:%d\tStatus:%s\n", res.StatusCode, res.Status)
	}
	fmt.Printf("Status Code:%d\tStatus:%s\n", res.StatusCode, res.Status)

	bodyBytes, bodyBytesErr := io.ReadAll(res.Body)
	if bodyBytesErr != nil {
		log.Fatalf("Error reading body bytes\n")
	}

	//fmt.Println("Starting to unmarshal JSON")
	var data map[string]interface{}
	if jsonError := json.Unmarshal(bodyBytes, &data); jsonError != nil {
		log.Fatalf("Error unmarshalling JSON data:%s\n", jsonError)
		return
	}
	//fmt.Println("Finishing unmarshalling data")

	//fmt.Printf("JSONMAP:%+v\n", data)
	var findQuestionData func(map[string]interface{})
	findQuestionData = func(m map[string]interface{}) {
		for k, v := range m {
			switch vtyped := v.(type) {
			case string:
				switch k {
				case "questionFrontendId":
					questionIdChan <- vtyped
				case "title":
					titleChan <- vtyped
				}
			case []interface{}: // JSON Array
				for _, iv := range vtyped {
					switch element := iv.(type) {
					case map[string]interface{}:
						findQuestionData(element)
					}
				}
			case map[string]interface{}: // JSON Object
				findQuestionData(vtyped)
			default:
				fmt.Printf("Unknown value type for key:%s\n", k)
			}
		}
	}

	findQuestionData(data)
	//fmt.Println("BodyBytes:", string(bodyBytes))
}

func main() {
	args := os.Args[1:]
	fmt.Println(args)
	url := args[0]
	fileExtension := args[1]

	go queryQuestion(url)

	//go scrapeProblemData(url)

	var title string
	var question string
	for i := 0; i < 2; i++ {
		select {
		case title = <-titleChan:
			title = strings.ReplaceAll(title, " ", "_")
		case question = <-questionIdChan:
		}
	}

	//fmt.Printf("Title:%s\tQuestionId:%s\n", title, question)
	folderTitle := question + "_" + title
	createFolder(folderTitle)

	filePath := folderTitle + string(os.PathSeparator) + folderTitle + "." + fileExtension
	createFile(filePath, fileExtension) // Will not overwrite existing files

	fmt.Printf("Created %s\n", filePath)
	fmt.Printf("\n\tvim %s\n\n", filePath)
	cmd := exec.Command("vim", filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if vimErr := cmd.Run(); vimErr != nil {
		log.Printf("Error opening editor:%s\n", vimErr)
	}

}

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.Fatalf("Error:%s\n", err)
	}
	return true
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
