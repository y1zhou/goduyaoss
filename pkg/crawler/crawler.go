package crawler

import (
	"fmt"

	"github.com/gocolly/colly"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector()

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// On every a element which has href attribute call callback
	c.OnHTML("h2", func(e *colly.HTMLElement) {
		link := e.Attr("id")
		// Print link
		fmt.Printf("Provider found: %q -> %s\n", e.Text, link)

	})

	// Start scraping on https://hackerspaces.org
	c.Visit("https://www.duyaoss.com/archives/1031/")
	// c.Visit("https://www.duyaoss.com/archives/3/")
	// c.Visit("https://www.duyaoss.com/archives/1/")
}
