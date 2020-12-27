package main

import (
	"log"
	"runtime"
	"sync"

	"github.com/y1zhou/goduyaoss/pkg/crawler"
	"github.com/y1zhou/goduyaoss/pkg/ocr"
)

func main() {
	netProvider := "ChinaMobile"
	url := "https://www.duyaoss.com/archives/1031/"
	log.Printf("Crawling %s\n", netProvider)
	doc := crawler.RequestPage(url)
	providers := crawler.FetchProviders(doc)
	provTest := providers[5:10]

	var wg sync.WaitGroup
	// Each Tesseract process uses a maximum of 4 threads
	// https://github.com/tesseract-ocr/tesseract/issues/1600
	numWorkers := runtime.NumCPU() / 4
	queue := make(chan ocr.Job, numWorkers)
	res := make(chan [][]string, len(providers)*3)

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go ocr.Worker(queue, res, &wg)
	}

	// Send jobs to the queue
	go func() {
		defer close(queue)
		for _, provider := range provTest {
			if provider.ImgURL != "" {
				img := crawler.FetchImage(provider.ImgURL)
				ocr.AddJob(queue, img, netProvider, provider.Name)
			} else {
				for _, subProvider := range provider.Subgroup {
					img := crawler.FetchImage(subProvider.ImgURL)
					ocr.AddJob(queue, img, netProvider, subProvider.Name)
				}
			}
		}
	}()

	// Close the res channel after all workers finish
	wg.Wait()
	close(res)

	for s := range res {
		ocr.PrintTable(s)
	}

	// for netProvider, url := range crawler.Pages {
	// 	log.Printf("Crawling %s\n", netProvider)
	// 	doc := crawler.RequestPage(url)
	// 	providers := crawler.FetchProviders(doc)

	// 	for _, provider := range providers {
	// 		log.Printf("Getting information for: %q\n", provider.Name)
	// 		if provider.ImgURL != "" {
	// 			img := crawler.FetchImage(provider.ImgURL)
	// 			imgMat := ocr.ImgToMat(img)

	// 			client := gosseract.NewClient()
	// 			defer client.Close()

	// 			testVersion, testTime := ocr.GetMetadata(imgMat, client)
	// 			log.Printf("%s\n%s\n", testVersion, testTime)

	// 			providerInfo := ocr.ImgToTable(imgMat)
	// 			log.Printf("%s has %d rows and %d columns\n", provider.Name, len(providerInfo), len(providerInfo[0]))
	// 		}
	// 	}
	// }
}
