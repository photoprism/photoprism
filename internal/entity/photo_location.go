package entity

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

// IndexLocation updates location and labels based on latitude and longitude.
func (m *Photo) IndexLocation(db *gorm.DB, geoApi string, labels classify.Labels) ([]string, classify.Labels) {
	var location = NewLocation(m.PhotoLat, m.PhotoLng)

	location.Lock()
	defer location.Unlock()

	var keywords []string

	err := location.Find(db, geoApi)

	if err == nil {
		if location.Place.New {
			event.Publish("count.places", event.Data{
				"count": 1,
			})
		}

		m.Location = location
		m.LocationID = location.ID
		m.Place = location.Place
		m.PlaceID = location.PlaceID
		m.LocationEstimated = false

		country := NewCountry(location.CountryCode(), location.CountryName()).FirstOrCreate(db)

		if country.New {
			event.Publish("count.countries", event.Data{
				"count": 1,
			})
		}

		locCategory := location.Category()
		keywords = append(keywords, location.Keywords()...)

		// Append category from reverse location lookup
		if locCategory != "" {
			labels = append(labels, classify.LocationLabel(locCategory, 0, -1))
		}

		if err := m.UpdateTitle(labels); err != nil {
			log.Warn(err)
		}
	} else {
		log.Warn(err)

		m.Place = UnknownPlace
		m.PlaceID = UnknownPlace.ID
	}

	if m.Place != nil && (!m.ModifiedLocation || m.PhotoCountry == "" || m.PhotoCountry == "zz") {
		m.PhotoCountry = m.Place.LocCountry
	}

	return keywords, labels
}

// UpdateTitle updated the photo title based on location and labels.
func (m *Photo) UpdateTitle(labels classify.Labels) error {
	if m.ModifiedTitle && m.HasTitle() {
		return errors.New("photo: won't update title, was modified")
	}

	hasLocation := m.Location != nil && m.Location.Place != nil

	if hasLocation {
		loc := m.Location

		if title := labels.Title(loc.Name()); title != "" { // TODO: User defined title format
			log.Infof("photo: using label \"%s\" to create photo title", title)
			if loc.NoCity() || loc.LongCity() || loc.CityContains(title) {
				m.PhotoTitle = fmt.Sprintf("%s / %s / %s", txt.Title(title), loc.CountryName(), m.TakenAt.Format("2006"))
			} else {
				m.PhotoTitle = fmt.Sprintf("%s / %s / %s", txt.Title(title), loc.City(), m.TakenAt.Format("2006"))
			}
		} else if loc.Name() != "" && loc.City() != "" {
			if len(loc.Name()) > 45 {
				m.PhotoTitle = txt.Title(loc.Name())
			} else if len(loc.Name()) > 20 || len(loc.City()) > 16 || strings.Contains(loc.Name(), loc.City()) {
				m.PhotoTitle = fmt.Sprintf("%s / %s", loc.Name(), m.TakenAt.Format("2006"))
			} else {
				m.PhotoTitle = fmt.Sprintf("%s / %s / %s", loc.Name(), loc.City(), m.TakenAt.Format("2006"))
			}
		} else if loc.City() != "" && loc.CountryName() != "" {
			if len(loc.City()) > 20 {
				m.PhotoTitle = fmt.Sprintf("%s / %s", loc.City(), m.TakenAt.Format("2006"))
			} else {
				m.PhotoTitle = fmt.Sprintf("%s / %s / %s", loc.City(), loc.CountryName(), m.TakenAt.Format("2006"))
			}
		}
	}

	if !hasLocation || m.NoTitle() {
		if len(labels) > 0 && labels[0].Priority >= -1 && labels[0].Uncertainty <= 85 && labels[0].Name != "" {
			m.PhotoTitle = fmt.Sprintf("%s / %s", txt.Title(labels[0].Name), m.TakenAt.Format("2006"))
		} else if !m.TakenAtLocal.IsZero() {
			m.PhotoTitle = fmt.Sprintf("Unknown / %s", m.TakenAtLocal.Format("2006"))
		} else {
			m.PhotoTitle = "Unknown"
		}

		log.Infof("photo: changed empty photo title to \"%s\"", m.PhotoTitle)
	} else {
		log.Infof("photo: new title is \"%s\"", m.PhotoTitle)
	}

	return nil
}
