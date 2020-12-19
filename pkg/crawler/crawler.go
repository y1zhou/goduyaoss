package crawler

import (
	"image"
	"image/png"
	"log"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

// Pages - the pages to crawl.
var Pages = map[string]string{
	"ChinaMobile":  "https://www.duyaoss.com/archives/1031/",
	"ChinaUnicom":  "https://www.duyaoss.com/archives/3/",
	"ChinaTelecom": "https://www.duyaoss.com/archives/1/",
}

// Provider holds the information about a provider. It's possible for a provider
// to have multiple subgroups with different names and speed test results (ImgURL).
type Provider struct {
	Name     string
	ImgURL   string
	Subgroup []Provider
}

// RequestPage - Given a URL, return the response as a goquery document.
func RequestPage(url string) *goquery.Document {
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

// FetchProviders - Find the providers in the <h2> elements,
// and parse the result into an array of `Provider` structs.
func FetchProviders(doc *goquery.Document) []Provider {
	var providers []Provider
	regexProvider := regexp.MustCompile(`^\d*\.`)
	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		// Each provider's name starts with a serial number.
		title := s.Text()
		if regexProvider.MatchString(title) {
			res := Provider{Name: title}

			// See if there's subgroups. Check for <h3> elements until the next provider
			s.NextFilteredUntil("h3", "h2").
				Each(func(i int, ss *goquery.Selection) {
					subTitle := ss.Text()

					link, found := ss.NextFilteredUntil("figure", "h3").First().
						Find("img").Attr("data-src")
					if found {
						subProvider := Provider{Name: subTitle, ImgURL: link}
						res.Subgroup = append(res.Subgroup, subProvider)
					}

				})

			// If there's no subgroups, find the image(s) for the provider
			if res.Subgroup == nil {
				link, found := s.NextFilteredUntil("figure", "h2").First().
					Find("img").Attr("data-src")
				if found {
					res.ImgURL = link
				}
			}

			providers = append(providers, res)
		}
	})
	return providers
}

// FetchImage - Given the URL to an image, return an `image.Image` interface.
func FetchImage(url string) image.Image {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	img, err := png.Decode(res.Body)
	if err != nil {
		log.Fatalf("Error loading png image: %s", err)
	}
	return img
}
