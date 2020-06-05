package entity

import (
	"fmt"
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

type Albums []Album

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

// AddPhotoToAlbums adds a photo UID to multiple albums and automatically creates them if needed.
func AddPhotoToAlbums(photo string, albums []string) (err error) {
	if photo == "" || len(albums) == 0 {
		// Do nothing.
		return nil
	}

	if !rnd.IsPPID(photo, 'p') {
		return fmt.Errorf("album: invalid photo uid %s", photo)
	}

	for _, album := range albums {
		var aUID string

		if album == "" {
			log.Debugf("album: empty album identifier while adding photo %s", photo)
			continue
		}

		if rnd.IsPPID(album, 'a') {
			aUID = album
		} else {
			a := NewAlbum(album, TypeAlbum)

			if err = a.Find(); err == nil {
				aUID = a.AlbumUID
			} else if err = a.Create(); err == nil {
				aUID = a.AlbumUID
			} else {
				log.Errorf("album: %s (add photo %s to albums)", err.Error(), photo)
			}
		}

		if aUID != "" {
			entry := PhotoAlbum{AlbumUID: aUID, PhotoUID: photo, Hidden: false}

			if err = entry.Save(); err != nil {
				log.Errorf("album: %s (add photo %s to albums)", err.Error(), photo)
			}
		}
	}

	return err
}

// NewAlbum creates a new album; default name is current month and year
func NewAlbum(albumTitle, albumType string) *Album {
	now := time.Now().UTC()

	if albumType == "" {
		albumType = TypeAlbum
	}

	result := &Album{
		AlbumOrder: SortOrderOldest,
		AlbumType:  albumType,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result.SetTitle(albumTitle)

	return result
}

// NewFolderAlbum creates a new folder album.
func NewFolderAlbum(albumTitle, albumSlug, albumFilter string) *Album {
	if albumTitle == "" || albumSlug == "" || albumFilter == "" {
		return nil
	}

	now := time.Now().UTC()

	result := &Album{
		AlbumOrder:  SortOrderOldest,
		AlbumType:   TypeFolder,
		AlbumTitle:  albumTitle,
		AlbumSlug:   albumSlug,
		AlbumFilter: albumFilter,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return result
}

// NewMomentsAlbum creates a new moment.
func NewMomentsAlbum(albumTitle, albumSlug, albumFilter string) *Album {
	if albumTitle == "" || albumSlug == "" || albumFilter == "" {
		return nil
	}

	now := time.Now().UTC()

	result := &Album{
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

// NewMonthAlbum creates a new month album.
func NewMonthAlbum(albumTitle, albumSlug string, year, month int) *Album {
	if albumTitle == "" || albumSlug == "" || year == 0 || month == 0 {
		return nil
	}

	f := form.PhotoSearch{
		Year:   year,
		Month:  month,
		Public: true,
	}

	now := time.Now().UTC()

	result := &Album{
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

// FindAlbumBySlug finds a matching album or returns nil.
func FindAlbumBySlug(slug, albumType string) *Album {
	result := Album{}

	if err := UnscopedDb().Where("album_slug = ? AND album_type = ?", slug, albumType).First(&result).Error; err != nil {
		return nil
	}

	return &result
}

// Find updates the entity with values from the database.
func (m *Album) Find() error {
	if rnd.IsPPID(m.AlbumUID, 'a') {
		log.Debugf("IS PPID: %s", m.AlbumUID)
		if err := UnscopedDb().First(m, "album_uid = ?", m.AlbumUID).Error; err != nil {
			return err
		}
	}

	if err := UnscopedDb().First(m, "album_slug = ? AND album_type = ?", m.AlbumSlug, m.AlbumType).Error; err != nil {
		return err
	}

	return nil
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.AlbumUID, 'a') {
		return nil
	}

	return scope.SetColumn("AlbumUID", rnd.PPID('a'))
}

// String returns the id or name as string.
func (m *Album) String() string {
	if m.AlbumSlug != "" {
		return m.AlbumSlug
	}

	if m.AlbumTitle != "" {
		return txt.Quote(m.AlbumTitle)
	}

	if m.AlbumUID != "" {
		return m.AlbumUID
	}

	return "[unknown album]"
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

	if m.AlbumType == TypeAlbum {
		if len(m.AlbumTitle) < txt.ClipSlug {
			m.AlbumSlug = slug.Make(m.AlbumTitle)
		} else {
			m.AlbumSlug = slug.Make(txt.Clip(m.AlbumTitle, txt.ClipSlug)) + "-" + m.AlbumUID
		}
	}
}

// Saves the entity using form data and stores it in the database.
func (m *Album) SaveForm(f form.Album) error {
	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	if f.AlbumCategory != "" {
		m.AlbumCategory = txt.Title(txt.Clip(f.AlbumCategory, txt.ClipKeyword))
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

// Returns the album title.
func (m *Album) Title() string {
	return m.AlbumTitle
}

// AddPhotos adds photos to an existing album.
func (m *Album) AddPhotos(UIDs []string) (added []PhotoAlbum) {
	for _, uid := range UIDs {
		entry := PhotoAlbum{AlbumUID: m.AlbumUID, PhotoUID: uid, Hidden: false}

		if err := entry.Save(); err != nil {
			log.Errorf("album: %s (add to album %s)", err.Error(), m)
		} else {
			added = append(added, entry)
		}
	}

	return added
}

// RemovePhotos removes photos from an album.
func (m *Album) RemovePhotos(UIDs []string) (removed []PhotoAlbum) {
	for _, uid := range UIDs {
		entry := PhotoAlbum{AlbumUID: m.AlbumUID, PhotoUID: uid, Hidden: true}

		if err := entry.Save(); err != nil {
			log.Errorf("album: %s (remove from album %s)", err.Error(), m)
		} else {
			removed = append(removed, entry)
		}
	}

	return removed
}
