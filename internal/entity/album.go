package entity

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/maps"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	AlbumDefault = "album"
	AlbumFolder  = "folder"
	AlbumMoment  = "moment"
	AlbumMonth   = "month"
	AlbumState   = "state"
	AlbumUID     = 'a'
)

type Albums []Album

// Album represents a photo album
type Album struct {
	ID               uint        `gorm:"primary_key" json:"ID" yaml:"-"`
	AlbumUID         string      `gorm:"type:VARBINARY(64);unique_index;" json:"UID" yaml:"UID"`
	ParentUID        string      `gorm:"type:VARBINARY(64);default:'';" json:"ParentUID,omitempty" yaml:"ParentUID,omitempty"`
	AlbumSlug        string      `gorm:"type:VARBINARY(160);index;" json:"Slug" yaml:"Slug"`
	AlbumPath        string      `gorm:"type:VARBINARY(500);index;" json:"Path,omitempty" yaml:"Path,omitempty"`
	AlbumType        string      `gorm:"type:VARBINARY(8);default:'album';" json:"Type" yaml:"Type,omitempty"`
	AlbumTitle       string      `gorm:"type:VARCHAR(160);index;" json:"Title" yaml:"Title"`
	AlbumLocation    string      `gorm:"type:VARCHAR(160);" json:"Location" yaml:"Location,omitempty"`
	AlbumCategory    string      `gorm:"type:VARCHAR(100);index;" json:"Category" yaml:"Category,omitempty"`
	AlbumCaption     string      `gorm:"type:VARCHAR(1024);" json:"Caption" yaml:"Caption,omitempty"`
	AlbumDescription string      `gorm:"type:VARCHAR(2048);" json:"Description" yaml:"Description,omitempty"`
	AlbumNotes       string      `gorm:"type:VARCHAR(1024);" json:"Notes" yaml:"Notes,omitempty"`
	AlbumFilter      string      `gorm:"type:VARBINARY(2048);" json:"Filter" yaml:"Filter,omitempty"`
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

// AfterUpdate flushes the album cache.
func (m *Album) AfterUpdate(tx *gorm.DB) (err error) {
	FlushAlbumCache()
	return
}

// AfterDelete flushes the album cache.
func (m *Album) AfterDelete(tx *gorm.DB) (err error) {
	FlushAlbumCache()
	return
}

// TableName returns the entity table name.
func (Album) TableName() string {
	return "albums"
}

// AddPhotoToAlbums adds a photo UID to multiple albums and automatically creates them if needed.
func AddPhotoToAlbums(photo string, albums []string) (err error) {
	if photo == "" || len(albums) == 0 {
		// Do nothing.
		return nil
	}

	if !rnd.IsUID(photo, PhotoUID) {
		return fmt.Errorf("album: invalid photo uid %s", photo)
	}

	for _, album := range albums {
		var aUID string

		if album == "" {
			log.Debugf("album: empty album identifier while adding photo %s", photo)
			continue
		}

		if rnd.IsUID(album, AlbumUID) {
			aUID = album
		} else {
			a := NewAlbum(album, AlbumDefault)

			if found := a.Find(); found != nil {
				aUID = found.AlbumUID
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
	albumSlug := txt.Slug(albumPath)

	if albumTitle == "" || albumSlug == "" || albumPath == "" || albumFilter == "" {
		return nil
	}

	now := TimeStamp()

	result := &Album{
		AlbumOrder:  SortOrderAdded,
		AlbumType:   AlbumFolder,
		AlbumSlug:   txt.Clip(albumSlug, txt.ClipSlug),
		AlbumPath:   txt.Clip(albumPath, txt.ClipPath),
		AlbumFilter: albumFilter,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result.SetTitle(albumTitle)

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
		AlbumSlug:   txt.Clip(albumSlug, txt.ClipSlug),
		AlbumFilter: albumFilter,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result.SetTitle(albumTitle)

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
		AlbumSlug:   txt.Clip(albumSlug, txt.ClipSlug),
		AlbumFilter: albumFilter,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result.SetTitle(albumTitle)

	return result
}

// NewMonthAlbum creates a new month album.
func NewMonthAlbum(albumTitle, albumSlug string, year, month int) *Album {
	albumTitle = strings.TrimSpace(albumTitle)
	albumSlug = strings.TrimSpace(albumSlug)

	if albumTitle == "" || albumSlug == "" || year == 0 || month == 0 {
		return nil
	}

	f := form.SearchPhotos{
		Year:   strconv.Itoa(year),
		Month:  strconv.Itoa(month),
		Public: true,
	}

	now := TimeStamp()

	result := &Album{
		AlbumOrder:  SortOrderOldest,
		AlbumType:   AlbumMonth,
		AlbumSlug:   albumSlug,
		AlbumFilter: f.Serialize(),
		AlbumYear:   year,
		AlbumMonth:  month,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result.SetTitle(albumTitle)

	return result
}

// FindMonthAlbum finds a matching month album or returns nil.
func FindMonthAlbum(year, month int) *Album {
	m := Album{}

	if UnscopedDb().First(&m, "album_year = ? AND album_month = ? AND album_type = ?", year, month, AlbumMonth).RecordNotFound() {
		return nil
	}

	return &m
}

// FindAlbumBySlug finds a matching album or returns nil.
func FindAlbumBySlug(albumSlug, albumType string) *Album {
	m := Album{}

	if UnscopedDb().First(&m, "album_slug = ? AND album_type = ?", albumSlug, albumType).RecordNotFound() {
		return nil
	}

	return &m
}

// FindAlbumByAttr finds an album by filters and slugs, or returns nil.
func FindAlbumByAttr(slugs, filters []string, albumType string) *Album {
	m := Album{}

	stmt := UnscopedDb()

	if albumType != "" {
		stmt = stmt.Where("album_type = ?", albumType)
	}

	stmt = stmt.Where("album_slug IN (?) OR album_filter IN (?)", slugs, filters)

	if stmt.First(&m).RecordNotFound() {
		return nil
	}

	return &m
}

// FindFolderAlbum finds a matching folder album or returns nil.
func FindFolderAlbum(albumPath string) *Album {
	albumPath = strings.Trim(albumPath, string(os.PathSeparator))
	albumSlug := txt.Slug(albumPath)

	if albumSlug == "" {
		return nil
	}

	m := Album{}

	stmt := UnscopedDb().Where("album_type = ?", AlbumFolder).
		Where("album_slug = ? OR album_path = ?", albumSlug, albumPath)

	if stmt.First(&m).RecordNotFound() {
		return nil
	}

	return &m
}

// FindAlbum retrieves the matching record from the database and updates the entity.
func FindAlbum(find Album) *Album {
	m := Album{}

	// Search by UID.
	if rnd.IsUID(find.AlbumUID, AlbumUID) {
		if UnscopedDb().First(&m, "album_uid = ?", find.AlbumUID).RecordNotFound() {
			return nil
		} else if m.AlbumUID != "" {
			albumCache.SetDefault(m.AlbumUID, m)
			return &m
		}
	}

	// Otherwise, album type and slug are required.
	if find.AlbumType == "" || find.AlbumSlug == "" {
		return nil
	}

	// Create search condition.
	stmt := UnscopedDb().Where("album_type = ?", find.AlbumType)
	if find.AlbumType != AlbumDefault && find.AlbumFilter != "" {
		stmt = stmt.Where("album_slug = ? OR album_filter = ?", find.AlbumSlug, find.AlbumFilter)
	} else {
		stmt = stmt.Where("album_slug = ? OR album_title LIKE ?", find.AlbumSlug, find.AlbumTitle)
	}

	// Find first matching record.
	if stmt.First(&m).RecordNotFound() {
		return nil
	}

	// Cache result.
	if m.AlbumUID != "" {
		albumCache.SetDefault(m.AlbumUID, m)
	}

	return &m
}

// Find retrieves the matching record from the database and updates the entity.
func (m *Album) Find() *Album {
	return FindAlbum(*m)
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUnique(m.AlbumUID, AlbumUID) {
		return nil
	}

	return scope.SetColumn("AlbumUID", rnd.GenerateUID(AlbumUID))
}

// String returns the id or name as string.
func (m *Album) String() string {
	if m.AlbumSlug != "" {
		return clean.Log(m.AlbumSlug)
	}

	if m.AlbumTitle != "" {
		return clean.Log(m.AlbumTitle)
	}

	if m.AlbumUID != "" {
		return clean.Log(m.AlbumUID)
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
	title = strings.Trim(title, "_&|{}<>: \n\r\t\\")
	title = strings.ReplaceAll(title, "\"", "'")
	title = txt.Shorten(title, txt.ClipDefault, txt.Ellipsis)

	if title == "" {
		title = m.CreatedAt.Format("January 2006")
	}

	m.AlbumTitle = title

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
	title = txt.Clip(title, txt.ClipDefault)
	slug = txt.Clip(slug, txt.ClipSlug)

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

	if title != "" {
		m.SetTitle(title)
	}

	return m.Updates(Values{"album_title": m.AlbumTitle, "album_slug": m.AlbumSlug})
}

// UpdateState updates the album location.
func (m *Album) UpdateState(title, slug, stateName, countryCode string) error {
	title = txt.Clip(title, txt.ClipDefault)
	slug = txt.Clip(slug, txt.ClipSlug)

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

	if title != "" {
		m.SetTitle(title)
	}

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

	return m.Save()
}

// Update sets a new value for a database column.
func (m *Album) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).Update(attr, value).Error
}

// Updates multiple columns in the database.
func (m *Album) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// UpdateFolder updates the path, filter and slug for a folder album.
func (m *Album) UpdateFolder(albumPath, albumFilter string) error {
	albumPath = strings.Trim(albumPath, string(os.PathSeparator))
	albumSlug := txt.Slug(albumPath)

	if albumSlug == "" {
		return nil
	}

	if err := UnscopedDb().Model(m).Updates(map[string]interface{}{
		"AlbumPath":   albumPath,
		"AlbumFilter": albumFilter,
		"AlbumSlug":   albumSlug,
	}).Error; err != nil {
		return err
	} else if err = UnscopedDb().Exec("UPDATE albums SET album_path = NULL WHERE album_path = ? AND id <> ?", albumPath, m.ID).Error; err != nil {
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
	if m.DeletedAt == nil {
		return false
	}

	return !m.DeletedAt.IsZero()
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
	s := txt.Slug(m.AlbumTitle)

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
