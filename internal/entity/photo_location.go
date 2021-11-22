package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/pkg/geo"
	"github.com/photoprism/photoprism/pkg/txt"
	"gopkg.in/photoprism/go-tz.v2/tz"
)

// UnknownLocation tests if the photo has an unknown location.
func (m *Photo) UnknownLocation() bool {
	return m.CellID == "" || m.CellID == UnknownLocation.ID || m.NoLatLng()
}

// RemoveLocation removes the current location.
func (m *Photo) RemoveLocation() {
	m.PhotoLat = 0
	m.PhotoLng = 0
	m.Cell = &UnknownLocation
	m.CellID = UnknownLocation.ID
}

// HasLocation tests if the photo has a known location.
func (m *Photo) HasLocation() bool {
	return !m.UnknownLocation()
}

// LocationLoaded tests if the photo has a known location that is currently loaded.
func (m *Photo) LocationLoaded() bool {
	if m.Cell == nil {
		return false
	}

	if m.Cell.Place == nil {
		return false
	}

	return !m.Cell.Unknown() && m.Cell.ID == m.CellID
}

// LoadLocation loads the photo location from the database if not done already.
func (m *Photo) LoadLocation() error {
	if m.LocationLoaded() {
		return nil
	}

	if m.UnknownLocation() {
		return fmt.Errorf("photo: unknown location (%s)", m)
	}

	var location Cell

	err := Db().Preload("Place").First(&location, "id = ?", m.CellID).Error

	if err != nil {
		return err
	}

	if location.Place == nil {
		location.Place = &UnknownPlace
		location.PlaceID = UnknownPlace.ID
	}

	m.Cell = &location

	return nil
}

// PlaceLoaded checks if the photo has a known place that is currently loaded.
func (m *Photo) PlaceLoaded() bool {
	if m.Place == nil {
		return false
	}

	return !m.Place.Unknown() && m.Place.ID == m.PlaceID
}

// LoadPlace loads the photo place from the database if not done already.
func (m *Photo) LoadPlace() error {
	if m.PlaceLoaded() {
		return nil
	}

	if m.UnknownPlace() {
		return fmt.Errorf("photo: unknown place (%s)", m)
	}

	var place Place

	err := Db().First(&place, "id = ?", m.PlaceID).Error

	if err != nil {
		return err
	}

	m.Place = &place

	return nil
}

// Position returns the coordinates as geo.Position.
func (m *Photo) Position() geo.Position {
	if m.NoLatLng() {
		return geo.Position{}
	}

	return geo.Position{Lat: float64(m.PhotoLat), Lng: float64(m.PhotoLng)}
}

// HasLatLng checks if the photo has a latitude and longitude.
func (m *Photo) HasLatLng() bool {
	return m.PhotoLat != 0.0 || m.PhotoLng != 0.0
}

// NoLatLng checks if latitude and longitude are missing.
func (m *Photo) NoLatLng() bool {
	return !m.HasLatLng()
}

// UnknownPlace checks if the photo has an unknown place.
func (m *Photo) UnknownPlace() bool {
	return m.PlaceID == "" || m.PlaceID == UnknownPlace.ID
}

// HasPlace checks if the photo has a known place.
func (m *Photo) HasPlace() bool {
	return !m.UnknownPlace()
}

// HasCountry checks if the photo has a known country.
func (m *Photo) HasCountry() bool {
	return !m.UnknownCountry()
}

// UnknownCountry checks if the photo has an unknown country.
func (m *Photo) UnknownCountry() bool {
	return m.CountryCode() == UnknownCountry.ID
}

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
	location, err := time.LoadLocation(m.TimeZone)

	if err != nil {
		return m.TakenAt
	}

	if takenAt, err := time.ParseInLocation("2006-01-02T15:04:05", m.TakenAtLocal.Format("2006-01-02T15:04:05"), location); err != nil {
		return m.TakenAt
	} else {
		return takenAt.UTC()
	}
}

// GetTakenAtLocal returns local time for TakenAt.
func (m *Photo) GetTakenAtLocal() time.Time {
	location, err := time.LoadLocation(m.TimeZone)

	if err != nil {
		return m.TakenAtLocal
	}

	if takenAtLocal, err := time.ParseInLocation("2006-01-02T15:04:05", m.TakenAt.In(location).Format("2006-01-02T15:04:05"), time.UTC); err != nil {
		return m.TakenAtLocal
	} else {
		return takenAtLocal.UTC()
	}
}

// UpdateLocation updates location and labels based on latitude and longitude.
func (m *Photo) UpdateLocation() (keywords []string, labels classify.Labels) {
	if m.HasLatLng() {
		var location = NewCell(m.PhotoLat, m.PhotoLng)

		err := location.Find(GeoApi)

		if location.Place == nil {
			log.Warnf("photo: failed fetching geo data (uid %s, cell %s)", m.PhotoUID, location.ID)
		} else if err == nil && location.ID != UnknownLocation.ID {
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
				labels = append(labels, classify.LocationLabel(locCategory, 0))
			}

			return keywords, labels
		}
	}

	keywords = []string{}
	labels = classify.Labels{}

	if m.UnknownLocation() {
		m.Cell = &UnknownLocation
		m.CellID = UnknownLocation.ID

		// Remove place estimate if better data is available.
		if SrcPriority[m.PlaceSrc] > SrcPriority[SrcEstimate] {
			m.Place = &UnknownPlace
			m.PlaceID = UnknownPlace.ID
		}
	} else if err := m.LoadLocation(); err == nil {
		m.Place = m.Cell.Place
		m.PlaceID = m.Cell.PlaceID
	} else {
		log.Warnf("photo: location %s not found in %s", m.CellID, m.PhotoName)
	}

	if m.UnknownPlace() {
		m.Place = &UnknownPlace
		m.PlaceID = UnknownPlace.ID
	} else if err := m.LoadPlace(); err == nil {
		m.PhotoCountry = m.Place.CountryCode()
	} else {
		log.Warnf("photo: place %s not found in %s", m.PlaceID, m.PhotoName)
	}

	if m.UnknownCountry() {
		m.EstimateCountry()
	}

	if m.HasCountry() {
		FirstOrCreateCountry(NewCountry(m.CountryCode(), m.CountryName()))
	}

	return keywords, labels
}

// SaveLocation updates location data and saves the photo metadata back to the index.
func (m *Photo) SaveLocation() error {
	locKeywords, labels := m.UpdateLocation()

	m.AddLabels(labels)

	w := txt.UniqueWords(txt.Words(m.GetDetails().Keywords))
	w = append(w, locKeywords...)

	m.GetDetails().Keywords = strings.Join(txt.UniqueWords(w), ", ")

	if err := m.SyncKeywordLabels(); err != nil {
		log.Errorf("photo %s: %s while syncing keywords and labels", m.PhotoUID, err)
	}

	if err := m.UpdateTitle(m.ClassifyLabels()); err != nil {
		log.Info(err)
	}

	if err := m.IndexKeywords(); err != nil {
		log.Errorf("photo %s: %s while indexing keywords", m.PhotoUID, err)
	}

	return m.Save()
}
