package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	//"io"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type LCQuestionData struct {
	QuestionsFrontendId int    `json:"questionFrontendId"`
	QuestionsTitle      string `json:"titleSlug"`
}

//"questionFrontendId":"205"

func get(url string, client *http.Client) {
	start := time.Now()
	resp, httpError := client.Get(url)

	if httpError != nil {
		fmt.Printf("Error: %v\n", httpError)
		return
	}

	//Read now to capture time to read full response, not just headers
	//_, read_err := ioutil.ReadAll(resp.Body)
	bodyBytes, readErr := io.ReadAll(resp.Body)
	fmt.Printf("-Printing bodyBytes-\n\n%s\n-End of bodyBytes-\n\n", string(bodyBytes))
	elapsed := time.Since(start)

	doc, gqerr := goquery.NewDocumentFromReader(resp.Body)
	if gqerr != nil {
		log.Fatal(gqerr)
	}
	//<script id="__NEXT_DATA__"

	title := doc.Find("html.dark body script#__NEXT_DATA__").Text()
	fmt.Printf("\n----PRINTING----\n\n%s\n\n----ENDING PRINTING DATA----\n", title)

	data := []*LCQuestionData{}
	if jsonError := json.Unmarshal([]byte(title), &data); jsonError != nil {
		fmt.Println("Error unmarshalling")
		fmt.Println(jsonError)
		//panic(jsonError)
	}
	fmt.Printf("LCQuestionData:%+v\n", data[0])

	//questionData, jsonError := json.UnMarshal(resp.body))

	if resp != nil {
		defer resp.Body.Close()
	}

	if httpError != nil {
		fmt.Printf("Error: %v, Time taken: %v\n", readErr, elapsed)
		return
	}

	fmt.Printf("Status: %v, Time taken: %v\n", resp.Status, elapsed)
}

func customClient() *http.Client {
	//ref: Copy and modify defaults from https://golang.org/src/net/http/transport.go
	//Note: Clients and Transports should only be created once and reused
	transport := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			// Modify the time to wait for a connection to establish
			Timeout:   1 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := http.Client{
		Transport: &transport,
		Timeout:   4 * time.Second,
	}

	return &client
}

func scrapeProblemData(url string) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.Body)

	fmt.Println("----------------------------")
	temp, _ := httputil.DumpResponse(res, true)
	fmt.Println(string(temp))
	fmt.Println("----------------------------")

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("Error Code: %d \t Status:%s\n", res.StatusCode, res.Status)
	}
	/*fmt.Println("Header:")
	for k, v := range res.Header {
		fmt.Printf("Key:%s\n", k)
		for a, b := range v {
			fmt.Printf("a:%d\tb:%s\n", a, b)
		}
		fmt.Printf("End Node.\n\n")
	}

	fmt.Println("\nBody:")
	bodyBytes, err := io.ReadAll(res.Body)
	fmt.Printf("%s\nEND Body.\n", bodyBytes)*/

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Finding...")

	title := doc.Find("#qd-content > div.h-full.flex-col.ssg__qd-splitter-primary-w > div > div > div > div.flex.h-full.w-full.overflow-y-auto > div > div > div.w-full.px-5.pt-4 > div > div:nth-child(1) > div.flex-1 > div > div")
	fmt.Printf("Title Nodes %+v\n", title)
	for _, v := range title.Nodes {
		fmt.Printf("Nodes %+v\n", v)
	}
	//fmt.Printf("title:%s\n", title)
	/*
		title := doc.Find(".mr-2 text-lg font-medium text-label-1").Each(func(i int, s *goquery.Selection) {
			fmt.Printf("Searching child node %d...\n", i)
			temp := s.Find("s").Text()
			fmt.Printf("Found Text: %#v string:%s | length:%d\n", temp, string(temp), len(temp))
		})
		fmt.Printf("title is: %T\n", title)
	*/
	fmt.Println("Ending")
}

func main() {
	args := os.Args[1:]
	fmt.Println(args)
	url := args[0]
	fmt.Println("Fetching: " + url)
	//get(url, http.DefaultClient)
	get(url, customClient())
}

/*func main() {
	args := os.Args[1:]
	fmt.Println(args)

	scrapeProblemData(args[0])
}*/
