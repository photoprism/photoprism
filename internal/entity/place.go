package entity

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Place used to associate photos to places
type Place struct {
	ID            string    `gorm:"type:varbinary(42);primary_key;auto_increment:false;" json:"PlaceID" yaml:"PlaceID"`
	PlaceLabel    string    `gorm:"type:varbinary(768);unique_index;" json:"Label" yaml:"Label"`
	PlaceCity     string    `gorm:"type:varchar(255);" json:"City" yaml:"City,omitempty"`
	PlaceState    string    `gorm:"type:varchar(255);" json:"State" yaml:"State,omitempty"`
	PlaceCountry  string    `gorm:"type:varbinary(2);" json:"Country" yaml:"Country,omitempty"`
	PlaceKeywords string    `gorm:"type:varchar(255);" json:"Keywords" yaml:"Keywords,omitempty"`
	PlaceFavorite bool      `json:"Favorite" yaml:"Favorite,omitempty"`
	PhotoCount    int       `gorm:"default:1" json:"PhotoCount" yaml:"-"`
	CreatedAt     time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt     time.Time `json:"UpdatedAt" yaml:"-"`
}

// UnknownPlace is PhotoPrism's default place.
var UnknownPlace = Place{
	ID:            "zz",
	PlaceLabel:    "Unknown",
	PlaceCity:     "Unknown",
	PlaceState:    "Unknown",
	PlaceCountry:  "zz",
	PlaceKeywords: "",
	PlaceFavorite: false,
	PhotoCount:    -1,
}

// CreateUnknownPlace creates the default place if not exists.
func CreateUnknownPlace() {
	FirstOrCreatePlace(&UnknownPlace)
}

// FindPlace finds a matching place or returns nil.
func FindPlace(id string, label string) *Place {
	place := Place{}

	if label == "" {
		if err := Db().Where("id = ?", id).First(&place).Error; err != nil {
			log.Debugf("places: %s (find place id %s)", err.Error(), id)
			return nil
		} else {
			return &place
		}
	}

	if err := Db().Where("id = ? OR place_label = ?", id, label).First(&place).Error; err != nil {
		log.Debugf("places: %s (find place id %s, label %s)", err.Error(), id, txt.Quote(label))
		return nil
	} else {
		return &place
	}
}

// Find fetches entity values from the database the primary key.
func (m *Place) Find() error {
	if err := Db().First(m, "id = ?", m.ID).Error; err != nil {
		return err
	}

	return nil
}

// Create inserts a new row to the database.
func (m *Place) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreatePlace fetches an existing row, inserts a new row or nil in case of errors.
func FirstOrCreatePlace(m *Place) *Place {
	if m.ID == "" {
		log.Errorf("places: place id must not be empty (first or create)")
		return nil
	}

	if m.PlaceLabel == "" {
		log.Errorf("places: label must not be empty (first or create place id %s)", m.ID)
		return nil
	}

	result := Place{}

	if findErr := Db().Where("id = ? OR place_label = ?", m.ID, m.PlaceLabel).First(&result).Error; findErr == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := Db().Where("id = ? OR place_label = ?", m.ID, m.PlaceLabel).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("places: %s (create place id %s)", createErr, m.ID)
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

// City returns place City
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
