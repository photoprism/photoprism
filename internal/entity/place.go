package entity

import (
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Place used to associate photos to places
type Place struct {
	ID          string    `gorm:"type:varbinary(16);primary_key;auto_increment:false;" json:"PlaceID" yaml:"PlaceID"`
	LocLabel    string    `gorm:"type:varbinary(768);unique_index;" json:"Label" yaml:"Label"`
	LocCity     string    `gorm:"type:varchar(255);" json:"City" yaml:"City,omitempty"`
	LocState    string    `gorm:"type:varchar(255);" json:"State" yaml:"State,omitempty"`
	LocCountry  string    `gorm:"type:varbinary(2);" json:"Country" yaml:"Country,omitempty"`
	LocKeywords string    `gorm:"type:varchar(255);" json:"Keywords" yaml:"Keywords,omitempty"`
	LocNotes    string    `gorm:"type:text;" json:"Notes" yaml:"Notes,omitempty"`
	LocFavorite bool      `json:"Favorite" yaml:"Favorite,omitempty"`
	PhotoCount  int       `json:"PhotoCount" yaml:"-"`
	CreatedAt   time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt   time.Time `json:"UpdatedAt" yaml:"-"`
	New         bool      `gorm:"-" json:"-" yaml:"-"`
}

// UnknownPlace is PhotoPrism's default place.
var UnknownPlace = Place{
	ID:          "zz",
	LocLabel:    "Unknown",
	LocCity:     "Unknown",
	LocState:    "Unknown",
	LocCountry:  "zz",
	LocKeywords: "",
	LocNotes:    "",
	LocFavorite: false,
	PhotoCount:  -1,
}

// CreateUnknownPlace creates the default place if not exists.
func CreateUnknownPlace() {
	FirstOrCreatePlace(&UnknownPlace)
}

// AfterCreate sets the New column used for database callback
func (m *Place) AfterCreate(scope *gorm.Scope) error {
	m.New = true
	return nil
}

// FindPlace finds a matching place or returns nil.
func FindPlace(id string, label string) *Place {
	place := &Place{}

	if label == "" {
		if err := Db().First(place, "id = ?", id).Error; err != nil {
			log.Debugf("place: %s for id %s", err.Error(), id)
			return nil
		}
	} else if err := Db().First(place, "id = ? OR loc_label = ?", id, label).Error; err != nil {
		log.Debugf("place: %s for id %s / label %s", err.Error(), id, txt.Quote(label))
		return nil
	}

	return place
}

// Find updates the entity with values from the database.
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

// FirstOrCreatePlace returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreatePlace(m *Place) *Place {
	if m.ID == "" {
		log.Errorf("place: id must not be empty")
		return nil
	}

	if m.LocLabel == "" {
		log.Errorf("place: label must not be empty (id %s)", m.ID)
		return nil
	}

	result := Place{}

	if err := Db().Where("id = ? OR loc_label = ?", m.ID, m.LocLabel).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("place: %s", err)
		return nil
	}

	return m
}

// Unknown returns true if this is an unknown place
func (m Place) Unknown() bool {
	return m.ID == "" || m.ID == UnknownPlace.ID
}

// Label returns place label
func (m Place) Label() string {
	return m.LocLabel
}

// City returns place City
func (m Place) City() string {
	return m.LocCity
}

// LongCity checks if the city name is more than 16 char.
func (m Place) LongCity() bool {
	return len(m.LocCity) > 16
}

// NoCity checks if the location has no city
func (m Place) NoCity() bool {
	return m.LocCity == ""
}

// CityContains checks if the location city contains the text string
func (m Place) CityContains(text string) bool {
	return strings.Contains(text, m.LocCity)
}

// State returns place State
func (m Place) State() string {
	return m.LocState
}

// CountryCode returns place CountryCode
func (m Place) CountryCode() string {
	return m.LocCountry
}

// CountryName returns place CountryName
func (m Place) CountryName() string {
	return maps.CountryNames[m.LocCountry]
}

// Notes returns place Notes
func (m Place) Notes() string {
	return m.LocNotes
}
