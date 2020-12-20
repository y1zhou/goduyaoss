package ocr

import (
	"os"
	"testing"

	"github.com/otiai10/gosseract"
	"gocv.io/x/gocv"
)

// Check the output file and see if there's changes in the boarder.
func BenchmarkEnhanceBorders(t *testing.B) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	enhanceBorders(img)
	gocv.IMWrite("cache/sample_img_enhanced.png", img)
}

// Check if the output file is grayscale.
func BenchmarkConvertToGrayscale(t *testing.B) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	convertToGrayscale(img)
	gocv.IMWrite("cache/sample_img_gray.png", img)
}

// Check if the output file is black and white.
func BenchmarkConvertToBin(t *testing.B) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	convertToGrayscale(img)
	convertToBin(img)
	gocv.IMWrite("cache/sample_img_binary.png", img)
}

// Check if the borders in the output file match the original borders.
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

	if len(rows) != 50 {
		t.Errorf("Should be 50 horizontal lines, found %d\n", len(rows))
	}
	if len(cols) != 8 {
		t.Errorf("Should be 8 vertical lines, found %d\n", len(cols))
	}
}

func TestTextOCR(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	convertToGrayscale(img)

	f := createTempfile("")
	defer f.Close()
	defer os.Remove(f.Name())

	client := gosseract.NewClient()
	defer client.Close()

	// Testcase from the "Remarks" column
	imgName := cropImage(img, 64, 479, 60, 90)
	defer imgName.Close()
	gocv.IMWrite(f.Name(), imgName)

	configTesseract(client, "Remarks", false)
	nodeName := fileOCR(f.Name(), client)
	trueName := "*Ultimate|IEPL-BGP广新01|3.0|INF* - 1063 单端口"
	if nodeName != trueName {
		t.Errorf("OCR text is %q, but should be %q", nodeName, trueName)
	}

	// Testcase from the "AvgSpeed" column
	imgSpeed := cropImage(img, 807, 897, 60, 90)
	defer imgSpeed.Close()
	gocv.IMWrite(f.Name(), imgSpeed)

	configTesseract(client, "AvgSpeed", true)
	nodeSpeed := fileOCR(f.Name(), client)
	if nodeSpeed != "21.48MB" {
		t.Errorf("OCR text is %q, but should be 21.48MB", nodeSpeed)
	}
}

func TestGetMetadata(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()
	convertToGrayscale(img)
	client := gosseract.NewClient()
	defer client.Close()

	version, timestamp := GetMetadata(img, client)

	if version != "SSRSpeed Result Table (v2.7.2)" {
		t.Errorf("Version detected is %q", version)
	}
	if timestamp != "Generated at 2020-12-11 20:30:03" {
		t.Errorf("Timestamp detected is %q", timestamp)
	}
}

func TestImgToTable(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	enhanceBorders(img)
	convertToGrayscale(img)
	res := ImgToTable(img)

	if len(res) != 45 {
		t.Errorf("Should be 45 rows, found %d", len(res))
	}
	if len(res[0]) != 7 {
		t.Errorf("Should be 7 columns, found %d", len(res[0]))
	}
}
