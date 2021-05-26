package entity

type Markers []Marker

// Save stores the markers in the database.
func (m Markers) Save(fileUID string) error {
	for _, marker := range m {
		marker.FileUID = fileUID
		if _, err := UpdateOrCreateMarker(&marker); err != nil {
			return err
		}
	}

	return nil
}
