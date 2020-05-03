package thumb

import (
	"image"

	"github.com/disintegration/imaging"
)

func Jpeg(srcFilename, jpgFilename string) (img image.Image, err error) {
	img, err = imaging.Open(srcFilename, imaging.AutoOrientation(true))

	if err != nil {
		log.Errorf("resample: can't open %s", srcFilename)
		return img, err
	}

	saveOption := imaging.JPEGQuality(JpegQuality)

	if err = imaging.Save(img, jpgFilename, saveOption); err != nil {
		log.Errorf("resample: failed to save %s", jpgFilename)
		return img, err
	}

	return img, nil
}
