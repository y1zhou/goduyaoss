package main

import (
	"log"
	"runtime"
	"sync"

	"github.com/y1zhou/goduyaoss/pkg/crawler"
	"github.com/y1zhou/goduyaoss/pkg/ocr"
)

func main() {
	dbName := "test.db"

	queue := make(chan ocr.Job, 5)

	// Send jobs to the queue
	var wgCrawler sync.WaitGroup
	wgCrawler.Add(1)
	go func() {
		defer wgCrawler.Done()
		for netProvider, url := range crawler.Pages {
			doc := crawler.RequestPage(url)
			providers := crawler.FetchProviders(doc)

			for _, provider := range providers {
				if provider.ImgURL != "" {
					img := crawler.FetchImage(provider.ImgURL)
					ocr.AddJob(queue, img, netProvider, provider.Name)

					log.Printf("[main] %s -> %s added to queue\n",
						netProvider, provider.Name)
				} else {
					for _, subProvider := range provider.Subgroup {
						img := crawler.FetchImage(subProvider.ImgURL)
						ocr.AddJob(queue, img, netProvider, subProvider.Name)

						log.Printf("[main] %s -> %s added to queue\n",
							netProvider, subProvider.Name)
					}
				}
			}
		}
		close(queue)
	}()

	// Each Tesseract process uses a maximum of 4 threads
	// https://github.com/tesseract-ocr/tesseract/issues/1600
	numWorkers := runtime.NumCPU() / 4
	log.Printf("Spawning %d workers\n", numWorkers)

	var wgWorker sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wgWorker.Add(1)
		go ocr.Worker(w+1, dbName, queue, &wgWorker)
	}

	wgCrawler.Wait()
	log.Printf("Crawler finished!")
	wgWorker.Wait()
	log.Printf("All jobs finished!")
}
