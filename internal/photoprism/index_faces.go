package photoprism

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
)

// DetectFaces finds faces in JPEG media files and returns them.
func DetectFaces(jpeg *MediaFile, expected int) (face.Faces, error) {
	if jpeg == nil {
		return face.Faces{}, fmt.Errorf("missing media file")
	}

	start := time.Now()

	engineName := face.ActiveEngineName()

	var thumbSize thumb.Name

	if engineName == face.EngineONNX || Config().ThumbSizePrecached() < 1280 {
		thumbSize = thumb.Fit720
	} else {
		thumbSize = thumb.Fit1280
	}

	thumbName, err := jpeg.Thumbnail(Config().ThumbCachePath(), thumbSize)

	if err != nil {
		log.Debugf("vision: %s in %s (detect faces)", err, clean.Log(jpeg.BaseName()))
		return face.Faces{}, err
	}

	if thumbName == "" {
		log.Debugf("vision: thumb %s not found in %s (detect faces)", thumbSize, clean.Log(jpeg.BaseName()))
		return face.Faces{}, fmt.Errorf("thumbnail %s not found", thumbSize)
	}

	faces, err := vision.DetectFaces(thumbName, Config().FaceSize(), true, expected)

	if err != nil {
		log.Debugf("vision: %s in %s (detect faces)", err, clean.Log(jpeg.BaseName()))
	}

	if l := len(faces); l > 0 {
		log.Infof("vision: found %s in %s [%s]", english.Plural(l, "face", "faces"), clean.Log(jpeg.BaseName()), time.Since(start))
	}

	return faces, err
}

// ApplyDetectedFaces persists detected faces on the given file and updates face
// counts. When XMP face-tag import is enabled and a media file is provided, it
// also reconciles XMP face regions onto the markers so deferred face detection
// (vision/metadata workers) imports XMP names the same way indexing does.
// file and m MUST be the photo's primary file: XMP regions attach to the
// primary file only, and m is used to read its embedded XMP and .xmp sidecar.
func ApplyDetectedFaces(m *MediaFile, file *entity.File, faces face.Faces) (saved bool, count int, err error) {
	if file == nil {
		return false, 0, fmt.Errorf("faces: file is nil")
	}

	importXmp := m != nil && Config().Settings().Index.Faces

	if len(faces) == 0 && !importXmp {
		return false, 0, nil
	}

	if len(faces) > 0 {
		file.AddFaces(faces)
	}

	if importXmp && file.FileHash != "" {
		reconcileXmpFaces(collectXmpFaces(m), file, file.Markers())
	}

	savedMarkers, saveErr := file.SaveMarkers()

	if saveErr != nil {
		return false, 0, saveErr
	}

	if savedMarkers == 0 {
		return false, 0, nil
	}

	if count, err = file.UpdatePhotoFaceCount(); err != nil {
		return true, count, err
	}

	return true, count, nil
}

// Faces finds faces in JPEG media files and returns them.
func (ind *Index) Faces(jpeg *MediaFile, expected int) face.Faces {
	faces, _ := DetectFaces(jpeg, expected)
	return faces
}
