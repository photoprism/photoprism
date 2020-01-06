package thumb

import (
	"image"

	"github.com/disintegration/imaging"
)

func Jpeg(srcFilename, jpgFilename string) (result image.Image, err error) {
	img, err := imaging.Open(srcFilename, imaging.AutoOrientation(true))

	if err != nil {
		log.Errorf("thumbs: can't open %s", srcFilename)
		return result, err
	}

	var saveOption imaging.EncodeOption
	saveOption = imaging.JPEGQuality(JpegQuality)

	err = imaging.Save(img, jpgFilename, saveOption)

	if err != nil {
		log.Errorf("thumbs: failed to save %s", jpgFilename)
		return result, err
	}

	return result, nil
}
