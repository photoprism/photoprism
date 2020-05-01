package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/ulule/deepcopier"
)

// Album represents a photo album
type Album struct {
	ID               uint   `gorm:"primary_key"`
	CoverUUID        string `gorm:"type:varbinary(36);"`
	AlbumUUID        string `gorm:"type:varbinary(36);unique_index;"`
	AlbumSlug        string `gorm:"type:varbinary(255);index;"`
	AlbumName        string `gorm:"type:varchar(255);"`
	AlbumDescription string `gorm:"type:text;"`
	AlbumNotes       string `gorm:"type:text;"`
	AlbumOrder       string `gorm:"type:varbinary(32);"`
	AlbumTemplate    string `gorm:"type:varbinary(255);"`
	AlbumFavorite    bool
	Links            []Link `gorm:"foreignkey:ShareUUID;association_foreignkey:AlbumUUID"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time `sql:"index"`
}

// BeforeCreate computes a random UUID when a new album is created in database
func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsPPID(m.AlbumUUID, 'a') {
		return nil
	}

	return scope.SetColumn("AlbumUUID", rnd.PPID('a'))
}

// NewAlbum creates a new album; default name is current month and year
func NewAlbum(name string) *Album {
	now := time.Now().UTC()

	result := &Album{
		AlbumUUID:  rnd.PPID('a'),
		AlbumOrder: SortOrderOldest,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result.SetName(name)

	return result
}

// SetName changes the album name.
func (m *Album) SetName(name string) {
	name = strings.TrimSpace(name)

	if name == "" {
		name = m.CreatedAt.Format("January 2006")
	}

	m.AlbumName = txt.Clip(name, txt.ClipDefault)

	if len(m.AlbumName) < txt.ClipSlug {
		m.AlbumSlug = slug.Make(m.AlbumName)
	} else {
		m.AlbumSlug = slug.Make(txt.Clip(m.AlbumName, txt.ClipSlug)) + "-" + m.AlbumUUID
	}
}

// Save updates the entity using form data and stores it in the database.
func (m *Album) Save(f form.Album) error {
	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	if f.AlbumName != "" {
		m.SetName(f.AlbumName)
	}

	return Db().Save(m).Error
}
