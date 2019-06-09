package models

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// Photo album
type Album struct {
	Model
	AlbumUUID        string `gorm:"unique_index;"`
	AlbumSlug        string
	AlbumName        string
	AlbumDescription string `gorm:"type:text;"`
	AlbumNotes       string `gorm:"type:text;"`
	AlbumPhoto       *Photo
	AlbumPhotoID     uint
	Photos           []Photo `gorm:"many2many:album_photos;"`
}

func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("AlbumUUID", uuid.NewV4().String())
}
