package entity

import (
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
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
	FolderUID        string     `gorm:"type:varbinary(36);index;" json:"FolderUID" yaml:"FolderUID,omitempty"`
	AlbumSlug        string     `gorm:"type:varbinary(255);index;" json:"Slug" yaml:"Slug"`
	AlbumType        string     `gorm:"type:varbinary(8);default:'album';" json:"Type" yaml:"Type,omitempty"`
	AlbumTitle       string     `gorm:"type:varchar(255);" json:"Title" yaml:"Title"`
	AlbumCategory    string     `gorm:"type:varchar(255);index;" json:"Category" yaml:"Category,omitempty"`
	AlbumCaption     string     `gorm:"type:text;" json:"Caption" yaml:"Caption,omitempty"`
	AlbumDescription string     `gorm:"type:text;" json:"Description" yaml:"Description,omitempty"`
	AlbumNotes       string     `gorm:"type:text;" json:"Notes" yaml:"Notes,omitempty"`
	AlbumFilter      string     `gorm:"type:varbinary(1024);" json:"Filter" yaml:"Filter,omitempty"`
	AlbumOrder       string     `gorm:"type:varbinary(32);" json:"Order" yaml:"Order,omitempty"`
	AlbumTemplate    string     `gorm:"type:varbinary(255);" json:"Template" yaml:"Template,omitempty"`
	AlbumCountry     string     `gorm:"type:varbinary(2);index:idx_albums_country_year_month;default:'zz'" json:"Country" yaml:"Country,omitempty"`
	AlbumYear        int        `gorm:"index:idx_albums_country_year_month;" json:"Year" yaml:"Year,omitempty"`
	AlbumMonth       int        `gorm:"index:idx_albums_country_year_month;" json:"Month" yaml:"Month,omitempty"`
	AlbumFavorite    bool       `json:"Favorite" yaml:"Favorite,omitempty"`
	AlbumPrivate     bool       `json:"Private" yaml:"Private,omitempty"`
	Links            []Link     `gorm:"foreignkey:share_uid;association_foreignkey:album_uid" json:"Links" yaml:"-"`
	CreatedAt        time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt        time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt        *time.Time `sql:"index" json:"-" yaml:"-"`
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.AlbumUID, 'a') {
		return nil
	}

	return scope.SetColumn("AlbumUID", rnd.PPID('a'))
}

// NewAlbum creates a new album; default name is current month and year
func NewAlbum(albumTitle, albumType string) *Album {
	now := time.Now().UTC()

	if albumType == "" {
		albumType = TypeAlbum
	}

	result := &Album{
		AlbumUID:   rnd.PPID('a'),
		AlbumOrder: SortOrderOldest,
		AlbumType:  albumType,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result.SetTitle(albumTitle)

	return result
}

// NewMoment creates a new moment.
func NewMoment(albumTitle, albumSlug, albumFilter string) *Album {
	if albumTitle == "" || albumSlug == "" || albumFilter == "" {
		return nil
	}

	now := time.Now().UTC()

	result := &Album{
		AlbumUID:    rnd.PPID('a'),
		AlbumOrder:  SortOrderOldest,
		AlbumType:   TypeMoment,
		AlbumTitle:  albumTitle,
		AlbumSlug:   albumSlug,
		AlbumFilter: albumFilter,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return result
}

// NewMonth creates a new month album.
func NewMonth(albumTitle, albumSlug string, year, month int) *Album {
	if albumTitle == "" || albumSlug == "" || year == 0 || month == 0 {
		return nil
	}

	f := form.PhotoSearch{
		Year:  year,
		Month: month,
	}

	now := time.Now().UTC()

	result := &Album{
		AlbumUID:    rnd.PPID('a'),
		AlbumOrder:  SortOrderOldest,
		AlbumType:   TypeMonth,
		AlbumTitle:  albumTitle,
		AlbumSlug:   albumSlug,
		AlbumFilter: f.Serialize(),
		AlbumYear:   year,
		AlbumMonth:  month,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return result
}

// Checks if the album is of type moment.
func (m *Album) IsMoment() bool {
	return m.AlbumType == TypeMoment
}

// SetTitle changes the album name.
func (m *Album) SetTitle(title string) {
	title = strings.TrimSpace(title)

	if title == "" {
		title = m.CreatedAt.Format("January 2006")
	}

	m.AlbumTitle = txt.Clip(title, txt.ClipDefault)

	if len(m.AlbumTitle) < txt.ClipSlug {
		m.AlbumSlug = slug.Make(m.AlbumTitle)
	} else {
		m.AlbumSlug = slug.Make(txt.Clip(m.AlbumTitle, txt.ClipSlug)) + "-" + m.AlbumUID
	}
}

// Saves the entity using form data and stores it in the database.
func (m *Album) SaveForm(f form.Album) error {
	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	if f.AlbumTitle != "" {
		m.SetTitle(f.AlbumTitle)
	}

	return Db().Save(m).Error
}

// Updates a column in the database.
func (m *Album) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Save updates the existing or inserts a new row.
func (m *Album) Save() error {
	return Db().Save(m).Error
}

// Create inserts a new row to the database.
func (m *Album) Create() error {
	if err := Db().Create(m).Error; err != nil {
		return err
	}

	switch m.AlbumType {
	case TypeAlbum:
		event.Publish("count.albums", event.Data{"count": 1})
	case TypeMoment:
		event.Publish("count.moments", event.Data{"count": 1})
	case TypeMonth:
		event.Publish("count.months", event.Data{"count": 1})
	case TypeFolder:
		event.Publish("count.folders", event.Data{"count": 1})
	}
	return nil
}

// FindAlbum finds a matching album or returns nil.
func FindAlbum(slug string) *Album {
	result := Album{}

	if err := Db().Where("album_slug = ?", slug).First(&result).Error; err != nil {
		return nil
	}

	return &result
}
