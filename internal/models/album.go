package models

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/util"
	uuid "github.com/satori/go.uuid"
)

// Photo album
type Album struct {
	Model
	AlbumUUID        string `gorm:"unique_index;"`
	AlbumSlug        string `gorm:"index;"`
	AlbumSecret      string `gorm:"type:varchar(64);"`
	AlbumName        string `gorm:"type:varchar(128);"`
	AlbumDescription string `gorm:"type:text;"`
	AlbumNotes       string `gorm:"type:text;"`
	AlbumViews       uint
	AlbumPhoto       *Photo
	AlbumPhotoID     uint
	AlbumFavorite    bool
	AlbumPublic      bool
}

func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("AlbumUUID", uuid.NewV4().String())
}

func NewAlbum(albumName string) *Album {
	albumName = strings.TrimSpace(albumName)

	if albumName == "" {
		albumName = time.Now().Format("January 2006")
	}

	albumSlug := slug.Make(albumName)
	albumUUID := uuid.NewV4().String()
	albumSecret := util.RandomToken(10)

	result := &Album{
		AlbumUUID: albumUUID,
		AlbumSlug: albumSlug,
		AlbumSecret: albumSecret,
		AlbumName: albumName,
	}

	return result
}

func (m *Album) Rename(albumName string) {
	if albumName == "" {
		albumName = m.CreatedAt.Format("January 2006")
	}

	m.AlbumName = strings.TrimSpace(albumName)
	m.AlbumSlug = slug.Make(m.AlbumName)
}
