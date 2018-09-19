package models

import (
	"github.com/jinzhu/gorm"
)

type Country struct {
	gorm.Model
	CountryCode        string
	CountryName        string
	CountryDescription string `gorm:"type:text;"`
	CountryNotes       string `gorm:"type:text;"`
	CountryPhoto       *Photo
	CountryPhotoID     uint
}
