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

	queue := make(chan ocr.Job)

	// Send jobs to the queue
	go func() {
		defer close(queue)
		for netProvider, url := range crawler.Pages {
			doc := crawler.RequestPage(url)
			providers := crawler.FetchProviders(doc)
			provTest := providers[4:6]

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
		}
	}()

	// Each Tesseract process uses a maximum of 4 threads
	// https://github.com/tesseract-ocr/tesseract/issues/1600
	numWorkers := runtime.NumCPU() / 4
	log.Printf("Spawning %d workers\n", numWorkers)
	var wgWorker sync.WaitGroup
	wgWorker.Add(numWorkers)
	for w := 0; w < numWorkers; w++ {
		go ocr.Worker(w+1, dbName, queue, &wgWorker)
	}

	wgWorker.Wait()
	log.Printf("All jobs finished!")
}
