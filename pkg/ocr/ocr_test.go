package ocr

import (
	"testing"

	"gocv.io/x/gocv"
)

func TestConvertToGrayscale(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	grayImg := convertToGrayscale(img)

	saveImg(grayImg, "cache/sample_img_gray.png")
}

func TestReadToBw(t *testing.T) {
	bwImg := readToBw("cache/sample_img_gray.png")
	defer bwImg.Close()
	gocv.IMWrite("cache/sample_img_bw.png", bwImg)
}
