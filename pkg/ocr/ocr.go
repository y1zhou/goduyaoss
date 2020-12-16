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

// readImg always reads an image assuming it's colored. The IMReadGrayScale
// flag is not used because it produces inconsistent results compared with
// the output of convertToGrayscale().
func readImg(filepath string) gocv.Mat {
	return gocv.IMRead(filepath, gocv.IMReadColor)
}

// enhanceBorders makes the edges easier to detect.
func enhanceBorders(img gocv.Mat) {
	// convert to HSV color space to build mask
	imgHSV := gocv.NewMat()
	defer imgHSV.Close()
	gocv.CvtColor(img, &imgHSV, gocv.ColorBGRToHSV)

	lowerGray, _ := gocv.NewMatFromBytes(3, 1, gocv.MatTypeCV8U, []byte{0, 0, 0})
	defer lowerGray.Close()
	upperGray, _ := gocv.NewMatFromBytes(3, 1, gocv.MatTypeCV8U, []byte{179, 50, 200})
	defer upperGray.Close()

	grayMask := gocv.NewMat()
	defer grayMask.Close()
	gocv.InRange(imgHSV, lowerGray, upperGray, &grayMask)
	gocv.BitwiseNot(grayMask, &grayMask) // invert mask

	res := gocv.NewMat()
	defer res.Close()
	gocv.BitwiseAndWithMask(img, img, &res, grayMask)
	res.CopyTo(&img)
}

func convertToGrayscale(img gocv.Mat) {
	gocv.CvtColor(img, &img, gocv.ColorBGRToGray)
}

// convertToBin applies a threshold to each pixel to transform a grayscale
// image to a binary one.
func convertToBin(img gocv.Mat) {
	gocv.AdaptiveThreshold(img, &img, 255, gocv.AdaptiveThresholdMean, gocv.ThresholdBinary, 11, 2)
}

// detectLinesMorph uses different approaches to find horizontal and vertical lines:
// - Horizontal lines are assumed to be at every other 30px.
// - Vertical lines are detected through morphological operations
func detectLinesMorph(imgBin gocv.Mat) (gocv.Mat, gocv.Mat) {
	rows, cols := imgBin.Rows(), imgBin.Cols()

	// Horizontal lines
	hLines := gocv.NewMatWithSize(rows, cols, gocv.MatTypeCV8U)
	for i := 0; i < rows; i += 30 {
		gocv.Line(&hLines, image.Point{0, i}, image.Point{cols, i}, white, 1)
	}
	// just to make sure the line at the bottom is added
	gocv.Line(&hLines, image.Point{0, rows - 1}, image.Point{cols, rows - 1}, white, 1)

	// Vertical lines
	vLines := gocv.NewMatWithSize(rows, cols, gocv.MatTypeCV8U)
	vertKernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{1, rows / rowHeight})
	gocv.Erode(imgBin, &vLines, vertKernel)
	gocv.Dilate(imgBin, &vLines, vertKernel)

	return hLines, vLines
}

func getIntersections() {}

func textOCR() {}

func cropImage() {}

func getColOCRConfig() {}
