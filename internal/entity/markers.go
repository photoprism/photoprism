package entity

type Markers []Marker

// Save stores the markers in the database.
func (m Markers) Save(fileID uint) error {
	for _, marker := range m {
		if fileID > 0 {
			marker.FileID = fileID
		}

		if _, err := UpdateOrCreateMarker(&marker); err != nil {
			return err
		}
	}

	return nil
}

// Contains returns true if a marker at the same position already exists.
func (m Markers) Contains(m2 Marker) bool {
	const d = 0.07

	for _, m1 := range m {
		if m2.X > (m1.X-d) && m2.X < (m1.X+d) && m2.Y > (m1.Y-d) && m2.Y < (m1.Y+d) {
			return true
		}
	}

	return false
}

// FaceCount returns the number of valid face markers.
func (m Markers) FaceCount() int {
	result := 0
	for _, marker := range m {
		if !marker.MarkerInvalid && marker.MarkerType == MarkerFace {
			result++
		}
	}

	return result
}

// FindMarkers returns all markers for a given file id.
func FindMarkers(fileID uint) (Markers, error) {
	m := Markers{}
	err := Db().Where(`file_id = ?`, fileID).Order("id").Offset(0).Limit(1000).Find(&m).Error

	return m, err
}
