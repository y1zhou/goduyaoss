# goduyaoss
Golang package to download data from www.duyaoss.com, and perform OCR using [gosseract](https://github.com/otiai10/gosseract) and [GoCV](https://gocv.io/).

## Installation

1. For `gosseract` to work, [Tesseract](https://github.com/tesseract-ocr/tesseract) needs to be installed on your system. It is included in most Linux distributions under the name `tesseract` or `tesseract-ocr`. You should also install two trained langulage data modules for Tesseract: `tesseract-data-eng` and `tesseract-data-chi_sim`. See [their official documentation](https://tesseract-ocr.github.io/tessdoc/) for details.
2. `GoCV` is used to preprocess the images for better OCR results. You must also install OpenCV 4.5.0 on your system. The [documentation of GoCV](https://pkg.go.dev/gocv.io/x/gocv#readme-how-to-install) goes through the process in great detail. Personally I found it necessary to also install the `vtk` and `glew` libraries.
