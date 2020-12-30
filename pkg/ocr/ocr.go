package ocr

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
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

func fileOCR(imgPath string, client *gosseract.Client) string {
	if err := client.SetImage(imgPath); err != nil {
		log.Fatal(err)
	}
	text, _ := client.Text()

	return text
}

func imgOCR(imgMat gocv.Mat, client *gosseract.Client) string {
	// Mat -> image.Image
	imgByte, err := gocv.IMEncode(gocv.PNGFileExt, imgMat)
	if err != nil {
		log.Fatalf("Can't convert gocv.Mat to []byte: %q", err)
	}

	if err := client.SetImageFromBytes(imgByte); err != nil {
		log.Fatalf("Can't send image bytes to Tesseract: %q", err)
	}
	text, err := client.Text()
	if err != nil {
		log.Fatalf("Can't get text from image: %q", err)
	}

	return text
}

func configTesseract(client *gosseract.Client, whitelistKey string, engOnly bool, colMode bool) {
	if engOnly {
		client.SetLanguage("eng")
	} else {
		client.SetLanguage("chi_sim", "eng")
	}
	if colMode {
		client.SetPageSegMode(gosseract.PSM_AUTO)
	} else {
		client.SetPageSegMode(gosseract.PSM_SINGLE_LINE)
	}

	whitelist, _ := charWhitelist[whitelistKey]
	client.SetWhitelist(whitelist) // sets whitelist to "" if key not in map
}

// GetMetadata retrieves information from the image that only need to be run once:
// The SSRSpeed software version at the very top, and
// the time the image was generated (timestamp in the last row).
func GetMetadata(img gocv.Mat) time.Time {
	// Convert to grayscale
	imgGray := img.Clone()
	defer imgGray.Close()
	convertToGrayscale(imgGray)

	// last row is the timestamp
	imgTimestamp := cropImage(imgGray,
		0, imgGray.Cols()/2,
		imgGray.Rows()-rowHeight, imgGray.Rows())
	defer imgTimestamp.Close()

	client := gosseract.NewClient()
	defer client.Close()
	configTesseract(client, "", true, false)
	resTimestr := imgOCR(imgTimestamp, client)
	resTimestamp := cleanTimestamp(&resTimestr)

	return resTimestamp
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
func ImgToTable(img gocv.Mat) [][]string {
	rows, cols := getBorderIndex(img)

	// Remove watermark and background colors
	removeColor(&img, cols)

	// Enhance row borders
	drawRowBorders(&img, rows)

	// Header names
	numRows, numCols := len(rows)-1, len(cols)-1
	header := GetHeader(numCols)

	// Group name stays the same for all rows
	imgGroup := cropImage(img, cols[0], cols[1], rows[2], rows[3])
	defer imgGroup.Close()

	client := gosseract.NewClient()

	configTesseract(client, "", false, false)
	txtGroup := imgOCR(imgGroup, client)

	// Duplicate to make first column
	firstCol := make([]string, numRows-4)
	for i := range firstCol {
		firstCol[i] = txtGroup
	}

	// OCR - no need to parse first two and last two rows.
	res := make([][]string, numCols)
	res[0] = firstCol

	newLineRegex := regexp.MustCompile(`\n+`)
	for j := 1; j < numCols; j++ {
		// Try to use column mode first because it's much faster
		if j == 1 {
			configTesseract(client, header[j], false, true)
		} else {
			configTesseract(client, header[j], true, true)
		}
		col := cropImage(img, cols[j], cols[j+1], rows[2], rows[numRows-2])
		defer col.Close()

		text := imgOCR(col, client)
		text = newLineRegex.ReplaceAllString(text, "\n")
		txtCol := strings.Split(text, "\n")

		// If the number of rows is incorrect, run OCR on each cell.
		// This is much slower but also more accurate.
		if len(txtCol) != numRows-4 {
			txtCol = make([]string, numRows-4)

			for i := 2; i < numRows-2; i++ {
				if j == 1 {
					configTesseract(client, header[j], false, false)
				} else {
					configTesseract(client, header[j], true, false)
				}
				cell := cropImage(img, cols[j], cols[j+1], rows[i], rows[i+1])
				defer cell.Close()

				txtCol[i-2] = imgOCR(cell, client)
			}
		}

		for i := range txtCol {
			txtCol[i] = strings.TrimSpace(txtCol[i])
		}

		res[j] = txtCol
	}
	client.Close()
	return res
}
