package entity

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/tensorflow/classify"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Markers represents a list of markers.
type Markers []Marker

// Save stores the markers in the database.
func (m Markers) Save(file *File) (count int, err error) {
	if file == nil {
		return 0, fmt.Errorf("file required for saving markers")
	}

	for i := range m {
		if m[i].UpdateFile(file) {
			continue
		}

		if created, err := CreateMarkerIfNotExists(&m[i]); err != nil {
			log.Errorf("markers: %s (save)", err)
		} else {
			m[i] = *created
		}
	}

	return file.UpdatePhotoFaceCount()
}

// Unsaved tests if any marker hasn't been saved yet.
func (m Markers) Unsaved() bool {
	for i := range m {
		if m[i].Unsaved() {
			return true
		}
	}

	return false
}

// Contains returns true if a marker at the same position already exists.
func (m Markers) Contains(other Marker) bool {
	for i := range m {
		if m[i].OverlapPercent(other) > face.OverlapThreshold {
			return true
		}
	}

	return false
}

// DetectedFaceCount returns the number of automatically detected face markers.
func (m Markers) DetectedFaceCount() (count int) {
	for i := range m {
		if m[i].DetectedFace() {
			count++
		}
	}

	return count
}

// ValidFaceCount returns the number of valid face markers.
func (m Markers) ValidFaceCount() (count int) {
	for i := range m {
		if m[i].ValidFace() {
			count++
		}
	}

	return count
}

// SubjectNames returns known subject names.
func (m Markers) SubjectNames() (names []string) {
	for i := range m {
		if m[i].MarkerInvalid || m[i].MarkerType != MarkerFace {
			continue
		} else if n := m[i].SubjectName(); n != "" {
			names = append(names, n)
		}
	}

	return txt.UniqueNames(names)
}

// Labels returns matching labels.
func (m Markers) Labels() (result classify.Labels) {
	faceCount := 0

	labelSrc := SrcImage
	labelUncertainty := 100

	for i := range m {
		if m[i].ValidFace() {
			faceCount++

			if u := m[i].Uncertainty(); u < labelUncertainty {
				labelUncertainty = u
			}

			if m[i].MarkerSrc != "" {
				labelSrc = m[i].MarkerSrc
			}
		}
	}

	if faceCount < 1 {
		return classify.Labels{}
	}

	var rule classify.LabelRule

	if faceCount == 1 {
		rule = classify.Rules["portrait"]
	} else {
		rule = classify.Rules["people"]
	}

	return classify.Labels{classify.Label{
		Name:        rule.Label,
		Source:      labelSrc,
		Uncertainty: labelUncertainty,
		Priority:    rule.Priority,
		Categories:  rule.Categories,
	}}
}

// Append adds a marker.
func (m *Markers) Append(marker Marker) {
	*m = append(*m, marker)
}

// AppendWithEmbedding adds a marker with face embedding.
func (m *Markers) AppendWithEmbedding(marker Marker) {
	if !marker.Embeddings().One() {
		// Ignore markers that don't have exactly one embedding.
		return
	}

	m.Append(marker)
}

// FindMarkers returns up to 1000 markers for a given file uid.
func FindMarkers(fileUid string) (Markers, error) {
	m := Markers{}

	err := Db().
		Where("file_uid = ?", fileUid).
		Order("x").
		Offset(0).Limit(1000).
		Find(&m).Error

	return m, err
}
