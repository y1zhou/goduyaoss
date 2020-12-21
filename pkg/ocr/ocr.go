package ocr

import (
	"log"
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

func imgOCR(imgMat gocv.Mat, client *gosseract.Client) string {
	// Mat -> image.Image
	imgMatClone := imgMat.Clone()
	defer imgMatClone.Close()
	img, err := imgMatClone.ToImage()
	if err != nil {
		log.Fatalf("Can't convert gocv.Mat to image.Image: %q", err)
	}

	if err := client.SetImageFromBytes(imgToBytes(img)); err != nil {
		log.Fatal(err)
	}
	text, _ := client.Text()

	return text
}

func configTesseract(client *gosseract.Client, whitelistKey string, engOnly bool, colMode bool) {
	if engOnly {
		client.SetLanguage("eng")
	} else {
		client.SetLanguage("chi_sim", "eng")
	}
	if colMode {
		client.SetPageSegMode(gosseract.PSM_SINGLE_BLOCK)
	}
	client.SetPageSegMode(gosseract.PSM_SINGLE_LINE)

	whitelist, _ := charWhitelist[whitelistKey]
	client.SetWhitelist(whitelist) // sets whitelist to "" if key not in map
}

// GetMetadata retrieves information from the image that only need to be run once:
// The SSRSpeed software version at the very top, and
// the time the image was generated (timestamp in the last row).
func GetMetadata(img gocv.Mat, client *gosseract.Client) (string, string) {
	// Convert to grayscale
	imgGray := img.Clone()
	defer imgGray.Close()
	convertToGrayscale(imgGray)

	// SSRSpeed version
	imgVersion := cropImage(imgGray, 0, imgGray.Cols(), 0, rowHeight)
	defer imgVersion.Close()
	configTesseract(client, "", true, false)
	resVersion := imgOCR(imgVersion, client)

	// last row is the timestamp
	imgTimestamp := cropImage(imgGray,
		0, imgGray.Cols()/2,
		imgGray.Rows()-rowHeight, imgGray.Rows())
	defer imgTimestamp.Close()
	configTesseract(client, "", true, false)
	resTimestamp := imgOCR(imgTimestamp, client)

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
func ImgToTable(img gocv.Mat) [][]string {
	rows, cols := GetBorderIndex(img)

	// Sanity check
	numRows, numCols := len(rows)-1, len(cols)-1
	header := GetHeader(numCols)

	// Group name stays the same for all rows
	imgGroup := cropImage(img, cols[0], cols[1], 2*rowHeight, 3*rowHeight)
	defer imgGroup.Close()

	client := gosseract.NewClient()
	defer client.Close()
	configTesseract(client, "", false, false)
	txtGroup := imgOCR(imgGroup, client)

	// OCR - no need to parse first two and last two rows.
	res := make([][]string, numRows-4)
	var wg sync.WaitGroup

	for i := 2; i < numRows-2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			row := make([]string, numCols)
			row[0] = txtGroup
			localClient := gosseract.NewClient()
			defer localClient.Close()
			// Skip first column because it's already handled
			for j := 1; j < numCols; j++ {
				if j == 1 {
					configTesseract(localClient, header[j], false)
				} else {
					configTesseract(localClient, header[j], true)
				}
				cell := cropImage(img, cols[j], cols[j+1], rows[i], rows[i+1])
				defer cell.Close()

				text := imgOCR(cell, localClient)
				row[j] = text
			}
			res[i-2] = row
		}(i)
	}
	wg.Wait()
	return res
}
