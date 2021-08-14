package photoprism

import (
	"time"

	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/pkg/txt"
)

// detectFaces detects faces in a JPEG image and returns them.
func (ind *Index) detectFaces(jpeg *MediaFile) face.Faces {
	if jpeg == nil {
		return face.Faces{}
	}

	var thumbSize string

	// Select best thumbnail depending on configured size.
	if Config().ThumbSize() < 1280 {
		thumbSize = "fit_720"
	} else {
		thumbSize = "fit_1280"
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

	faces, err := ind.faceNet.Detect(thumbName)

	if err != nil {
		log.Debugf("%s in %s", err, txt.Quote(jpeg.BaseName()))
	}

	if len(faces) > 0 {
		log.Infof("index: detected %d faces in %s [%s]", len(faces), txt.Quote(jpeg.BaseName()), time.Since(start))
	}

	return faces
}
