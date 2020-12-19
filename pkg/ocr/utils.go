package ocr

import (
	"bytes"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"

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
