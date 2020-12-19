package ocr

import (
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"

	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

var (
	rowHeight            = 30
	colWidthAvgSpeed     = 90
	colWidthUDPNAT       = 200
	defaultTesseractConf = "--psm 7 --oem 3"
	white                = color.RGBA{255, 255, 255, 0}
)

// The first 6 columns are always:
//   "Group", "Remarks", "Loss", "Ping", "Google Ping", and "AvgSpeed"
// In some cases, there is a 7th column at the end:
//   "UDP NAT Type"
// In some rare cases, there are two more columns at the end:
//   "MaxSpeed" and "UDP NAT Type"
var charWhitelist = map[string]string{
	"Loss":         "0123456789%.",
	"Ping":         "0123456789.",
	"Google Ping":  "0123456789.",
	"AvgSpeed":     "0123456789.KMGBNA",
	"MaxSpeed":     "0123456789.KMGBNA",
	"UDP NAT Type": "- ABDFNOPRSTUacdeiklmnoprstuwy", // See https://github.com/arantonitis/pynat/blob/c5fe553bbbb79deecedcce83c4d4d2974b139355/pynat.py#L51-L59
}

// readImg always reads an image assuming it's colored. The IMReadGrayScale
// flag is not used because it produces inconsistent results compared with
// the output of convertToGrayscale().
func readImg(imgPath string) gocv.Mat {
	return gocv.IMRead(imgPath, gocv.IMReadColor)
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

func textOCR(imgPath string, client *gosseract.Client) string {
	if err := client.SetImage(imgPath); err != nil {
		log.Fatal(err)
	}
	text, _ := client.Text()

	return text
}

func configTesseract(client *gosseract.Client, whitelistKey string, engOnly bool) {
	if engOnly {
		client.SetLanguage("eng")
	} else {
		client.SetLanguage("chi_sim", "eng")
	}
	client.SetPageSegMode(gosseract.PSM_SINGLE_LINE)

	whitelist, _ := charWhitelist[whitelistKey]
	client.SetWhitelist(whitelist) // sets whitelist to "" if key not in map
}

// getMetadata retrieves information from the image that only need to be run once:
// The SSRSpeed software version at the very top,
// the "Group" (all rows have the same value), and
// the time the image was generated (timestamp in the last row).
func getMetadata(img gocv.Mat, client *gosseract.Client, cols []int) (string, string, string) {
	// Temporary file to store the cropped images
	f, err := ioutil.TempFile("", "goduyaoss-*.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	// SSRSpeed version
	imgVersion := cropImage(img, 0, img.Cols(), 0, rowHeight)
	defer imgVersion.Close()
	gocv.IMWrite(f.Name(), imgVersion)
	configTesseract(client, "", true)
	resVersion := textOCR(f.Name(), client)

	// Group name
	imgGroup := cropImage(img, cols[0], cols[1], 2*rowHeight, 3*rowHeight)
	defer imgGroup.Close()
	gocv.IMWrite(f.Name(), imgGroup)
	configTesseract(client, "", false)
	resGroup := textOCR(f.Name(), client)

	// last row is the timestamp
	imgTimestamp := cropImage(img, 0, img.Cols()/2, img.Rows()-rowHeight, img.Rows())
	defer imgTimestamp.Close()
	gocv.IMWrite(f.Name(), imgTimestamp)
	configTesseract(client, "", true)
	resTimestamp := textOCR(f.Name(), client)

	return resVersion, resGroup, resTimestamp
}

func getColOCRConfig() {}
