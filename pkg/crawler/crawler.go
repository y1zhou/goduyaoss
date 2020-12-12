package crawler

import (
	"log"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

// Provider holds the information about a provider. It's possible for a provider
// to have multiple subgroups with different names and speed test results (ImgURL).
type Provider struct {
	Name     string
	TitleID  string
	ImgURL   string
	Subgroup []Provider
}

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
func fetchProviders(doc *goquery.Document) []Provider {
	var providers []Provider
	regexProvider := regexp.MustCompile(`^\d*\.`)
	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		// Each provider's name starts with a serial number.
		title := s.Text()
		if regexProvider.MatchString(title) {
			providerID, _ := s.Attr("id")
			res := Provider{Name: title, TitleID: providerID}

			// See if there's subgroups. Check for <h3> elements until the next provider
			s.NextFilteredUntil("h3", "h2").Each(func(i int, ss *goquery.Selection) {
				subTitle := ss.Text()
				subID, _ := ss.Attr("id")
				res.Subgroup = append(res.Subgroup, Provider{Name: subTitle, TitleID: subID})
			})

			providers = append(providers, res)
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
