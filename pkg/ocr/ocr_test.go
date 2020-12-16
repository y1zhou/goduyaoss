package ocr

import (
	"image"
	"testing"

	"gocv.io/x/gocv"
)

// not really a test. Check the output file and see if there's changes in the boarder.
func BenchmarkEnhanceBorders(t *testing.B) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	enhanceBorders(img)
	gocv.IMWrite("cache/sample_img_enhanced.png", img)
}

// not really a test. Check the output file and see if it's grayscale.
func BenchmarkConvertToGrayscale(t *testing.B) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	convertToGrayscale(img)
	gocv.IMWrite("cache/sample_img_gray.png", img)
}

// not really a test. Check the output file and see if it's black and white.
func BenchmarkConvertToBin(t *testing.B) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	convertToGrayscale(img)
	convertToBin(img)
	gocv.IMWrite("cache/sample_img_binary.png", img)
}

func BenchmarkDetectLinesMorph(t *testing.B) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	enhanceBorders(img)
	convertToGrayscale(img)
	convertToBin(img)
	hLines, vLines := detectLinesMorph(img)
	defer hLines.Close()
	defer vLines.Close()

	borders := gocv.NewMat()
	defer borders.Close()
	gocv.BitwiseOr(hLines, vLines, &borders)
	gocv.IMWrite("cache/borders.png", borders)
}

func TestGetIntersections(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	enhanceBorders(img)
	convertToGrayscale(img)
	convertToBin(img)
	hLines, vLines := detectLinesMorph(img)
	defer hLines.Close()
	defer vLines.Close()

	rows, cols := getIntersections(hLines, vLines)

	res := gocv.NewMatWithSize(img.Rows(), img.Cols(), gocv.MatTypeCV8U)
	for _, i := range rows {
		gocv.Line(&res, image.Point{0, i}, image.Point{img.Cols(), i}, white, 1)
	}
	for _, j := range cols {
		gocv.Line(&res, image.Point{j, 0}, image.Point{j, img.Rows()}, white, 1)
	}
	gocv.IMWrite("cache/intersection.png", res)
}
