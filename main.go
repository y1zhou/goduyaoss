package main

import (
	"log"
	"runtime"
	"sync"

	"github.com/y1zhou/goduyaoss/pkg/crawler"
	"github.com/y1zhou/goduyaoss/pkg/db"
	"github.com/y1zhou/goduyaoss/pkg/ocr"
)

func main() {
	dbName := "test.db"
	log.Printf("Connecting to database: %s\n", dbName)
	DB := db.ConnectDb(dbName)

	var wgWorker sync.WaitGroup
	var wgSaver sync.WaitGroup
	// Each Tesseract process uses a maximum of 4 threads
	// https://github.com/tesseract-ocr/tesseract/issues/1600
	numWorkers := runtime.NumCPU() / 4
	queue := make(chan ocr.Job, numWorkers)
	res := make(chan ocr.Result, numWorkers*2)

	log.Printf("Spawning %d workers\n", numWorkers)
	for w := 0; w < numWorkers; w++ {
		wgWorker.Add(1)
		go ocr.Worker(DB, queue, res, &wgWorker)
	}

	// Send jobs to the queue
	go func() {
		defer close(queue)
		for netProvider, url := range crawler.Pages {
			log.Printf("Requesting page for %s: %s\n", netProvider, url)
			doc := crawler.RequestPage(url)
			providers := crawler.FetchProviders(doc)

			for _, provider := range providers {
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

	go func() {
		wgSaver.Add(1)
		defer wgSaver.Done()

		for s := range res {
			db.InsertRows(DB, s.NetProvider, s.Provider, s.Timestamp, s.Table)
		}
	}()

	// Close the res channel after all workers finish
	wgWorker.Wait()
	close(res)
	wgSaver.Wait()
}
