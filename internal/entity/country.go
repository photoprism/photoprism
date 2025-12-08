package entity

import (
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/maps"
	"github.com/photoprism/photoprism/pkg/txt"
)

// altCountryNames captures alternate spellings mapped to canonical country names.
var altCountryNames = map[string]string{
	"us":                       "United States",
	"usa":                      "United States",
	"US":                       "United States",
	"USA":                      "United States",
	"United States of America": "United States",
	"":                         "Unknown",
}

// Countries represents a list of countries.
type Countries []Country

// Country represents a country entry used for labeling photos and locations.
type Country struct {
	ID                 string `gorm:"type:VARBINARY(2);primary_key" json:"ID" yaml:"ID"`
	CountrySlug        string `gorm:"type:VARBINARY(160);unique_index;" json:"Slug" yaml:"-"`
	CountryName        string `gorm:"type:VARCHAR(160);" json:"Name" yaml:"Name,omitempty"`
	CountryDescription string `gorm:"type:VARCHAR(2048);" json:"Description,omitempty" yaml:"Description,omitempty"`
	CountryNotes       string `gorm:"type:VARCHAR(1024);" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CountryPhoto       *Photo `json:"-" yaml:"-"`
	CountryPhotoID     uint   `json:"-" yaml:"-"`
	New                bool   `gorm:"-" json:"-" yaml:"-"`
}

// TableName returns the entity table name.
func (Country) TableName() string {
	return "countries"
}

// UnknownCountry is the fallback entry used when no country code is known.
var UnknownCountry = Country{
	ID:          UnknownID,
	CountryName: maps.CountryNames[UnknownID],
	CountrySlug: UnknownID,
}

// CreateUnknownCountry ensures the fallback country exists in the database.
func CreateUnknownCountry() {
	UnknownCountry = *FirstOrCreateCountry(&UnknownCountry)
}

// NewCountry creates a new Country entity, normalizing name and slug.
func NewCountry(countryCode string, countryName string) *Country {
	if countryCode == "" {
		return &UnknownCountry
	}

	if altName, ok := altCountryNames[countryName]; ok {
		countryName = altName
	}

	result := &Country{
		ID:          countryCode,
		CountryName: txt.Clip(countryName, txt.ClipName),
		CountrySlug: txt.Slug(countryName),
	}

	return result
}

// Create inserts a new row to the database.
func (m *Country) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreateCountry returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateCountry(m *Country) *Country {
	if cacheData, ok := countryCache.Get(m.ID); ok {
		log.Tracef("country: cache hit for %s", m.ID)

		return cacheData.(*Country)
	}

	result := Country{}

	if findErr := Db().Where("id = ?", m.ID).First(&result).Error; findErr == nil {
		countryCache.SetDefault(m.ID, &result)
		return &result
	} else if createErr := m.Create(); createErr == nil {
		if !m.Unknown() {
			event.EntitiesCreated("countries", []*Country{m})

			event.Publish("count.countries", event.Data{
				"count": 1,
			})
		}
		countryCache.SetDefault(m.ID, m)
		return m
	} else if err := Db().Where("id = ?", m.ID).First(&result).Error; err == nil {
		countryCache.SetDefault(m.ID, &result)
		return &result
	} else {
		log.Errorf("country: %s (find or create %s)", createErr, m.ID)
	}

	return &UnknownCountry
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
