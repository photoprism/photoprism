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
	CoverUUID        string `gorm:"type:varchar(64);"`
	AlbumUUID        string `gorm:"unique_index;"`
	AlbumToken       string `gorm:"type:varchar(64);"`
	AlbumSlug        string `gorm:"index;"`
	AlbumName        string `gorm:"type:varchar(128);"`
	AlbumDescription string `gorm:"type:text;"`
	AlbumNotes       string `gorm:"type:text;"`
	AlbumViews       uint
	AlbumFavorite    bool
	AlbumPublic      bool
	AlbumLat         float64
	AlbumLong        float64
	AlbumRadius      float64
	AlbumOrder       string `gorm:"type:varchar(16);"`
	AlbumTemplate    string `gorm:"type:varchar(128);"`
}

func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("AlbumUUID", uuid.NewV4().String()); err != nil {
		return err
	}

	if err := scope.SetColumn("AlbumToken", util.RandomToken(4)); err != nil {
		return err
	}

	return nil
}

func NewAlbum(albumName string) *Album {
	albumName = strings.TrimSpace(albumName)

	if albumName == "" {
		albumName = time.Now().Format("January 2006")
	}

	albumSlug := slug.Make(albumName)

	result := &Album{
		AlbumSlug: albumSlug,
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
