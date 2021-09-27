package entity

import (
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/pkg/txt"
)

type Markers []Marker

// Save stores the markers in the database.
func (m Markers) Save(file *File) (count int, err error) {
	for _, marker := range m {
		if file != nil {
			marker.FileUID = file.FileUID
		}

		if _, err := UpdateOrCreateMarker(&marker); err != nil {
			log.Errorf("markers: %s (save)", err)
		}
	}

	if file == nil {
		return len(m), nil
	}

	return file.UpdatePhotoFaceCount()
}

// Unsaved tests if any marker hasn't been saved yet.
func (m Markers) Unsaved() bool {
	for _, marker := range m {
		if marker.Unsaved() {
			return true
		}
	}

	return false
}

// Contains returns true if a marker at the same position already exists.
func (m Markers) Contains(other Marker) bool {
	for _, marker := range m {
		if marker.OverlapPercent(other) > face.OverlapThreshold {
			return true
		}
	}

	return false
}

// DetectedFaceCount returns the number of automatically detected face markers.
func (m Markers) DetectedFaceCount() (count int) {
	for _, marker := range m {
		if marker.DetectedFace() {
			count++
		}
	}

	return count
}

// ValidFaceCount returns the number of valid face markers.
func (m Markers) ValidFaceCount() (count int) {
	for _, marker := range m {
		if marker.ValidFace() {
			count++
		}
	}

	return count
}

// SubjectNames returns known subject names.
func (m Markers) SubjectNames() (names []string) {
	for _, marker := range m {
		if marker.MarkerInvalid || marker.MarkerType != MarkerFace {
			continue
		} else if n := marker.SubjectName(); n != "" {
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

	for _, marker := range m {
		if marker.ValidFace() {
			faceCount++

			if u := marker.Uncertainty(); u < labelUncertainty {
				labelUncertainty = u
			}

			if marker.MarkerSrc != "" {
				labelSrc = marker.MarkerSrc
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

// FindMarkers returns up to 1000 markers for a given file uid.
func FindMarkers(fileUID string) (Markers, error) {
	m := Markers{}

	err := Db().
		Where(`file_uid = ?`, fileUID).
		Order("x").
		Offset(0).Limit(1000).
		Find(&m).Error

	return m, err
}
