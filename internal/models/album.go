package models

import (
	"strings"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// Photo album
type Album struct {
	Model
	AlbumUUID        string `gorm:"unique_index;"`
	AlbumSlug        string `gorm:"index;"`
	AlbumName        string
	AlbumDescription string `gorm:"type:text;"`
	AlbumNotes       string `gorm:"type:text;"`
	AlbumViews       uint
	AlbumPhoto       *Photo
	AlbumPhotoID     uint
	AlbumFavorite    bool
	Photos           []Photo `gorm:"many2many:album_photos;"`
}

func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("AlbumUUID", uuid.NewV4().String())
}

func NewAlbum(albumName string) *Album {
	albumName = strings.TrimSpace(albumName)

	if albumName == "" {
		albumName = "New Album"
	}

	albumSlug := slug.Make(albumName)
	albumUUID := uuid.NewV4().String()

	result := &Album{
		AlbumUUID: albumUUID,
		AlbumSlug: albumSlug,
		AlbumName: albumName,
	}

	return result
}
