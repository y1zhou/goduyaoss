package ocr

import (
	"testing"

	"gocv.io/x/gocv"
)

// not really a test. Check the output file and see if there's changes in the boarder.
func TestEnhanceBorders(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	enhanceBorders(img)
	gocv.IMWrite("cache/sample_img_border.png", img)
}

// not really a test. Check the output file and see if it's grayscale.
func TestConvertToGrayscale(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	convertToGrayscale(img)
	gocv.IMWrite("cache/sample_img_gray.png", img)
}

// not really a test. Check the output file and see if it's black and white.
func TestConvertToBin(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	convertToGrayscale(img)
	convertToBin(img)
	gocv.IMWrite("cache/sample_img_binary.png", img)
}

func TestDetectLinesMorph(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	convertToGrayscale(img)
	convertToBin(img)
	hLines, vLines := detectLinesMorph(img)
	gocv.IMWrite("cache/hLines.png", hLines)
	gocv.IMWrite("cache/vLines.png", vLines)
}
