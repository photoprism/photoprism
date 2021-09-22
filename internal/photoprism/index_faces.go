package photoprism

import (
	"time"

	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/thumb"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Faces finds faces in JPEG media files and returns them.
func (ind *Index) Faces(jpeg *MediaFile, expected int) face.Faces {
	if jpeg == nil {
		return face.Faces{}
	}

	var minSize int
	var thumbSize thumb.Name

	// Select best thumbnail depending on configured size.
	if Config().ThumbSizePrecached() < 1280 {
		minSize = 20
		thumbSize = thumb.Fit720
	} else {
		minSize = 30
		thumbSize = thumb.Fit1280
	}

	thumbName, err := jpeg.Thumbnail(Config().ThumbPath(), thumbSize)

	if err != nil {
		log.Debugf("index: %s in %s (faces)", err, txt.Quote(jpeg.BaseName()))
		return face.Faces{}
	}

	if thumbName == "" {
		log.Debugf("index: thumb %s not found in %s (faces)", thumbSize, txt.Quote(jpeg.BaseName()))
		return face.Faces{}
	}

	start := time.Now()

	faces, err := ind.faceNet.Detect(thumbName, minSize, true, expected)

	if err != nil {
		log.Debugf("%s in %s", err, txt.Quote(jpeg.BaseName()))
	}

	if l := len(faces); l == 1 {
		log.Infof("index: found %d face in %s [%s]", l, txt.Quote(jpeg.BaseName()), time.Since(start))
	} else if l > 1 {
		log.Infof("index: found %d faces in %s [%s]", l, txt.Quote(jpeg.BaseName()), time.Since(start))
	}

	return faces
}
