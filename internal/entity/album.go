package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

// Photo album
type Album struct {
	ID               uint   `gorm:"primary_key"`
	CoverUUID        string `gorm:"type:varbinary(36);"`
	AlbumUUID        string `gorm:"type:varbinary(36);unique_index;"`
	AlbumSlug        string `gorm:"type:varbinary(128);index;"`
	AlbumName        string `gorm:"type:varchar(128);"`
	AlbumDescription string `gorm:"type:text;"`
	AlbumNotes       string `gorm:"type:text;"`
	AlbumViews       uint
	AlbumFavorite    bool
	AlbumPublic      bool
	AlbumLat         float64
	AlbumLng         float64
	AlbumRadius      float64
	AlbumOrder       string `gorm:"type:varchar(16);"`
	AlbumTemplate    string `gorm:"type:varchar(128);"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `sql:"index"`
}

func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("AlbumUUID", ID('a')); err != nil {
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
