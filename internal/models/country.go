package models

import (
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

type Country struct {
	ID                 string `gorm:"primary_key"`
	CountrySlug        string
	CountryName        string
	CountryDescription string `gorm:"type:text;"`
	CountryNotes       string `gorm:"type:text;"`
	CountryPhoto       *Photo
	CountryPhotoID     uint
}

// Create a new country
func NewCountry(countryCode string, countryName string) *Country {
	if countryCode == "" {
		countryCode = "zz"
	}

	if countryName == "" {
		countryName = "Unknown"
	}

	countrySlug := slug.MakeLang(countryName, "en")

	result := &Country{
		ID:          countryCode,
		CountryName: countryName,
		CountrySlug: countrySlug,
	}

	return result
}

func (m *Country) FirstOrCreate(db *gorm.DB) *Country {
	db.FirstOrCreate(m, "id = ?", m.ID)

	return m
}
