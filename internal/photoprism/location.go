package photoprism

import (
	"errors"

	"github.com/photoprism/photoprism/internal/entity"
)

// Location returns the Location of a MediaFile.
func (m *MediaFile) Location() (*entity.Location, error) {
	if m.location != nil {
		return m.location, nil
	}

	data := m.MetaData()

	if data.Lat == 0 && data.Lng == 0 {
		return nil, errors.New("mediafile: no latitude and longitude in metadata")
	}

	m.location = entity.NewLocation(data.Lat, data.Lng)

	return m.location, nil
}
