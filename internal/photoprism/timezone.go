package photoprism

import (
	"errors"

	"gopkg.in/ugjka/go-tz.v2/tz"
)

// TimeZone returns the time zone where the photo was taken.
func (m *MediaFile) TimeZone() (string, error) {
	meta, err := m.Exif()

	if err != nil {
		return "UTC", errors.New("no image metadata")
	}

	if meta.Lat == 0 && meta.Long == 0 {
		return "UTC", errors.New("no latitude and longitude in image metadata")
	}

	zones, err := tz.GetZone(tz.Point{
		Lon: meta.Long, Lat: meta.Lat,
	})

	if err != nil {
		return "UTC", errors.New("no matching zone found")
	}

	return zones[0], nil
}
