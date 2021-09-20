package entity

import (
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/pkg/txt"
)

type Markers []Marker

// Save stores the markers in the database.
func (m Markers) Save(fileUID string) error {
	for _, marker := range m {
		if fileUID != "" {
			marker.FileUID = fileUID
		}

		if _, err := UpdateOrCreateMarker(&marker); err != nil {
			return err
		}
	}

	return nil
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

// FaceCount returns the number of valid face markers.
func (m Markers) FaceCount() (faces int) {
	for _, marker := range m {
		if !marker.MarkerInvalid && marker.MarkerType == MarkerFace {
			faces++
		}
	}

	return faces
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
