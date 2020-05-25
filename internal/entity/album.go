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
	ID               uint       `gorm:"primary_key" json:"ID" yaml:"-"`
	AlbumUID         string     `gorm:"type:varbinary(36);unique_index;" json:"UID" yaml:"UID"`
	CoverUID         string     `gorm:"type:varbinary(36);" json:"CoverUID" yaml:"CoverUID,omitempty"`
	ParentUID        string     `gorm:"type:varbinary(36);index;" json:"ParentUID" yaml:"ParentUID,omitempty"`
	FolderUID        string     `gorm:"type:varbinary(36);index;" json:"FolderUID" yaml:"FolderUID,omitempty"`
	AlbumSlug        string     `gorm:"type:varbinary(255);index;" json:"Slug" yaml:"Slug"`
	AlbumName        string     `gorm:"type:varchar(255);" json:"Name" yaml:"Name"`
	AlbumType        string     `gorm:"type:varbinary(8);default:'';" json:"Type" yaml:"Type"`
	AlbumFilter      string     `gorm:"type:varchar(1024);" json:"Filter" yaml:"Filter,omitempty"`
	AlbumDescription string     `gorm:"type:text;" json:"Description" yaml:"Description,omitempty"`
	AlbumNotes       string     `gorm:"type:text;" json:"Notes" yaml:"Notes,omitempty"`
	AlbumOrder       string     `gorm:"type:varbinary(32);" json:"Order" yaml:"Order,omitempty"`
	AlbumTemplate    string     `gorm:"type:varbinary(255);" json:"Template" yaml:"Template,omitempty"`
	AlbumFavorite    bool       `json:"Favorite" yaml:"Favorite,omitempty"`
	Links            []Link     `gorm:"foreignkey:share_uid;association_foreignkey:album_uid" json:"Links" yaml:"-"`
	CreatedAt        time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt        time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt        *time.Time `sql:"index" json:"-" yaml:"-"`
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsPPID(m.AlbumUID, 'a') {
		return nil
	}

	return scope.SetColumn("AlbumUID", rnd.PPID('a'))
}

// NewAlbum creates a new album; default name is current month and year
func NewAlbum(albumName, albumType string) *Album {
	now := time.Now().UTC()

	result := &Album{
		AlbumUID:   rnd.PPID('a'),
		AlbumOrder: SortOrderOldest,
		AlbumType:  albumType,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result.SetName(albumName)

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
		m.AlbumSlug = slug.Make(txt.Clip(m.AlbumName, txt.ClipSlug)) + "-" + m.AlbumUID
	}
}

// Saves the entity using form data and stores it in the database.
func (m *Album) Save(f form.Album) error {
	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	if f.AlbumName != "" {
		m.SetName(f.AlbumName)
	}

	return Db().Save(m).Error
}
