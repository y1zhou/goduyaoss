package ocr

import (
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

var (
	rowHeight        = 30
	colWidthAvgSpeed = 90
	colWidthUDPNAT   = 200
	white            = color.RGBA{255, 255, 255, 0}
)

// GetBorderIndex returns the indices of the rows and columns.
func getBorderIndex(img gocv.Mat) ([]int, []int) {
	imgGray := img.Clone()
	defer imgGray.Close()

	convertToGrayscale(imgGray)
	convertToBin(imgGray)
	hLines, vLines := detectLinesMorph(imgGray)
	defer hLines.Close()
	defer vLines.Close()
	rows, cols := getIntersections(hLines, vLines)
	return rows, cols
}

// removeColor - vertical colored text in between the "Remarks" and "Loss" columns.
func removeColor(img *gocv.Mat, cols []int) {
	// convert to HSV color space to build mask
	imgHSV := gocv.NewMat()
	defer imgHSV.Close()
	gocv.CvtColor(*img, &imgHSV, gocv.ColorBGRToHSV)

	// Detect gray pixels
	lowerGray, _ := gocv.NewMatFromBytes(3, 1, gocv.MatTypeCV8U, []byte{0, 0, 0})
	defer lowerGray.Close()
	upperGray, _ := gocv.NewMatFromBytes(3, 1, gocv.MatTypeCV8U, []byte{180, 255, 254})
	defer upperGray.Close()

	grayMask := gocv.NewMat()
	defer grayMask.Close()
	gocv.InRange(imgHSV, lowerGray, upperGray, &grayMask)
	gocv.BitwiseNot(grayMask, &grayMask) // invert mask

	grayMask.CopyTo(img)
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
	for i := 0; i < rows; i += rowHeight {
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

func cropImage(img gocv.Mat, x0 int, x1 int, y0 int, y1 int) gocv.Mat {
	rect := image.Rect(x0, y0, x1, y1)
	return img.Region(rect)
}
