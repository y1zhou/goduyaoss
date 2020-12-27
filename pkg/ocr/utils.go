package ocr

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"

	"gocv.io/x/gocv"
)

func createTempfile(dir string) *os.File {
	f, err := ioutil.TempFile(dir, "goduyaoss-*.png")
	if err != nil {
		log.Fatal(err)
	}
	return f
}

// readImg always reads an image assuming it's colored. The IMReadGrayScale
// flag is not used because it produces inconsistent results compared with
// the output of convertToGrayscale().
func readImg(imgPath string) gocv.Mat {
	return gocv.IMRead(imgPath, gocv.IMReadColor)
}

func imgToBytes(img image.Image) []byte {
	buf := new(bytes.Buffer)
	png.Encode(buf, img)
	return buf.Bytes()
}

// ImgToMat - Takes the `image.Image` from the crawler and convert to `gocv.Mat`.
func ImgToMat(img image.Image) gocv.Mat {
	imgBytes := imgToBytes(img)
	imgMat, err := gocv.IMDecode(imgBytes, gocv.IMReadColor)
	if err != nil {
		log.Fatal(err)
	}
	return imgMat
}

func cleanVersion(s *string) {
	regexVersion := regexp.MustCompile(`^.*v\s*(\d+\.\d+\.\d+).*$`)
	*s = regexVersion.ReplaceAllString(*s, `$1`)
}

func cleanTimestamp(s *string) time.Time {
	regexTime := regexp.MustCompile(`^.*?(\d+-\d+-\d+)\s+(\d+:\d+:\d+).*$`)
	sNew := regexTime.ReplaceAllString(*s, `${1}T$2`)
	res, err := time.Parse("2006-01-02T15:04:05", sNew)
	if err != nil {
		log.Printf("Error parsing timestamp in %q\n", *s)
		return time.Time{}
	}
	return res
}

func PrintTable(t [][]string) {
	headers := GetHeader(len(t))
	fmt.Printf("%s", headers[0])
	for i := 1; i < len(headers); i++ {
		fmt.Printf(",%s", headers[i])
	}
	for i := 0; i < len(t[0]); i++ {
		fmt.Printf("\n%s", t[0][i])
		for j := 1; j < len(t); j++ {
			fmt.Printf(",%s", t[j][i])
		}
	}
	fmt.Printf("\n\n================================================================\n\n")
}
