package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/ulule/deepcopier"
)

// Album represents a photo album
type Album struct {
	ID               uint   `gorm:"primary_key"`
	CoverUUID        string `gorm:"type:varbinary(36);"`
	AlbumUUID        string `gorm:"type:varbinary(36);unique_index;"`
	AlbumSlug        string `gorm:"type:varbinary(128);index;"`
	AlbumName        string `gorm:"type:varchar(128);"`
	AlbumDescription string `gorm:"type:text;"`
	AlbumNotes       string `gorm:"type:text;"`
	AlbumOrder       string `gorm:"type:varbinary(32);"`
	AlbumTemplate    string `gorm:"type:varbinary(256);"`
	AlbumFavorite    bool
	Links            []Link `gorm:"foreignkey:ShareUUID;association_foreignkey:AlbumUUID"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `sql:"index"`
}

// BeforeCreate computes a random UUID when a new album is created in database
func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("AlbumUUID", rnd.PPID('a')); err != nil {
		return err
	}

	return nil
}

// NewAlbum creates a new album; default name is current month and year
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

// Rename an existing album
func (m *Album) Rename(albumName string) {
	if albumName == "" {
		albumName = m.CreatedAt.Format("January 2006")
	}

	m.AlbumName = strings.TrimSpace(albumName)
	m.AlbumSlug = slug.Make(m.AlbumName)
}

// Save updates the entity using form data and stores it in the database.
func (m *Album) Save(f form.Album, db *gorm.DB) error {
	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	if f.AlbumName != "" {
		m.Rename(f.AlbumName)
	}

	return db.Save(m).Error
}
