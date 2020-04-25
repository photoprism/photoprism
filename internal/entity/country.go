package entity

import (
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/internal/mutex"
)

// altCountryNames defines mapping between different names for the same countriy
var altCountryNames = map[string]string{
	"United States of America": "USA",
	"United States":            "USA",
	"":                         "Unknown",
}

// Country represents a country location, used for labeling photos.
type Country struct {
	ID                 string `gorm:"primary_key"`
	CountrySlug        string `gorm:"type:varbinary(128);unique_index;"`
	CountryName        string
	CountryDescription string `gorm:"type:text;"`
	CountryNotes       string `gorm:"type:text;"`
	CountryPhoto       *Photo
	CountryPhotoID     uint
	New                bool `gorm:"-"`
}

// UnknownCountry is defined here to use it as a default
var UnknownCountry = Country{
	ID:          "zz",
	CountryName: maps.CountryNames["zz"],
	CountrySlug: "zz",
}

// CreateUnknownCountry is used to initialize the database with the default country
func CreateUnknownCountry(db *gorm.DB) {
	UnknownCountry.FirstOrCreate(db)
}

// NewCountry creates a new country, with default country code if not provided
func NewCountry(countryCode string, countryName string) *Country {
	if countryCode == "" {
		return &UnknownCountry
	}

	if altName, ok := altCountryNames[countryName]; ok {
		countryName = altName
	}

	countrySlug := slug.MakeLang(countryName, "en")

	result := &Country{
		ID:          countryCode,
		CountryName: countryName,
		CountrySlug: countrySlug,
	}

	return result
}

// FirstOrCreate checks wether the country exist already in the database (using countryCode)
func (m *Country) FirstOrCreate(db *gorm.DB) *Country {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "id = ?", m.ID).Error; err != nil {
		log.Errorf("country: %s", err)
	}

	return m
}

// AfterCreate sets the New column used for database callback
func (m *Country) AfterCreate(scope *gorm.Scope) error {
	return scope.SetColumn("New", true)
}

// Code returns country code
func (m *Country) Code() string {
	return m.ID
}

// Name returns country name
func (m *Country) Name() string {
	return m.CountryName
}
