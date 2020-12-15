package ocr

import (
	"image"
	"image/png"
	"os"

	"gocv.io/x/gocv"
)

const (
	rowHeight            int    = 30
	colWidthAvgSpeed     int    = 90
	colWidthUDPNAT       int    = 200
	defaultTesseractConf string = "--psm 7 --oem 3"
)

func readImg(filepath string) image.Image {
	f, _ := os.Open(filepath)
	defer f.Close()

	img, _ := png.Decode(f)
	return img
}

func saveImg(img image.Image, filepath string) {
	f, _ := os.Create(filepath)
	defer f.Close()

	png.Encode(f, img)
}

func convertToGrayscale(img image.Image) *image.Gray {
	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}
	return grayImg
}

func readToBw(filepath string) gocv.Mat {
	img := gocv.IMRead(filepath, gocv.IMReadGrayScale)
	bwImg := gocv.NewMat() // don't forget to close it later

	gocv.AdaptiveThreshold(img, &bwImg, 255, gocv.AdaptiveThresholdMean, gocv.ThresholdBinary, 11, 2)

	return bwImg
}

func enhanceBoarders() {}

func detectLinesMorph() {}

func getIntersections() {}

func textOCR() {}

func cropImage() {}

func getColOCRConfig() {}
