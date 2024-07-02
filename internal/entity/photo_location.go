package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/internal/tensorflow/classify"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/geo"
	"github.com/photoprism/photoprism/pkg/txt"
	"gopkg.in/photoprism/go-tz.v2/tz"
)

// SetCoordinates changes the photo lat, lng and altitude if not empty and from an acceptable source.
func (m *Photo) SetCoordinates(lat, lng float32, altitude float64, source string) {
	m.SetAltitude(altitude, source)

	if lat == 0.0 && lng == 0.0 {
		return
	}

	if SrcPriority[source] < SrcPriority[m.PlaceSrc] && m.HasLatLng() {
		return
	}

	m.PhotoLat = lat
	m.PhotoLng = lng
	m.PlaceSrc = source
}

// SetAltitude sets the photo altitude if not empty and from an acceptable source.
func (m *Photo) SetAltitude(altitude float64, source string) {
	a := clean.Altitude(altitude)

	if a == 0 && source != SrcManual {
		return
	}

	if SrcPriority[source] < SrcPriority[m.PlaceSrc] {
		return
	}

	m.PhotoAltitude = a
}

// UnknownLocation tests if the photo has an unknown location.
func (m *Photo) UnknownLocation() bool {
	return m.CellID == "" || m.CellID == UnknownLocation.ID || m.NoLatLng()
}

// SetPosition sets a position estimate.
func (m *Photo) SetPosition(pos geo.Position, source string, force bool) {
	if SrcPriority[m.PlaceSrc] > SrcPriority[source] && !force {
		return
	} else if pos.Lat == 0 && pos.Lng == 0 {
		return
	}

	if m.CellID != UnknownID && pos.InRange(float64(m.PhotoLat), float64(m.PhotoLng), geo.Meter*50) {
		log.Debugf("photo: %s keeps position %f, %f", m.String(), m.PhotoLat, m.PhotoLng)
	} else {
		if pos.Estimate {
			pos.Randomize(geo.Meter * 5)
		}

		m.PhotoLat = float32(pos.Lat)
		m.PhotoLng = float32(pos.Lng)
		m.PlaceSrc = source
		m.CellAccuracy = pos.Accuracy
		m.SetAltitude(pos.Altitude, source)

		log.Debugf("photo: %s %s", m.String(), pos.String())

		m.UpdateLocation()

		if m.Place == nil {
			log.Warnf("photo: failed to update position of %s", m)
		} else {
			log.Debugf("photo: approximate place of %s is %s (id %s)", m, clean.Log(m.Place.Label()), m.PlaceID)
		}
	}
}

// AdoptPlace sets the place based on another photo.
func (m *Photo) AdoptPlace(other *Photo, source string, force bool) {
	if other == nil {
		return
	} else if SrcPriority[m.PlaceSrc] > SrcPriority[source] && !force {
		return
	} else if other.Place == nil {
		return
	} else if other.Place.Unknown() {
		return
	}

	// Remove existing location labels if place changes.
	if other.Place.ID != m.PlaceID {
		m.RemoveLocationLabels()
	}

	m.RemoveLocation(source, force)

	m.Place = other.Place
	m.PlaceID = other.PlaceID
	m.PhotoCountry = other.PhotoCountry
	m.PlaceSrc = source

	m.UpdateTimeZone(other.TimeZone)

	log.Debugf("photo: %s now located at %s (id %s)", m.String(), clean.Log(m.Place.Label()), m.PlaceID)
}

// RemoveLocation removes the current location.
func (m *Photo) RemoveLocation(source string, force bool) {
	if SrcPriority[m.PlaceSrc] > SrcPriority[source] && !force {
		return
	}

	// Reset latitude and longitude.
	m.PhotoLat = 0
	m.PhotoLng = 0

	// Reset cell reference.
	m.Cell = &UnknownLocation
	m.CellID = UnknownLocation.ID
	m.CellAccuracy = 0

	// Reset country code.
	m.PhotoCountry = UnknownCountry.ID

	// Reset place reference.
	m.Place = &UnknownPlace
	m.PlaceID = UnknownPlace.ID

	// Reset place source.
	m.PlaceSrc = SrcAuto
}

