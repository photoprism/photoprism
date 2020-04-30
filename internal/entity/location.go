package entity

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Location used to associate photos to location
type Location struct {
	ID          string `gorm:"type:varbinary(16);primary_key;auto_increment:false;"`
	PlaceID     string `gorm:"type:varbinary(16);"`
	Place       *Place
	LocName     string `gorm:"type:varchar(255);"`
	LocCategory string `gorm:"type:varchar(64);"`
	LocSource   string `gorm:"type:varbinary(16);"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewLocation creates a location using a token extracted from coordinate
func NewLocation(lat, lng float32) *Location {
	result := &Location{}

	result.ID = s2.Token(float64(lat), float64(lng))

	return result
}

// Find gets the location using either the db or the api if not in the db
func (m *Location) Find(api string) error {
	db := Db()

	if err := db.Preload("Place").First(m, "id = ?", m.ID).Error; err == nil {
		return nil
	}

	l := &maps.Location{
		ID: m.ID,
	}

	if err := l.QueryApi(api); err != nil {
		return err
	}

	if place := FindPlaceByLabel(l.S2Token(), l.Label()); place != nil {
		m.Place = place
	} else {
		m.Place = &Place{
			ID:          l.S2Token(),
			LocLabel:    l.Label(),
			LocCity:     l.City(),
			LocState:    l.State(),
			LocCountry:  l.CountryCode(),
			LocKeywords: l.KeywordString(),
		}
	}

	m.LocName = l.Name()
	m.LocCategory = l.Category()
	m.LocSource = l.Source()

	if err := db.Create(m).Error; err == nil {
		return nil
	} else if err := db.Preload("Place").First(m, "id = ?", m.ID).Error; err != nil {
		return err
	}

	return nil
}

// Keywords computes keyword based on a Location
func (m *Location) Keywords() (result []string) {
	if m.Place == nil {
		log.Errorf("location: place for %s is nil - you might have found a bug", m.ID)
		return result
	}

	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.City(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.State(), "-"))...)
	result = append(result, txt.Keywords(txt.ReplaceSpaces(m.CountryName(), "-"))...)
	result = append(result, txt.Keywords(m.Category())...)
	result = append(result, txt.Keywords(m.Name())...)
	result = append(result, txt.Keywords(m.Place.LocKeywords)...)

	result = txt.UniqueWords(result)

	return result
}

// Unknown checks if the location has no id
func (m *Location) Unknown() bool {
	return m.ID == ""
}

// Name returns name of location
func (m *Location) Name() string {
	return m.LocName
}

// NoName checks if the location has no name
func (m *Location) NoName() bool {
	return m.LocName == ""
}

// Category returns the location category
func (m *Location) Category() string {
	return m.LocCategory
}

// NoCategory checks id the location has no category
func (m *Location) NoCategory() bool {
	return m.LocCategory == ""
}

// Label returns the location place label
func (m *Location) Label() string {
	return m.Place.Label()
}

// City returns the location place city
func (m *Location) City() string {
	return m.Place.City()
}

// LongCity checks if the city name is more than 16 char
func (m *Location) LongCity() bool {
	return len(m.City()) > 16
}

// NoCity checks if the location has no city
func (m *Location) NoCity() bool {
	return m.City() == ""
}

// CityContains checks if the location city contains the text string
func (m *Location) CityContains(text string) bool {
	return strings.Contains(text, m.City())
}

// State returns the location place state
func (m *Location) State() string {
	return m.Place.State()
}

// NoState checks if the location place has no state
func (m *Location) NoState() bool {
	return m.Place.State() == ""
}

// CountryCode returns the location place country code
func (m *Location) CountryCode() string {
	return m.Place.CountryCode()
}

// CountryName returns the location place country name
func (m *Location) CountryName() string {
	return m.Place.CountryName()
}

// Notes returns the locations place notes
func (m *Location) Notes() string {
	return m.Place.Notes()
}

// Source returns the source of location information
func (m *Location) Source() string {
	return m.LocSource
}
