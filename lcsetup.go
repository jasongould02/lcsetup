package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	//"io"
	"encoding/json"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
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

	/*fmt.Println(res.Body)

	fmt.Println("----------------------------")
	temp, _ := httputil.DumpResponse(res, true)
	fmt.Println(string(temp))
	fmt.Println("----------------------------")
	*/

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("Error Code: %d \t Status:%s\n", res.StatusCode, res.Status)
	}
	/*fmt.Println("Starting to Wait")
	time.Sleep(2 * time.Second)
	fmt.Println("Finished Waiting")*/
	doc, e1 := html.Parse(res.Body)
	if e1 != nil {
		log.Fatalf("Error:%s\n", e1)
	}
	var collectText func(*html.Node, *bytes.Buffer)
	collectText = func(n *html.Node, buf *bytes.Buffer) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collectText(c, buf)
		}
	}

	//var jsonString string
	var bufferOut *bytes.Buffer
	var f func(*html.Node)
	//__NEXT_DATA__
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			if n.Attr != nil && len(n.Attr) > 0 && n.Attr[0].Val == "__NEXT_DATA__" {
				fmt.Println("FOUND NEXT DATA")
				text := &bytes.Buffer{}
				collectText(n, text)
				fmt.Println(text.String())
				//jsonString = text.String()
				bufferOut = text
			}
			//fmt.Printf("NameSpace:%s\tData:%s\tAttr:%+v\n", n.Namespace, n.Data, n.Attr)
		} else {
			//fmt.Printf("Type:%d\tNameSpace:%s\tData:%s\tAttr:%+v\n", n.Type, n.Namespace, n.Data, n.Attr)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	var data map[string]*json.RawMessage
	fmt.Printf("bufferOut:%s\n", bufferOut)
	if jsonError := json.Unmarshal(bufferOut.Bytes(), &data); jsonError != nil {
		fmt.Println("Error unmarshalling")
		fmt.Println(jsonError)
		//panic(jsonError)
	}

	// type question struct {
	// 	PageProps string `json:"pageProps"`
	// 	dehydratedState struct {
	// 		queries struct
	// 	} `json:"dehydratedState"`
	// }

	var findQuestionData func(map[string]*json.RawMessage)
	findQuestionData = func(m map[string]*json.RawMessage) {
		for k, v := range m {
			arr := []byte(*v)
			asString := string(arr[:])
			if strings.Contains(asString, "titleSlug") {
				fmt.Printf("---\tFOUNDTITLESLUG:%s\t---\n", asString)
			}
			var next map[string]*json.RawMessage // maybe have struct of question data i want and also a map[string]*json.RawMessage of data i dont understand,
			// then scan the unknown data recursively until i find the question data i want
			if nextErr := json.Unmarshal([]byte(*v), &next); nextErr != nil {
				// Errors start when it is normal variables be parsed and not a json object/array
				fmt.Printf("Error in findingQuestionData: %s\n", nextErr)
			}
			val, exists := next["question"]
			if exists {
				fmt.Printf("\n\n---\n%s\n---\n\n", val)
			}
			if k == "question" {
				fmt.Printf("Found and printing:%s\n", v)
			}
			findQuestionData(next)
		}
	}

	findQuestionData(data)

	/*for k, v := range data {
		fmt.Printf("\tk:%s\tvtype:%T v:%s\n", k, v, v)
		var extra map[string]*json.RawMessage
		json.Unmarshal([]byte(*v), &extra)
		for x, y := range extra {
			fmt.Printf("\t\tlolololololololk:%s\tvtype:%T v:%s\n", x, y, y)
		}
	}*/
	//fmt.Printf("LCQuestionData:%+v\n", data[0])
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

	/*doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Finding...")

	title := doc.Find("#qd-content > div.h-full.flex-col.ssg__qd-splitter-primary-w > div > div > div > div.flex.h-full.w-full.overflow-y-auto > div > div > div.w-full.px-5.pt-4 > div > div:nth-child(1) > div.flex-1 > div > div")
	fmt.Printf("Title Nodes %+v\n", title)
	for _, v := range title.Nodes {
		fmt.Printf("Nodes %+v\n", v)
	}*/
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

/*func main() {
	args := os.Args[1:]
	fmt.Println(args)
	url := args[0]
	fmt.Println("Fetching: " + url)
	//get(url, http.DefaultClient)
	get(url, customClient())
}*/

func main() {
	args := os.Args[1:]
	fmt.Println(args)

	scrapeProblemData(args[0])
}
