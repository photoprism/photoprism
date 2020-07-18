package thumb

import (
	"image"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/pkg/txt"
)

func Jpeg(srcFilename, jpgFilename string) (img image.Image, err error) {
	img, err = imaging.Open(srcFilename, imaging.AutoOrientation(true))

	if err != nil {
		log.Errorf("resample: can't open %s", txt.Quote(filepath.Base(srcFilename)))
		return img, err
	}

	saveOption := imaging.JPEGQuality(JpegQuality)

	if err = imaging.Save(img, jpgFilename, saveOption); err != nil {
		log.Errorf("resample: failed to save %s", txt.Quote(filepath.Base(jpgFilename)))
		return img, err
	}

	return img, nil
}
