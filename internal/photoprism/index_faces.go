package photoprism

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
)

// DetectFaces finds faces in JPEG media files and returns them.
func DetectFaces(jpeg *MediaFile, expected int) (face.Faces, error) {
	if jpeg == nil {
		return face.Faces{}, fmt.Errorf("missing media file")
	}

	start := time.Now()

	// Always Fit720, which entity.ClusterSizeCond depends on: markers.size is recorded in the
	// pixels of whatever image detection ran on, and it is a lower bound on the extent an
	// embedding was sampled from only while that image is the narrowest rendition.
	thumbName, err := jpeg.Thumbnail(Config().ThumbCachePath(), thumb.Fit720)

	if err != nil {
		log.Debugf("%s (detect faces)", err)
		return face.Faces{}, err
	}

	if thumbName == "" {
		log.Debugf("vision: thumb %s not found in %s (detect faces)", thumb.Fit720, clean.Log(jpeg.BaseName()))
		return face.Faces{}, fmt.Errorf("thumbnail %s not found", thumb.Fit720)
	}

	faces, err := vision.DetectFaces(thumbName, Config().FaceSize(), Config().FaceSizeRetry(), true, expected)

	if err != nil {
		log.Debugf("vision: %s in %s (detect faces)", err, clean.Log(jpeg.BaseName()))
	}

	if l := len(faces); l > 0 {
		log.Infof("vision: found %s in %s [%s]", english.Plural(l, "face", "faces"), clean.Log(jpeg.BaseName()), time.Since(start))
	}

	return faces, err
}

// ApplyXmpFaces imports XMP face regions from a still-image source onto the
// primary file, persists marker changes, and updates the photo face count.
func ApplyXmpFaces(m *MediaFile, file *entity.File) (saved bool, count int, err error) {
	if file == nil {
		return false, 0, fmt.Errorf("faces: file is nil")
	} else if m == nil || file.FileHash == "" {
		return false, 0, nil
	}

	regions, collectErr := collectXmpFaces(m)
	if collectErr != nil {
		return false, 0, collectErr
	}

	return applyXmpFaceRegions(regions, file)
}

// applyXmpFaceRegions reconciles already collected XMP face regions onto the
// primary file. It is separate from ApplyXmpFaces so the indexer can collect
// first and skip the primary-file lookup when a source declares no regions.
func applyXmpFaceRegions(regions meta.FaceRegions, file *entity.File) (saved bool, count int, err error) {
	if file == nil {
		return false, 0, fmt.Errorf("faces: file is nil")
	} else if file.FileHash == "" {
		return false, 0, nil
	}

	changed, reconcileErr := reconcileXmpFaces(regions, file, file.Markers())
	if reconcileErr != nil {
		return false, 0, reconcileErr
	} else if changed == 0 {
		return false, 0, nil
	}

	if _, err = file.SaveMarkers(); err != nil {
		return false, 0, err
	}

	count, err = file.UpdatePhotoFaceCount()

	return true, count, err
}

// ApplyDetectedFaces persists detected faces on the given primary file and
// updates face counts. When XMP face-tag import is enabled, m may be either the
// primary preview or a related still-image source; both resolve to the same
// embedded and sidecar XMP collection.
func ApplyDetectedFaces(m *MediaFile, file *entity.File, faces face.Faces) (saved bool, count int, err error) {
	if file == nil {
		return false, 0, fmt.Errorf("faces: file is nil")
	}

	importXmp := m != nil && Config().XMPFaces()

	if len(faces) == 0 && !importXmp {
		return false, 0, nil
	}

	if len(faces) > 0 {
		file.AddFaces(faces)
	}

	xmpChanges := 0
	if importXmp && file.FileHash != "" {
		// XMP import is independent of AI detection, so a malformed sidecar or an
		// unreadable source must not discard the detected faces added above: log
		// and continue to SaveMarkers instead of returning the error.
		if regions, collectErr := collectXmpFaces(m); collectErr != nil {
			log.Warnf("faces: %s while reading xmp regions for %s", clean.Error(collectErr), clean.Log(m.BaseName()))
		} else if changed, reconcileErr := reconcileXmpFaces(regions, file, file.Markers()); reconcileErr != nil {
			log.Warnf("faces: %s while reconciling xmp regions for %s", clean.Error(reconcileErr), clean.Log(m.BaseName()))
		} else {
			xmpChanges = changed
		}
	}

	savedMarkers, saveErr := file.SaveMarkers()

	if saveErr != nil {
		return false, 0, saveErr
	}

	if savedMarkers == 0 && xmpChanges == 0 {
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
