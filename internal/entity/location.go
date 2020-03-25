package entity

import (
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

var locationMutex = sync.Mutex{}

// Location used to associate photos to location
type Location struct {
	ID          string `gorm:"type:varbinary(16);primary_key;auto_increment:false;"`
	PlaceID     string `gorm:"type:varbinary(16);"`
	Place       *Place
	LocName     string `gorm:"type:varchar(128);"`
	LocCategory string `gorm:"type:varchar(64);"`
	LocSource   string `gorm:"type:varbinary(16);"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Lock location for updates
func (Location) Lock() {
	locationMutex.Lock()
}

// Unlock location for updates
func (Location) Unlock() {
	locationMutex.Unlock()
}

// NewLocation creates a location using a token extracted from coordinate
func NewLocation(lat, lng float64) *Location {
	result := &Location{}

	result.ID = s2.Token(lat, lng)

	return result
}

// Find gets the location using either the db or the api if not in the db
func (m *Location) Find(db *gorm.DB, api string) error {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.First(m, "id = ?", m.ID).Error; err == nil {
		m.Place = FindPlace(m.PlaceID, db)
		return nil
	}

	l := &maps.Location{
		ID: m.ID,
	}

	if err := l.QueryApi(api); err != nil {
		return err
	}

	m.Place = FindPlaceByLabel(l.ID, l.LocLabel, db)

	if m.Place.NoID() {
		m.Place.ID = l.ID
		m.Place.LocLabel = l.LocLabel
		m.Place.LocCity = l.LocCity
		m.Place.LocState = l.LocState
		m.Place.LocCountry = l.LocCountry
	}

	m.LocName = l.LocName
	m.LocCategory = l.LocCategory
	m.LocSource = l.LocSource

	if err := db.Create(m).Error; err != nil {
		log.Errorf("location: %s", err)
		return err
	}

	return nil
}

// Keywords computes keyword based on a Location
func (m *Location) Keywords() (result []string) {
	result = append(result, txt.Keywords(m.City())...)
	result = append(result, txt.Keywords(m.State())...)
	result = append(result, txt.Keywords(m.CountryName())...)
	result = append(result, txt.Keywords(m.Category())...)

	result = append(result, txt.Keywords(m.Name())...)
	result = append(result, txt.Keywords(m.Label())...)
	result = append(result, txt.Keywords(m.Notes())...)

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