// RemoveLocationLabels removes existing location labels.
func (m *Photo) RemoveLocationLabels() {
	if len(m.Labels) == 0 {
		res := Db().Delete(PhotoLabel{}, "photo_id = ? AND label_src = ?", m.ID, SrcLocation)

		if res.Error != nil {
			Log("photo", "remove location labels", res.Error)
		} else if res.RowsAffected > 0 {
			log.Infof("photo: removed %s from %s",
				english.Plural(int(res.RowsAffected), "location label", "location labels"), m)
		}

		return
	}

	labels := make([]PhotoLabel, 0, len(m.Labels))

	for _, l := range m.Labels {
		if l.LabelSrc != SrcLocation {
			labels = append(labels, l)
			continue
		}

		Log("photo", "remove location label", l.Delete())
	}

	removed := len(m.Labels) - len(labels)

	if removed > 0 {
		log.Infof("photo: removed %s from %s",
			english.Plural(removed, "location label", "location labels"), m)
		m.Labels = labels
	}
}

// HasLocation tests if the photo has a known location.
func (m *Photo) HasLocation() bool {
	return !m.UnknownLocation()
}

// TrustedLocation tests if the photo has a known location from a trusted source.
func (m *Photo) TrustedLocation() bool {
	return m.HasLocation() && SrcPriority[m.PlaceSrc] > SrcPriority[SrcEstimate]
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
	return geo.Position{Name: m.String(), Time: m.TakenAt.UTC(),
		Lat: float64(m.PhotoLat), Lng: float64(m.PhotoLng), Altitude: float64(m.PhotoAltitude)}
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
	location := txt.TimeZone(m.TimeZone)

	if location == nil {
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
	location := txt.TimeZone(m.TimeZone)

	if location == nil {
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
		var loc = NewCell(m.PhotoLat, m.PhotoLng)

		if loc.Unknown() {
			// Empty or unknown S2 cell id... should not happen, unless coordinates are invalid.
			log.Warnf("photo: unknown cell id for lat %f, lng %f (uid %s)", m.PhotoLat, m.PhotoLng, m.PhotoUID)
		} else if err := loc.Find(GeoApi); err != nil {
			log.Errorf("photo: %s (find location)", err)
		} else if loc.Place == nil {
			log.Warnf("photo: failed fetching geo data (uid %s, cell %s)", m.PhotoUID, loc.ID)
		} else if loc.ID != UnknownLocation.ID {
			changed := m.CellID != loc.ID

			if changed {
				log.Debugf("photo: changing location of %s from %s to %s", m.String(), m.CellID, loc.ID)
				m.RemoveLocationLabels()
			}

			m.Cell = loc
			m.CellID = loc.ID
			m.Place = loc.Place
			m.PlaceID = loc.PlaceID
			m.PhotoCountry = loc.CountryCode()

			if changed && m.TakenSrc != SrcManual {
				m.UpdateTimeZone(m.GetTimeZone())
			}

			FirstOrCreateCountry(NewCountry(loc.CountryCode(), loc.CountryName()))

			locCategory := loc.Category()
			keywords = append(keywords, loc.Keywords()...)

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

	if m.UnknownCountry() && m.CellID == UnknownID && m.PlaceID == UnknownID {
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
		log.Errorf("photo: %s %s while syncing keywords and labels", m.String(), err)
	}

	if err := m.UpdateTitle(m.ClassifyLabels()); err != nil {
		log.Info(err)
	}

	if err := m.IndexKeywords(); err != nil {
		log.Errorf("photo: %s %s while indexing keywords", m.String(), err)
	}

	return m.Save()
}
