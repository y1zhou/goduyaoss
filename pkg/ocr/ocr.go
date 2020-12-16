package ocr

import (
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

var (
	rowHeight            = 30
	colWidthAvgSpeed     = 90
	colWidthUDPNAT       = 200
	defaultTesseractConf = "--psm 7 --oem 3"
	white                = color.RGBA{255, 255, 255, 0}
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
// - Vertical lines are detected through morphological operations.
// White border lines are drawn on a black background.
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
	gocv.BitwiseNot(vLines, &vLines)

	return hLines, vLines
}

// getIntersections returns two arrays of indices of the table border lines.
func getIntersections(hLines gocv.Mat, vLines gocv.Mat) ([]int, []int) {
	height, width := hLines.Rows(), hLines.Cols()
	numRows := height / rowHeight

	// We already know where the horizontal lines are
	var iHorizontalLines []int
	for i := 0; i < height; i++ {
		if hLines.GetUCharAt(i, 0) != 0 {
			iHorizontalLines = append(iHorizontalLines, i)
		}
	}

	// For vertical lines, we only consider the rows where the hLines are.
	// If there's >50% of white pixels in a column at those rows, then it's
	// very likely to be a vertical line.

	// focus on rows where the hLines are
	crossPoints := gocv.NewMat()
	defer crossPoints.Close()
	gocv.BitwiseAnd(hLines, vLines, &crossPoints)

	// count non-zero values of each column in those rows
	colSums := make([]int, width)
	for j := 0; j < width; j++ {
		for _, num := range iHorizontalLines {
			if crossPoints.GetUCharAt(num, j) != 0 {
				colSums[j]++
			}
		}
	}

	var possibleCols []int
	for j, num := range colSums {
		if num > numRows/5 {
			possibleCols = append(possibleCols, j)
		}
	}

	// remove clusters of vertical lines
	var colsDiff []int
	for i := 0; i < len(possibleCols)-1; i++ {
		colsDiff = append(colsDiff, possibleCols[i+1]-possibleCols[i])
	}

	var iVerticalLines []int
	for i, num := range colsDiff {
		if num > 10 {
			iVerticalLines = append(iVerticalLines, possibleCols[i])
		}
	}

	// trick to split the last two columns. The colored "AvgSpeed" column
	// is the hardest to identify
	for width-iVerticalLines[len(iVerticalLines)-1] > colWidthUDPNAT {
		iVerticalLines = append(iVerticalLines, iVerticalLines[len(iVerticalLines)-1]+colWidthAvgSpeed)
	}
	if width-iVerticalLines[len(iVerticalLines)-1] > colWidthAvgSpeed {
		iVerticalLines = append(iVerticalLines, width-1)
	}

	return iHorizontalLines, iVerticalLines
}

func textOCR() {}

func cropImage() {}

func getColOCRConfig() {}
