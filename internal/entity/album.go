package entity

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/rnd"
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
	AlbumFavorite    bool
	AlbumOrder       string `gorm:"type:varbinary(32);"`
	ShareTemplate    string `gorm:"type:varbinary(256);"`
	SharePassword    string `gorm:"type:varbinary(256);"`
	ShareExpires     sql.NullTime
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `sql:"index"`
}

func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("AlbumUUID", rnd.PPID('a')); err != nil {
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
