package ocr

import (
	"image"
	"image/color"
	"log"
	"os"

	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

var (
	rowHeight        = 30
	colWidthAvgSpeed = 90
	colWidthUDPNAT   = 200
	white            = color.RGBA{255, 255, 255, 0}
)

// The first 6 columns are always:
//   "Group", "Remarks", "Loss", "Ping", "Google Ping", and "AvgSpeed"
// In some cases, there is a 7th column at the end:
//   "UDP NAT Type"
// In some rare cases, there are two more columns at the end:
//   "MaxSpeed" and "UDP NAT Type"
var charWhitelist = map[string]string{
	"loss":         "0123456789%.",
	"ping":         "0123456789.",
	"google_ping":  "0123456789.",
	"avg_speed":    "0123456789.KMGBNA",
	"max_speed":    "0123456789.KMGBNA",
	"udp_nat_type": "- ABDFNOPRSTUacdeiklmnoprstuwy", // See https://github.com/arantonitis/pynat/blob/c5fe553bbbb79deecedcce83c4d4d2974b139355/pynat.py#L51-L59
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

// GetMetadata retrieves information from the image that only need to be run once:
// The SSRSpeed software version at the very top, and
// the time the image was generated (timestamp in the last row).
func GetMetadata(img gocv.Mat, client *gosseract.Client) (string, string) {
	// Temporary file to store the cropped images
	f := createTempfile("")
	defer f.Close()
	defer os.Remove(f.Name())

	// SSRSpeed version
	imgVersion := cropImage(img, 0, img.Cols(), 0, rowHeight)
	defer imgVersion.Close()
	gocv.IMWrite(f.Name(), imgVersion)
	configTesseract(client, "", true)
	resVersion := textOCR(f.Name(), client)

	// last row is the timestamp
	imgTimestamp := cropImage(img, 0, img.Cols()/2, img.Rows()-rowHeight, img.Rows())
	defer imgTimestamp.Close()
	gocv.IMWrite(f.Name(), imgTimestamp)
	configTesseract(client, "", true)
	resTimestamp := textOCR(f.Name(), client)

	return resVersion, resTimestamp
}

// GetHeader returns the column names based on the number of columns.
func GetHeader(numCols int) []string {
	var header = []string{"group", "remarks", "loss", "ping", "google_ping", "avg_speed"}
	switch numCols {
	case 6:
		break
	case 7:
		header = append(header, "udp_nat_type")
		break
	case 8:
		header = append(header, "max_speed", "udp_nat_type")
		break
	default:
		log.Fatalf("%d columns detected (should be 6-8)", numCols)
	}

	return header
}

// ImgToTable runs Tesseract on each cell and returns a parsed table.
func ImgToTable(img gocv.Mat, client *gosseract.Client) [][]string {
	// Convert to grayscale
	enhanceBorders(img)
	convertToGrayscale(img)

	imgGray := img.Clone()
	defer imgGray.Close()

	// Detect table borders
	convertToBin(img)
	hLines, vLines := detectLinesMorph(img)
	defer hLines.Close()
	defer vLines.Close()

	rows, cols := getIntersections(hLines, vLines)

	// Sanity check
	numRows, numCols := len(rows)-1, len(cols)-1
	header := GetHeader(numCols)

	// Temp file for saving cropped images
	f := createTempfile("")
	defer f.Close()
	defer os.Remove(f.Name())

	// Group name stays the same for all rows
	imgGroup := cropImage(img, cols[0], cols[1], 2*rowHeight, 3*rowHeight)
	defer imgGroup.Close()
	gocv.IMWrite(f.Name(), imgGroup)
	configTesseract(client, "", false)
	resGroup := textOCR(f.Name(), client)

	// OCR
	var res [][]string
	// No need to parse first two and last two rows
	for i := 2; i < numRows-2; i++ {
		var row = []string{resGroup}
		// Skip first column because it's handled in `getMetadata()`
		for j := 1; j < numCols; j++ {
			if j == 1 {
				configTesseract(client, header[j], false)
			} else {
				configTesseract(client, header[j], true)
			}
			cell := cropImage(imgGray, cols[j], cols[j+1], rows[i], rows[i+1])
			gocv.IMWrite(f.Name(), cell)
			text := textOCR(f.Name(), client)
			row = append(row, text)
		}
		res = append(res, row)
	}
	return res
}
