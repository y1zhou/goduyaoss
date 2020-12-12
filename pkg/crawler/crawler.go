package crawler

import (
	"log"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

// Given a URL, return the response as a goquery document.
func requestPage(url string) *goquery.Document {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

// Find the providers in the <h2> elements, and return the id
// attr of those elements as a list.
func fetchProviders(doc *goquery.Document) []string {
	var providers []string
	regexProvider := regexp.MustCompile(`^\d*\.`)
	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		title := s.Text()
		if regexProvider.MatchString(title) {
			ProviderID, _ := s.Attr("id")
			log.Printf("Provider found: %q -> %q\n", title, ProviderID)
			providers = append(providers, ProviderID)
		}
	})
	return providers
}

func main() {
	doc := requestPage("https://www.duyaoss.com/archives/1031/")
	providers := fetchProviders(doc)
	for _, provider := range providers {
		log.Printf("Provider ID: %q\n", provider)
	}
	// FetchProviders("https://www.duyaoss.com/archives/3/")
	// FetchProviders("https://www.duyaoss.com/archives/1/")
}

// func extractFigs(e *goquery.Selection) {
// 	e.Find("figure").Siblings().NextUntil("h2").Each(func(i int, s *goquery.Selection) {
// 		link, _ := s.Attr("data-src")
// 		fmt.Printf("  %d, Image link: %s\n", i, link)
// 	}
// }
