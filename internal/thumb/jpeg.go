package thumb

import (
	"image"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/pkg/clean"
)

func Jpeg(srcFilename, jpgFilename string, orientation int) (img image.Image, err error) {
	img, err = imaging.Open(srcFilename)

	if err != nil {
		log.Errorf("resample: cannot open %s", clean.Log(filepath.Base(srcFilename)))
		return img, err
	}

	if orientation > 1 {
		img = Rotate(img, orientation)
	}

	quality := JpegQuality.EncodeOption()

	if err = imaging.Save(img, jpgFilename, quality); err != nil {
		log.Errorf("resample: failed to save %s", clean.Log(filepath.Base(jpgFilename)))
		return img, err
	}

	return img, nil
}
