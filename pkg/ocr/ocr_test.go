package ocr

import (
	"testing"
	"time"

	"github.com/otiai10/gosseract"
)

func TestCleanTimestamp(t *testing.T) {
	s := "Generated at 2020-12-11 20:30:03"
	ans, _ := time.Parse("2006-01-02T15:04:05", "2020-12-11T20:30:03")
	res := cleanTimestamp(&s)
	if res != ans {
		t.Fatalf("Found timestamp %q, should be %q\n", res, ans)
	}
}

func TestGetIntersections(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

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

func TestFileOCR(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	convertToGrayscale(img)

	client := gosseract.NewClient()
	defer client.Close()

	// Testcase from the "Remarks" column
	imgName := cropImage(img, 64, 479, 60, 90)
	defer imgName.Close()

	configTesseract(client, "Remarks", false, false)
	nodeName := imgOCR(imgName, client)
	trueName := "*Ultimate|IEPL-BGP广新01|3.0|INF* - 1063 单端口"
	if nodeName != trueName {
		t.Errorf("OCR text is %q, but should be %q", nodeName, trueName)
	}

	// Testcase from the "AvgSpeed" column
	imgSpeed := cropImage(img, 807, 897, 60, 90)
	defer imgSpeed.Close()

	configTesseract(client, "AvgSpeed", true, false)
	nodeSpeed := imgOCR(imgSpeed, client)
	if nodeSpeed != "21.48MB" {
		t.Errorf("OCR text is %q, but should be 21.48MB", nodeSpeed)
	}
}

func TestGetMetadata(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	timestamp := GetMetadata(img)

	ans, _ := time.Parse("2006-01-02T15:04:05", "2020-12-11T20:30:03")
	if timestamp != ans {
		t.Errorf("Timestamp detected is %q", timestamp)
	}
}

func TestImgToTable(t *testing.T) {
	img := readImg("testdata/sample_img.png")
	defer img.Close()

	res := ImgToTable(img)

	if len(res) != 7 {
		t.Errorf("Should be 7 columns, found %d\n", len(res))
	}
	for i := range res {
		if len(res[i]) != 45 {
			t.Errorf("Should be 45 rows, found %d in column %d\n",
				len(res[0]), i)
		}
	}

}
