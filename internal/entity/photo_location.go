package entity

import (
	"time"

	"github.com/photoprism/photoprism/internal/classify"
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
	var location = NewLocation(m.PhotoLat, m.PhotoLng)

	err := location.Find(geoApi)

	if location.Place == nil {
		log.Warnf("photo: location place is nil (uid %s, loc_uid %s) - bug?", m.PhotoUID, location.LocUID)
	}

	if err == nil && location.Place != nil && location.LocUID != UnknownLocation.LocUID {
		m.Location = location
		m.LocUID = location.LocUID
		m.Place = location.Place
		m.PlaceUID = location.PlaceUID
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
	} else {
		log.Warn(err)

		m.Location = &UnknownLocation
		m.LocUID = UnknownLocation.LocUID
		m.Place = &UnknownPlace
		m.PlaceUID = UnknownPlace.PlaceUID
	}

	if m.Place != nil && (m.PhotoCountry == "" || m.PhotoCountry == UnknownCountry.Code()) {
		m.PhotoCountry = m.Place.LocCountry
	}

	return keywords, labels
}
