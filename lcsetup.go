import (
	"fmt"
	"log"
	"os"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func scrapeProblemData(url string) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("Error Code: %d \t Status:%s\n", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Finding...")
	title := doc.Find(".mr-2 text-lg font-medium text-label-1").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("Searching child node %d...\n", i)
		temp := s.Find("s").Text()
		fmt.Printf("Found Text: %#v string:%s | length:%d\n", temp, string(temp), len(temp))
	})
	fmt.Printf("title is: %T\n", title)
	fmt.Println("Ending")
}

func main() {
	args := os.Args[1:]
	fmt.Println(args)

	scrapeProblemData(args[0])
}
