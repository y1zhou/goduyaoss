package ocr

import (
	"io/ioutil"
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

	imgName := cropImage(img, 64, 479, 60, 90)
	defer imgName.Close()

	f, err := ioutil.TempFile("", "goduyaoss-*.png")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(f.Name())
	gocv.IMWrite(f.Name(), imgName)

	client := gosseract.NewClient()
	defer client.Close()
	nodeName := textOCR(f.Name(), client, "", "", false)
	trueName := "*Ultimate|IEPL-BGP广新01|3.0|INF* - 1063 单端口"
	if nodeName != trueName {
		t.Errorf("OCR text is %q, but should be %q", nodeName, trueName)
	}

	imgSpeed := cropImage(img, 807, 897, 60, 90)
	defer imgSpeed.Close()
	gocv.IMWrite(f.Name(), imgSpeed)

	nodeSpeed := textOCR(f.Name(), client, "", "", false)
	if nodeName != "21.48MB" {
		t.Errorf("OCR text is %q, but should be 21.48MB", nodeSpeed)
	}
	f.Close()
}
