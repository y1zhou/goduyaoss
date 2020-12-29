package main

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/y1zhou/goduyaoss/pkg/crawler"
	"github.com/y1zhou/goduyaoss/pkg/db"
	"github.com/y1zhou/goduyaoss/pkg/ocr"
)

func main() {
	var wgWorker sync.WaitGroup
	var wgSaver sync.WaitGroup
	// Each Tesseract process uses a maximum of 4 threads
	// https://github.com/tesseract-ocr/tesseract/issues/1600
	numWorkers := runtime.NumCPU() / 4
	queue := make(chan ocr.Job, numWorkers)
	res := make(chan ocr.Result, numWorkers*2)

	for w := 0; w < numWorkers; w++ {
		wgWorker.Add(1)
		go ocr.Worker(queue, res, &wgWorker)
	}

	// Send jobs to the queue
	go func() {
		defer close(queue)
		for netProvider, url := range crawler.Pages {
			doc := crawler.RequestPage(url)
			providers := crawler.FetchProviders(doc)
			provTest := providers[4:5]

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

	go func() {
		wgSaver.Add(1)
		defer wgSaver.Done()

		DB := db.ConnectDb("test.db")
		for s := range res {
			fmt.Printf("Net provider: %s\nService provider: %s\n\n", s.NetProvider, s.Provider)
			// ocr.PrintTable(s.Table)
			db.InsertRows(DB, s.NetProvider, s.Provider, s.Timestamp, s.Table)
		}
	}()

	// Close the res channel after all workers finish
	wgWorker.Wait()
	close(res)
	wgSaver.Wait()
}
