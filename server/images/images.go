package images

import (
	"image"
	"io"

	"golang.org/x/image/bmp"
	"golang.org/x/image/draw"
)

func GetImageSize(img image.Image) (int, int) {
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy()
}

func ToBitmap(img image.Image, width, height int, writer *io.Writer) (err error) {

	imgWidth, imgHeight := GetImageSize(img)

	// First we crop to the correct aspect ratio

	desiredAspectRatio := float64(width) / float64(height)
	currentAspectRatio := float64(imgWidth) / float64(imgHeight)

	crop := img.Bounds()
	if currentAspectRatio < desiredAspectRatio {
		// Image is too wide, crop y
		cropHeight := int(float64(imgHeight) / desiredAspectRatio)
		crop = image.Rect(0, (imgHeight-cropHeight)/2, imgWidth, (imgHeight-cropHeight)/2+cropHeight)
	} else if currentAspectRatio > desiredAspectRatio {
		// Image is too narrow, crop x
		cropWidth := int(float64(imgWidth) * desiredAspectRatio)
		crop = image.Rect((imgWidth-cropWidth)/2, 0, (imgWidth-cropWidth)/2+cropWidth, imgHeight)
	}

	// then we resize it to the correct size
	resized := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.CatmullRom.Scale(resized, resized.Bounds(), img, crop, draw.Over, nil)

	bmp.Encode(*writer, resized)

	return nil
}
