package ocr

import (
	"image"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/y1zhou/goduyaoss/pkg/db"
	"gocv.io/x/gocv"
)

// Job defines the OCR task to run
type Job struct {
	NetProvider string   // 电信/联通/移动
	Provider    string   // service provider from the <h2> title
	Image       gocv.Mat // image used for OCR
}

// Result is sent to the queue to be stored in the database.
type Result struct {
	NetProvider string
	Provider    string
	Timestamp   time.Time
	Table       [][]string
}

// AddJob puts jobs to a queue for Worker to process.
func AddJob(queue chan Job, img image.Image, netProvider string, provider string) {
	imgMat := ImgToMat(img)

	queue <- Job{
		NetProvider: netProvider, Provider: provider, Image: imgMat,
	}
}

// Worker performs OCR on the tables and add results to a channel.
func Worker(DB *sqlx.DB, queue chan Job, res chan Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range queue {
		timestamp := GetMetadata(job.Image)
		lastTime := db.QueryTime(DB, job.NetProvider, job.Provider)
		if timestamp.After(lastTime) {
			log.Printf("Running OCR on: %s -> %s\n", job.NetProvider, job.Provider)
			jobTable := ImgToTable(job.Image)

			res <- Result{
				NetProvider: job.NetProvider,
				Provider:    job.Provider,
				Timestamp:   timestamp,
				Table:       jobTable,
			}
		} else {
			log.Printf("%s -> %s is up to date\n", job.NetProvider, job.Provider)
		}
	}
}
