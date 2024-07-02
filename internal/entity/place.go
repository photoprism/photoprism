package entity

import (
	"strings"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/remote/maps"
	"github.com/photoprism/photoprism/pkg/clean"
)

var placeMutex = sync.Mutex{}

// Place represents a distinct region identified by city, district, state, and country.
type Place struct {
	ID            string    `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"PlaceID" yaml:"PlaceID"`
	PlaceLabel    string    `gorm:"type:VARCHAR(400);" json:"Label" yaml:"Label"`
	PlaceDistrict string    `gorm:"type:VARCHAR(100);index;" json:"District" yaml:"District,omitempty"`
	PlaceCity     string    `gorm:"type:VARCHAR(100);index;" json:"City" yaml:"City,omitempty"`
	PlaceState    string    `gorm:"type:VARCHAR(100);index;" json:"State" yaml:"State,omitempty"`
	PlaceCountry  string    `gorm:"type:VARBINARY(2);" json:"Country" yaml:"Country,omitempty"`
	PlaceKeywords string    `gorm:"type:VARCHAR(300);" json:"Keywords" yaml:"Keywords,omitempty"`
	PlaceFavorite bool      `json:"Favorite" yaml:"Favorite,omitempty"`
	PhotoCount    int       `gorm:"default:1" json:"PhotoCount" yaml:"-"`
	CreatedAt     time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt     time.Time `json:"UpdatedAt" yaml:"-"`
}

// TableName returns the entity table name.
func (Place) TableName() string {
	return "places"
}

// UnknownPlace is PhotoPrism's default place.
var UnknownPlace = Place{
	ID:            UnknownID,
	PlaceLabel:    "Unknown",
	PlaceDistrict: "Unknown",
	PlaceCity:     "Unknown",
	PlaceState:    "Unknown",
	PlaceCountry:  UnknownID,
	PlaceKeywords: "",
	PlaceFavorite: false,
	PhotoCount:    -1,
}

// CreateUnknownPlace creates the default place if not exists.
func CreateUnknownPlace() {
	UnknownPlace = *FirstOrCreatePlace(&UnknownPlace)
}

// FindPlace finds a matching place or returns nil.
func FindPlace(id string) *Place {
	m := Place{}

	if Db().First(&m, "id = ?", id).Error != nil {
		log.Debugf("place: %s not found", clean.Log(id))
		return nil
	}

	return &m
}

// Create inserts a new row to the database.
func (m *Place) Create() error {
	placeMutex.Lock()
	defer placeMutex.Unlock()

	return Db().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Place) Save() error {
	placeMutex.Lock()
	defer placeMutex.Unlock()

	return Db().Save(m).Error
}

// Delete removes the entity from the index.
func (m *Place) Delete() (err error) {
	return UnscopedDb().Delete(m).Error
}

// FirstOrCreatePlace fetches an existing row, inserts a new row or nil in case of errors.
func FirstOrCreatePlace(m *Place) *Place {
	if m.ID == "" {
		log.Errorf("place: id must not be empty (find or create)")
		return nil
	}

	if m.PlaceLabel == "" {
		log.Errorf("place: label must not be empty (find or create place %s)", m.ID)
		return nil
	}

	result := Place{}

	if findErr := Db().Where("id = ?", m.ID).First(&result).Error; findErr == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := Db().Where("id = ?", m.ID).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("place: %s (create %s)", createErr, m.ID)
	}

	return nil
}

// Unknown returns true if this is an unknown place
func (m Place) Unknown() bool {
	return m.ID == "" || m.ID == UnknownPlace.ID
}

// Label returns place label
func (m Place) Label() string {
	return m.PlaceLabel
}

// District returns the place district name if any.
func (m Place) District() string {
	return m.PlaceDistrict
}

// City returns place city if any.
func (m Place) City() string {
	return m.PlaceCity
}

// LongCity checks if the city name is more than 16 char.
func (m Place) LongCity() bool {
	return len(m.PlaceCity) > 16
}

// NoCity checks if the location has no city
func (m Place) NoCity() bool {
	return m.PlaceCity == ""
}

// CityContains checks if the location city contains the text string
func (m Place) CityContains(text string) bool {
	return strings.Contains(text, m.PlaceCity)
}

// State returns place State
func (m Place) State() string {
	return m.PlaceState
}

// CountryCode returns place CountryCode
func (m Place) CountryCode() string {
	return m.PlaceCountry
}

// CountryName returns place CountryName
func (m Place) CountryName() string {
	return maps.CountryNames[m.PlaceCountry]
}
