package photoprism

import (
	"errors"
)

// TimeZone returns the time zone where the photo was taken.
func (m *MediaFile) TimeZone() (string, error) {
	meta, err := m.MetaData()

	if err != nil {
		return "UTC", errors.New("mediafile: unknown time zone, using UTC")
	}

	return meta.TimeZone, nil
}
