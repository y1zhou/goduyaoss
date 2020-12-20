package ocr

import (
	"log"
	"os"
	"sync"

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
	resVersion := fileOCR(f.Name(), client)

	// last row is the timestamp
	imgTimestamp := cropImage(img, 0, img.Cols()/2, img.Rows()-rowHeight, img.Rows())
	defer imgTimestamp.Close()
	gocv.IMWrite(f.Name(), imgTimestamp)
	configTesseract(client, "", true)
	resTimestamp := fileOCR(f.Name(), client)

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
