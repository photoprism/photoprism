package entity

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/ulule/deepcopier"
)

const (
	AlbumDefault = "album"
	AlbumFolder  = "folder"
	AlbumMoment  = "moment"
	AlbumMonth   = "month"
	AlbumState   = "state"
)

type Albums []Album

// Album represents a photo album
type Album struct {
	ID               uint        `gorm:"primary_key" json:"ID" yaml:"-"`
	AlbumUID         string      `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	ParentUID        string      `gorm:"type:VARBINARY(42);default:'';" json:"ParentUID,omitempty" yaml:"ParentUID,omitempty"`
	AlbumSlug        string      `gorm:"type:VARBINARY(160);index;" json:"Slug" yaml:"Slug"`
	AlbumPath        string      `gorm:"type:VARBINARY(500);index;" json:"Path,omitempty" yaml:"Path,omitempty"`
	AlbumType        string      `gorm:"type:VARBINARY(8);default:'album';" json:"Type" yaml:"Type,omitempty"`
	AlbumTitle       string      `gorm:"type:VARCHAR(160);index;" json:"Title" yaml:"Title"`
	AlbumLocation    string      `gorm:"type:VARCHAR(160);" json:"Location" yaml:"Location,omitempty"`
	AlbumCategory    string      `gorm:"type:VARCHAR(100);index;" json:"Category" yaml:"Category,omitempty"`
	AlbumCaption     string      `gorm:"type:TEXT;" json:"Caption" yaml:"Caption,omitempty"`
	AlbumDescription string      `gorm:"type:TEXT;" json:"Description" yaml:"Description,omitempty"`
	AlbumNotes       string      `gorm:"type:TEXT;" json:"Notes" yaml:"Notes,omitempty"`
	AlbumFilter      string      `gorm:"type:VARBINARY(1024);" json:"Filter" yaml:"Filter,omitempty"`
	AlbumOrder       string      `gorm:"type:VARBINARY(32);" json:"Order" yaml:"Order,omitempty"`
	AlbumTemplate    string      `gorm:"type:VARBINARY(255);" json:"Template" yaml:"Template,omitempty"`
	AlbumState       string      `gorm:"type:VARCHAR(100);index;" json:"State" yaml:"State,omitempty"`
	AlbumCountry     string      `gorm:"type:VARBINARY(2);index:idx_albums_country_year_month;default:'zz';" json:"Country" yaml:"Country,omitempty"`
	AlbumYear        int         `gorm:"index:idx_albums_ymd;index:idx_albums_country_year_month;" json:"Year" yaml:"Year,omitempty"`
	AlbumMonth       int         `gorm:"index:idx_albums_ymd;index:idx_albums_country_year_month;" json:"Month" yaml:"Month,omitempty"`
	AlbumDay         int         `gorm:"index:idx_albums_ymd;" json:"Day" yaml:"Day,omitempty"`
	AlbumFavorite    bool        `json:"Favorite" yaml:"Favorite,omitempty"`
	AlbumPrivate     bool        `json:"Private" yaml:"Private,omitempty"`
	Thumb            string      `gorm:"type:VARBINARY(128);index;default:'';" json:"Thumb" yaml:"Thumb,omitempty"`
	ThumbSrc         string      `gorm:"type:VARBINARY(8);default:'';" json:"ThumbSrc,omitempty" yaml:"ThumbSrc,omitempty"`
	CreatedAt        time.Time   `json:"CreatedAt" yaml:"CreatedAt,omitempty"`
	UpdatedAt        time.Time   `json:"UpdatedAt" yaml:"UpdatedAt,omitempty"`
	DeletedAt        *time.Time  `sql:"index" json:"DeletedAt" yaml:"DeletedAt,omitempty"`
	Photos           PhotoAlbums `gorm:"foreignkey:AlbumUID;association_foreignkey:AlbumUID;" json:"-" yaml:"Photos,omitempty"`
}

// TableName returns the entity database table name.
func (Album) TableName() string {
	return "albums"
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
			a := NewAlbum(album, AlbumDefault)

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
	now := TimeStamp()

	if albumType == "" {
		albumType = AlbumDefault
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
func NewFolderAlbum(albumTitle, albumPath, albumFilter string) *Album {
	albumSlug := slug.Make(albumPath)

	if albumTitle == "" || albumSlug == "" || albumPath == "" || albumFilter == "" {
		return nil
	}

	now := TimeStamp()

	result := &Album{
		AlbumOrder:  SortOrderAdded,
		AlbumType:   AlbumFolder,
		AlbumTitle:  albumTitle,
		AlbumSlug:   albumSlug,
		AlbumPath:   albumPath,
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

	now := TimeStamp()

	result := &Album{
		AlbumOrder:  SortOrderOldest,
		AlbumType:   AlbumMoment,
		AlbumTitle:  albumTitle,
		AlbumSlug:   albumSlug,
		AlbumFilter: albumFilter,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return result
}

// NewStateAlbum creates a new moment.
func NewStateAlbum(albumTitle, albumSlug, albumFilter string) *Album {
	albumTitle = strings.TrimSpace(albumTitle)
	albumSlug = strings.TrimSpace(albumSlug)

	if albumTitle == "" || albumSlug == "" || albumFilter == "" {
		return nil
	}

	now := TimeStamp()

	result := &Album{
		AlbumOrder:  SortOrderNewest,
		AlbumType:   AlbumState,
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
	albumTitle = strings.TrimSpace(albumTitle)
	albumSlug = strings.TrimSpace(albumSlug)

	if albumTitle == "" || albumSlug == "" || year == 0 || month == 0 {
		return nil
	}

	f := form.PhotoSearch{
		Year:   strconv.Itoa(year),
		Month:  strconv.Itoa(month),
		Public: true,
	}

	now := TimeStamp()

	result := &Album{
		AlbumOrder:  SortOrderOldest,
		AlbumType:   AlbumMonth,
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

// FindMonthAlbum finds a matching month album or returns nil.
func FindMonthAlbum(year, month int) *Album {
	result := Album{}

	if err := UnscopedDb().Where("album_year = ? AND album_month = ? AND album_type = ?", year, month, AlbumMonth).First(&result).Error; err != nil {
		return nil
	}

	return &result
}

// FindAlbumBySlug finds a matching album or returns nil.
func FindAlbumBySlug(albumSlug, albumType string) *Album {
	result := Album{}

	if err := UnscopedDb().Where("album_slug = ? AND album_type = ?", albumSlug, albumType).First(&result).Error; err != nil {
		return nil
	}

	return &result
}

// FindAlbumByAttr finds an album by filters and slugs, or returns nil.
func FindAlbumByAttr(slugs, filters []string, albumType string) *Album {
	result := Album{}

	stmt := UnscopedDb()

	if albumType != "" {
		stmt = stmt.Where("album_type = ?", albumType)
	}

	stmt = stmt.Where("album_slug IN (?) OR album_filter IN (?)", slugs, filters)

	if err := stmt.First(&result).Error; err != nil {
		return nil
	}

	return &result
}

// FindFolderAlbum finds a matching folder album or returns nil.
func FindFolderAlbum(albumPath string) *Album {
	albumPath = strings.Trim(albumPath, string(os.PathSeparator))
	albumSlug := slug.Make(albumPath)

	if albumSlug == "" {
		return nil
	}

	result := Album{}

	stmt := UnscopedDb().Where("album_type = ?", AlbumFolder)
	stmt = stmt.Where("album_slug = ? OR album_path = ?", albumSlug, albumPath)

	if err := stmt.First(&result).Error; err != nil {
		return nil
	}

	return &result
}

// Find returns an entity from the database.
func (m *Album) Find() error {
	if rnd.IsPPID(m.AlbumUID, 'a') {
		if err := UnscopedDb().First(m, "album_uid = ?", m.AlbumUID).Error; err != nil {
			return err
		}
	}

	if m.AlbumType == "" {
		return fmt.Errorf("album type missing")
	}

	if m.AlbumSlug == "" {
		return fmt.Errorf("album slug missing")
	}

	stmt := UnscopedDb().Where("album_type = ?", m.AlbumType)

	if m.AlbumType != AlbumDefault && m.AlbumFilter != "" {
		stmt = stmt.Where("album_slug = ? OR album_filter = ?", m.AlbumSlug, m.AlbumFilter)
	} else {
		stmt = stmt.Where("album_slug = ?", m.AlbumSlug)
	}

	if err := stmt.First(m).Error; err != nil {
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

// IsMoment tests if the album is of type moment.
func (m *Album) IsMoment() bool {
	return m.AlbumType == AlbumMoment
}

// IsState tests if the album is of type state.
func (m *Album) IsState() bool {
	return m.AlbumType == AlbumState
}

// IsDefault tests if the album is a regular album.
func (m *Album) IsDefault() bool {
	return m.AlbumType == AlbumDefault
}

// SetTitle changes the album name.
func (m *Album) SetTitle(title string) {
	title = strings.TrimSpace(title)

	if title == "" {
		title = m.CreatedAt.Format("January 2006")
	}

	m.AlbumTitle = txt.Clip(title, txt.ClipDefault)

	if m.AlbumType == AlbumDefault || m.AlbumSlug == "" {
		if len(m.AlbumTitle) < txt.ClipSlug {
			m.AlbumSlug = txt.Slug(m.AlbumTitle)
		} else {
			m.AlbumSlug = txt.Slug(m.AlbumTitle) + "-" + m.AlbumUID
		}
	}

	if m.AlbumSlug == "" {
		m.AlbumSlug = "-"
	}
}

// UpdateSlug updates title and slug of generated albums if needed.
func (m *Album) UpdateSlug(title, slug string) error {
	title = strings.TrimSpace(title)
	slug = strings.TrimSpace(slug)

	if title == "" || slug == "" {
		return nil
	}

	changed := false

	if m.AlbumSlug != slug {
		m.AlbumSlug = slug
		changed = true
	}

	if !changed {
		return nil
	}

	m.AlbumTitle = title

	return m.Updates(Values{"album_title": m.AlbumTitle, "album_slug": m.AlbumSlug})
}

// UpdateState updates the album location.
func (m *Album) UpdateState(title, slug, stateName, countryCode string) error {
	if title == "" || slug == "" || stateName == "" || countryCode == "" {
		return nil
	}

	changed := false
	countryName := maps.CountryName(countryCode)

	if m.AlbumCountry != countryCode {
		m.AlbumCountry = countryCode
		changed = true
	}

	if changed || m.AlbumLocation == "" {
		m.AlbumLocation = countryName
		changed = true
	}

	if m.AlbumState != stateName {
		m.AlbumState = stateName
		changed = true
	}

	if m.AlbumSlug != slug {
		m.AlbumSlug = slug
		changed = true
	}

	if !changed {
		return nil
	}

	m.AlbumTitle = title

	return m.Updates(Values{"album_title": m.AlbumTitle, "album_slug": m.AlbumSlug, "album_location": m.AlbumLocation, "album_country": m.AlbumCountry, "album_state": m.AlbumState})
}

// SaveForm updates the entity using form data and stores it in the database.
func (m *Album) SaveForm(f form.Album) error {
	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	if f.AlbumCategory != "" {
		m.AlbumCategory = txt.Clip(txt.Title(f.AlbumCategory), txt.ClipCategory)
	}

	if f.AlbumTitle != "" {
		m.SetTitle(f.AlbumTitle)
	}

	return Db().Save(m).Error
}

// Update sets a new value for a database column.
func (m *Album) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Updates multiple columns in the database.
func (m *Album) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// UpdateFolder updates the path, filter and slug for a folder album.
func (m *Album) UpdateFolder(albumPath, albumFilter string) error {
	albumPath = strings.Trim(albumPath, string(os.PathSeparator))
	albumSlug := slug.Make(albumPath)

	if albumSlug == "" {
		return nil
	}

	if err := UnscopedDb().Model(m).UpdateColumns(map[string]interface{}{
		"AlbumPath":   albumPath,
		"AlbumFilter": albumFilter,
		"AlbumSlug":   albumSlug,
	}).Error; err != nil {
		return err
	} else if err := UnscopedDb().Exec("UPDATE albums SET album_path = NULL WHERE album_path = ? AND id <> ?", albumPath, m.ID).Error; err != nil {
		return err
	}

	return nil
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

	m.PublishCountChange(1)

	return nil
}

// PublishCountChange publishes an event with the added or removed number of albums.
func (m *Album) PublishCountChange(n int) {
	data := event.Data{"count": n}

	switch m.AlbumType {
	case AlbumDefault:
		event.Publish("count.albums", data)
	case AlbumMoment:
		event.Publish("count.moments", data)
	case AlbumMonth:
		event.Publish("count.months", data)
	case AlbumFolder:
		event.Publish("count.folders", data)
	}
}

// Delete marks the entity as deleted in the database.
func (m *Album) Delete() error {
	if m.Deleted() {
		return nil
	}

	if err := Db().Delete(m).Error; err != nil {
		return err
	}

	m.PublishCountChange(-1)

	return DeleteShareLinks(m.AlbumUID)
}

// DeletePermanently permanently removes an album from the index.
func (m *Album) DeletePermanently() error {
	wasDeleted := m.Deleted()

	if err := UnscopedDb().Delete(m).Error; err != nil {
		return err
	}

	if !wasDeleted {
		m.PublishCountChange(-1)
	}

	return DeleteShareLinks(m.AlbumUID)
}

// Deleted tests if the entity is deleted.
func (m *Album) Deleted() bool {
	return m.DeletedAt != nil
}

// Restore restores the entity in the database.
func (m *Album) Restore() error {
	if !m.Deleted() {
		return nil
	}

	if err := UnscopedDb().Model(m).Update("DeletedAt", nil).Error; err != nil {
		return err
	}

	m.DeletedAt = nil

	m.PublishCountChange(1)

	return nil
}

// Title returns the album title.
func (m *Album) Title() string {
	return m.AlbumTitle
}

// ZipName returns the zip download filename.
func (m *Album) ZipName() string {
	s := slug.Make(m.AlbumTitle)

	if len(s) < 2 {
		s = fmt.Sprintf("photoprism-album-%s", m.AlbumUID)
	}

	return fmt.Sprintf("%s.zip", s)
}

// AddPhotos adds photos to an existing album.
func (m *Album) AddPhotos(UIDs []string) (added PhotoAlbums) {
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
func (m *Album) RemovePhotos(UIDs []string) (removed PhotoAlbums) {
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

// Links returns all share links for this entity.
func (m *Album) Links() Links {
	return FindLinks("", m.AlbumUID)
}
