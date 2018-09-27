package models

import (
	"github.com/jinzhu/gorm"
)

type Album struct {
	gorm.Model
	AlbumSlug        string
	AlbumName        string
	AlbumDescription string `gorm:"type:text;"`
	AlbumNotes       string `gorm:"type:text;"`
	AlbumPhoto       *Photo
	AlbumPhotoID     uint
	Photos           []Photo `gorm:"many2many:album_photos;"`
}
