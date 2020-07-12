package entity

import (
	"time"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/maps"
	"gopkg.in/ugjka/go-tz.v2/tz"
)

// GetTimeZone uses PhotoLat and PhotoLng to guess the time zone of the photo.
func (m *Photo) GetTimeZone() string {
	result := "UTC"

	if m.HasLatLng() {
		zones, err := tz.GetZone(tz.Point{
			Lat: float64(m.PhotoLat),
			Lon: float64(m.PhotoLng),
		})

		if err == nil && len(zones) > 0 {
			result = zones[0]
		}
	}

	return result
}

// CountryName returns the photo country name.
func (m *Photo) CountryName() string {
	if name, ok := maps.CountryNames[m.CountryCode()]; ok {
		return name
	}

	return UnknownCountry.CountryName
}

// CountryCode returns the photo country code.
func (m *Photo) CountryCode() string {
	if len(m.PhotoCountry) != 2 {
		m.PhotoCountry = UnknownCountry.ID
	}

	return m.PhotoCountry
}

// GetTakenAt returns UTC time for TakenAtLocal.
func (m *Photo) GetTakenAt() time.Time {
	loc, err := time.LoadLocation(m.TimeZone)

	if err != nil {
		return m.TakenAt
	}

	if takenAt, err := time.ParseInLocation("2006-01-02T15:04:05", m.TakenAtLocal.Format("2006-01-02T15:04:05"), loc); err != nil {
		return m.TakenAt
	} else {
		return takenAt.UTC()
	}
}

// UpdateLocation updates location and labels based on latitude and longitude.
func (m *Photo) UpdateLocation(geoApi string) (keywords []string, labels classify.Labels) {
	if m.HasLatLng() {
		var location = NewCell(m.PhotoLat, m.PhotoLng)

		err := location.Find(geoApi)

		if location.Place == nil {
			log.Warnf("photo: location place is nil (uid %s, location %s) - bug?", m.PhotoUID, location.ID)
		}

		if err == nil && location.Place != nil && location.ID != UnknownLocation.ID {
			m.Cell = location
			m.CellID = location.ID
			m.Place = location.Place
			m.PlaceID = location.PlaceID
			m.PhotoCountry = location.CountryCode()

			if m.TakenSrc != SrcManual {
				m.TimeZone = m.GetTimeZone()
				m.TakenAt = m.GetTakenAt()
			}

			FirstOrCreateCountry(NewCountry(location.CountryCode(), location.CountryName()))

			locCategory := location.Category()
			keywords = append(keywords, location.Keywords()...)

			// Append category from reverse location lookup
			if locCategory != "" {
				labels = append(labels, classify.LocationLabel(locCategory, 0, -1))
			}

			return keywords, labels
		}
	}

	keywords = []string{}
	labels = classify.Labels{}

	if m.UnknownLocation() {
		m.Cell = &UnknownLocation
		m.CellID = UnknownLocation.ID
	} else if err := m.LoadLocation(); err == nil {
		m.Place = m.Cell.Place
		m.PlaceID = m.Cell.PlaceID
	} else {
		log.Warn(err)
	}

	if m.UnknownPlace() {
		m.Place = &UnknownPlace
		m.PlaceID = UnknownPlace.ID
	} else if err := m.LoadPlace(); err == nil {
		m.PhotoCountry = m.Place.CountryCode()
	} else {
		log.Warn(err)
	}

	if m.UnknownCountry() {
		m.EstimateCountry()
	}

	if m.HasCountry() {
		FirstOrCreateCountry(NewCountry(m.CountryCode(), m.CountryName()))
	}

	return keywords, labels
}
