package ocr

import (
	"image"
	"log"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

// Job defines the OCR task to run
type Job struct {
	NetProvider string   // 电信/联通/移动
	Provider    string   // service provider from the <h2> title
	Image       gocv.Mat // image used for OCR
}

func AddJob(queue chan Job, img image.Image, netProvider string, provider string) {
	imgMat := ImgToMat(img)

	queue <- Job{
		NetProvider: netProvider, Provider: provider, Image: imgMat,
	}
}

func Worker(queue chan Job, res chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range queue {
		log.Printf("Parsing %q -> %q\n", job.NetProvider, job.Provider)
		time.Sleep(5 * time.Second)

		res <- job.Provider
	}
}

// func runOCR(item Job) string {
// 	client := gosseract.NewClient()
// 	configTesseract(client, item.headerKey, item.engOnly, item.colMode)
// 	text := imgOCR(item.image, client)
// 	client.Close()

// 	return text
// }
