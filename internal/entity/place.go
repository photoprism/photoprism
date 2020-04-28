package entity

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/internal/mutex"
)

// Place used to associate photos to places
type Place struct {
	ID          string `gorm:"type:varbinary(16);primary_key;auto_increment:false;"`
	LocLabel    string `gorm:"type:varbinary(768);unique_index;"`
	LocCity     string `gorm:"type:varchar(255);"`
	LocState    string `gorm:"type:varchar(255);"`
	LocCountry  string `gorm:"type:varbinary(2);"`
	LocKeywords string `gorm:"type:varchar(255);"`
	LocNotes    string `gorm:"type:text;"`
	LocFavorite bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	New         bool `gorm:"-"`
}

// UnknownPlace is defined here to use it as a default
var UnknownPlace = Place{
	ID:          "zz",
	LocLabel:    "Unknown",
	LocCity:     "Unknown",
	LocState:    "Unknown",
	LocCountry:  "zz",
	LocKeywords: "",
	LocNotes:    "",
	LocFavorite: false,
}

// CreateUnknownPlace initializes default place in the database
func CreateUnknownPlace(db *gorm.DB) {
	UnknownPlace.FirstOrCreate(db)
}

// AfterCreate sets the New column used for database callback
func (m *Place) AfterCreate(scope *gorm.Scope) error {
	return scope.SetColumn("New", true)
}

// FindPlaceByLabel returns a place from an id or a label
func FindPlaceByLabel(id string, label string, db *gorm.DB) *Place {
	place := &Place{}

	if err := db.First(place, "id = ? OR loc_label = ?", id, label).Error; err != nil {
		log.Debugf("place: %s for id %s or label \"%s\"", err.Error(), id, label)
		return nil
	}

	return place
}

// Find returns db record of place
func (m *Place) Find(db *gorm.DB) error {
	if err := db.First(m, "id = ?", m.ID).Error; err != nil {
		return err
	}

	return nil
}

// FirstOrCreate checks wether the place already exists in the database
func (m *Place) FirstOrCreate(db *gorm.DB) *Place {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "id = ? OR loc_label = ?", m.ID, m.LocLabel).Error; err != nil {
		log.Debugf("place: %s for token %s or label \"%s\"", err.Error(), m.ID, m.LocLabel)
	}

	return m
}

// Unknown returns true if this is an unknown place
func (m Place) Unknown() bool {
	return m.ID == UnknownPlace.ID
}

// Label returns place label
func (m Place) Label() string {
	return m.LocLabel
}

// City returns place City
func (m Place) City() string {
	return m.LocCity
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
