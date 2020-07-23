package entity

import (
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/maps"
)

// altCountryNames defines mapping between different names for the same countriy
var altCountryNames = map[string]string{
	"United States of America": "USA",
	"United States":            "USA",
	"":                         "Unknown",
}

// Country represents a country location, used for labeling photos.
type Country struct {
	ID                 string `gorm:"type:varbinary(2);primary_key" json:"ID" yaml:"ID"`
	CountrySlug        string `gorm:"type:varbinary(255);unique_index;" json:"Slug" yaml:"-"`
	CountryName        string `json:"Name" yaml:"Name,omitempty"`
	CountryDescription string `gorm:"type:text;" json:"Description,omitempty" yaml:"Description,omitempty"`
	CountryNotes       string `gorm:"type:text;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CountryPhoto       *Photo `json:"-" yaml:"-"`
	CountryPhotoID     uint   `json:"-" yaml:"-"`
	New                bool   `gorm:"-" json:"-" yaml:"-"`
}

// UnknownCountry is defined here to use it as a default
var UnknownCountry = Country{
	ID:          "zz",
	CountryName: maps.CountryNames["zz"],
	CountrySlug: "zz",
}

// CreateUnknownCountry is used to initialize the database with the default country
func CreateUnknownCountry() {
	FirstOrCreateCountry(&UnknownCountry)
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

// Create inserts a new row to the database.
func (m *Country) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreateCountry returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateCountry(m *Country) *Country {
	result := Country{}

	if findErr := Db().Where("id = ?", m.ID).First(&result).Error; findErr == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		if !m.Unknown() {
			event.EntitiesCreated("countries", []*Country{m})

			event.Publish("count.countries", event.Data{
				"count": 1,
			})
		}

		return m
	} else if err := Db().Where("id = ?", m.ID).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("country: %s (first or create %s)", createErr, m.ID)
	}

	return nil
}

// AfterCreate sets the New column used for database callback
func (m *Country) AfterCreate(scope *gorm.Scope) error {
	m.New = true
	return nil
}

// Code returns country code
func (m *Country) Code() string {
	return m.ID
}

// Name returns country name
func (m *Country) Name() string {
	return m.CountryName
}

// Unknown returns true if the country is not a known country.
func (m *Country) Unknown() bool {
	return m.ID == "" || m.ID == UnknownCountry.ID
}
