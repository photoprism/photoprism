package photoprism

import (
	"errors"

	"github.com/photoprism/photoprism/internal/entity"
)

// Location returns the S2 cell entity of a MediaFile.
func (m *MediaFile) Location() (*entity.Cell, error) {
	if m.location != nil {
		return m.location, nil
	}

	data := m.MetaData()

	if data.Lat == 0 && data.Lng == 0 {
		return nil, errors.New("media: found no latitude and longitude")
	}

	m.location = entity.NewCell(data.Lat, data.Lng)

	return m.location, nil
}
