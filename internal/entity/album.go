package entity

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/service/maps"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	AlbumUID    = byte('a')
	AlbumManual = "album"
	AlbumFolder = "folder"
	AlbumMoment = "moment"
	AlbumMonth  = "month"
	AlbumState  = "state"
)

var (
	DefaultOrderAlbum  = sortby.Oldest
	DefaultOrderFolder = sortby.Added
	DefaultOrderMoment = sortby.Oldest
	DefaultOrderState  = sortby.Newest
	DefaultOrderMonth  = sortby.Oldest
)

var (
	albumGlobalLock = sync.Mutex{}
	albumLocks      sync.Map // map[string]*sync.Mutex keyed by UID or normalized title
)

// Albums is a helper slice type for working with groups of albums.
type Albums []Album

// Album represents a photo album and its metadata, including filter definitions for virtual albums.
type Album struct {
	ID               uint        `gorm:"primary_key" json:"ID" yaml:"-"`
	AlbumUID         string      `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	ParentUID        string      `gorm:"type:VARBINARY(42);default:'';" json:"ParentUID,omitempty" yaml:"ParentUID,omitempty"`
	AlbumSlug        string      `gorm:"type:VARBINARY(160);index;" json:"Slug" yaml:"Slug"`
	AlbumPath        string      `gorm:"type:VARCHAR(1024);index;" json:"Path,omitempty" yaml:"Path,omitempty"`
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
	CreatedBy        string      `gorm:"type:VARBINARY(42);index" json:"CreatedBy,omitempty" yaml:"CreatedBy,omitempty"`
	CreatedAt        time.Time   `json:"CreatedAt" yaml:"CreatedAt,omitempty"`
	UpdatedAt        time.Time   `json:"UpdatedAt" yaml:"UpdatedAt,omitempty"`
	PublishedAt      *time.Time  `sql:"index" json:"PublishedAt,omitempty" yaml:"PublishedAt,omitempty"`
	DeletedAt        *time.Time  `sql:"index" json:"DeletedAt" yaml:"DeletedAt,omitempty"`
	Photos           PhotoAlbums `gorm:"foreignkey:AlbumUID;association_foreignkey:AlbumUID;" json:"-" yaml:"Photos,omitempty"`
}

// AfterUpdate flushes the album cache when an album is updated.
func (m *Album) AfterUpdate(tx *gorm.DB) (err error) {
	FlushAlbumCache()
	return
}

// AfterDelete flushes the album cache when an album is deleted.
func (m *Album) AfterDelete(tx *gorm.DB) (err error) {
	FlushAlbumCache()
	return
}

// TableName returns the entity table name.
func (Album) TableName() string {
	return "albums"
}

// UpdateAlbum updates album attributes directly in the database by UID.
func UpdateAlbum(albumUID string, values interface{}) (err error) {
	if rnd.InvalidUID(albumUID, AlbumUID) {
		return fmt.Errorf("album: invalid uid %s", clean.Log(albumUID))
	} else if err = Db().Model(Album{}).Where("album_uid = ?", albumUID).UpdateColumns(values).Error; err != nil {
		return err
	}

	return nil
}

// AddPhotoToAlbums adds a photo UID to multiple albums and automatically creates them with default ownership when required.
func AddPhotoToAlbums(uid string, albums []string) (err error) {
	return AddPhotoToUserAlbums(uid, albums, DefaultOrderAlbum, OwnerUnknown)
}

// AddPhotoToUserAlbums adds a photo UID to multiple albums while creating any missing albums for the given user.
func AddPhotoToUserAlbums(photoUid string, albums []string, sortOrder, userUid string) (err error) {
	if photoUid == "" || len(albums) == 0 {
		// Do nothing.
		return nil
	}

	if !rnd.IsUID(photoUid, PhotoUID) {
		return fmt.Errorf("album: can not add invalid photo uid %s", clean.Log(photoUid))
	}

	for _, album := range albums {
		if album == "" {
			log.Debugf("album: cannot add photo uid %s because album id was not specified", clean.Log(photoUid))
			continue
		}

		unlock := lockAlbumKey(album)
		if lockErr := addPhotoToAlbumLocked(photoUid, album, sortOrder, userUid); lockErr != nil {
			err = lockErr
		}
		unlock()
	}

	return err
}

// lockAlbumKey acquires a per-album mutex keyed by UID or normalized title to avoid
// serializing unrelated album updates while still preventing duplicate creation when
// multiple goroutines target the same album concurrently.
func lockAlbumKey(album string) func() {
	key := strings.TrimSpace(album)

	if key == "" {
		albumGlobalLock.Lock()
		return albumGlobalLock.Unlock
	}

	if rnd.IsUID(key, AlbumUID) {
		// keep UID as-is so existing albums share the same lock
	} else {
		key = strings.ToLower(key)
	}

	locker, _ := albumLocks.LoadOrStore(key, &sync.Mutex{})
	mu := locker.(*sync.Mutex)
	mu.Lock()

	return mu.Unlock
}

// addPhotoToAlbumLocked performs the actual album lookup/creation and relation insert
// while assuming the caller already holds the per-album mutex.
func addPhotoToAlbumLocked(photoUid, album, sortOrder, userUid string) (err error) {
	var albumUid string

	if rnd.IsUID(album, AlbumUID) {
		albumUid = album
	} else {
		a := NewUserAlbum(album, AlbumManual, sortOrder, userUid)

		if found := a.Find(); found != nil {
			albumUid = found.AlbumUID
		} else if err = a.Create(); err == nil {
			albumUid = a.AlbumUID
		} else {
			log.Errorf("album: %s (add photo %s to albums)", err.Error(), photoUid)
		}
	}

	if albumUid == "" {
		return err
	}

	entry := PhotoAlbum{AlbumUID: albumUid, PhotoUID: photoUid, Hidden: false}

	if err = entry.Save(); err != nil {
		log.Errorf("album: %s (add photo %s to albums)", err.Error(), photoUid)
	}

	if updateErr := UpdateAlbum(albumUid, Values{"updated_at": TimeStamp()}); updateErr != nil {
		if err == nil {
			err = updateErr
		}
	}

	return err
}

// NewAlbum creates a new album of the given type using default ownership.
func NewAlbum(albumTitle, albumType string) *Album {
	return NewUserAlbum(albumTitle, albumType, sortby.Oldest, OwnerUnknown)
}

// NewUserAlbum creates a new album owned by a user and pre-fills timestamps/order defaults.
func NewUserAlbum(albumTitle, albumType, sortOrder, userUid string) *Album {
	now := Now()

	// Set default type.
	if albumType == "" {
		albumType = AlbumManual
	}

	// Set default sort order.
	if sortOrder == "" {
		sortOrder = DefaultOrderAlbum
	}

	// Set default values.
	result := &Album{
		AlbumOrder: sortOrder,
		AlbumType:  albumType,
		CreatedAt:  now,
		UpdatedAt:  now,
		CreatedBy:  userUid,
	}

	// Set album title.
	result.SetTitle(albumTitle)

	return result
}

// NewFolderAlbum creates a new album representing a filesystem folder.
func NewFolderAlbum(albumTitle, albumPath, albumFilter string) *Album {
	albumSlug := txt.Slug(albumPath)

	if albumTitle == "" || albumSlug == "" || albumPath == "" || albumFilter == "" {
		return nil
	}

	now := Now()

	result := &Album{
		AlbumOrder:  DefaultOrderFolder,
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

// NewMomentsAlbum creates a new automatically generated moment album.
func NewMomentsAlbum(albumTitle, albumSlug, albumFilter string) *Album {
	if albumTitle == "" || albumSlug == "" || albumFilter == "" {
		return nil
	}

	now := Now()

	result := &Album{
		AlbumOrder:  DefaultOrderMoment,
		AlbumType:   AlbumMoment,
		AlbumSlug:   txt.Clip(albumSlug, txt.ClipSlug),
		AlbumFilter: albumFilter,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result.SetTitle(albumTitle)

	return result
}

// NewStateAlbum creates an automatically generated album for a specific state or region.
func NewStateAlbum(albumTitle, albumSlug, albumFilter string) *Album {
	albumTitle = strings.TrimSpace(albumTitle)
	albumSlug = strings.TrimSpace(albumSlug)

	if albumTitle == "" || albumSlug == "" || albumFilter == "" {
		return nil
	}

	now := Now()

	result := &Album{
		AlbumOrder:  DefaultOrderState,
		AlbumType:   AlbumState,
		AlbumSlug:   txt.Clip(albumSlug, txt.ClipSlug),
		AlbumFilter: albumFilter,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result.SetTitle(albumTitle)

	return result
}

// NewMonthAlbum creates an automatically generated album for a specific month.
func NewMonthAlbum(albumTitle, albumSlug string, year, month int) *Album {
	albumTitle = strings.TrimSpace(albumTitle)
	albumSlug = strings.TrimSpace(albumSlug)

	if albumTitle == "" || albumSlug == "" || year < 1 || month < 1 || month > 12 {
		return nil
	}

	f := form.SearchPhotos{
		Year:   strconv.Itoa(year),
		Month:  strconv.Itoa(month),
		Public: true,
	}

	now := Now()

	result := &Album{
		AlbumOrder:  DefaultOrderMonth,
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

// FindMonthAlbum returns the matching month album or nil if none exists.
func FindMonthAlbum(year, month int) *Album {
	m := Album{}

	if year < 1 || month < 1 || month > 12 {
		return nil
	}

	if UnscopedDb().First(&m, "album_year = ? AND album_month = ? AND album_type = ?", year, month, AlbumMonth).Error != nil {
		return nil
	}

	return &m
}

// FindAlbumBySlug returns the album with the given slug/type combination.
func FindAlbumBySlug(albumSlug, albumType string) *Album {
	m := Album{}

	if albumSlug == "" || albumSlug == UnknownSlug {
		return nil
	}

	if UnscopedDb().First(&m, "album_slug = ? AND album_type = ?", albumSlug, albumType).Error != nil {
		return nil
	}

	return &m
}

// FindAlbumByAttr returns an album matching any of the provided slugs or filters.
func FindAlbumByAttr(slugs, filters []string, albumType string) *Album {
	m := Album{}

	if len(slugs) == 0 && len(filters) == 0 {
		return nil
	}

	stmt := UnscopedDb()

	if albumType != "" {
		stmt = stmt.Where("album_type = ?", albumType)
	}

	if len(filters) == 0 {
		stmt = stmt.Where("album_slug IN (?)", slugs)
	} else {
		stmt = stmt.Where("album_slug IN (?) OR album_filter IN (?)", slugs, filters)
	}

	if stmt.First(&m).Error != nil {
		return nil
	}

	return &m
}

// FindFolderAlbum looks up a folder album by its canonical path or slug.
func FindFolderAlbum(albumPath string) *Album {
	albumPath = strings.Trim(albumPath, string(os.PathSeparator))
	albumSlug := txt.Slug(albumPath)

	if albumSlug == "" {
		return nil
	}

	m := Album{}

	stmt := UnscopedDb().Where("album_type = ?", AlbumFolder).
		Where("album_slug = ? OR album_path = ?", albumSlug, albumPath)

	if stmt.First(&m).Error != nil {
		return nil
	}

	return &m
}

// AlbumSearch creates a new Album to be used as parameter for FindAlbum.
func AlbumSearch(albumUid, albumTitle, albumType string) Album {
	// Set default type.
	if albumType == "" {
		albumType = AlbumManual
	}

	// Set default values.
	result := Album{
		AlbumType: albumType,
		AlbumUID:  albumUid,
	}

	// Set album title.
	if albumTitle != "" {
		result.SetTitle(albumTitle)
	}

	return result
}

// FindAlbum retrieves the matching record from the database and updates the entity.
func FindAlbum(find Album) *Album {
	m := Album{}

	// Search by UID.
	if rnd.IsUID(find.AlbumUID, AlbumUID) {
		if UnscopedDb().First(&m, "album_uid = ?", find.AlbumUID).Error != nil {
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

	// Search by slug and filter or title.
	if find.AlbumType != AlbumManual {
		if find.AlbumFilter != "" && find.AlbumSlug != "" && find.AlbumSlug != UnknownSlug {
			stmt = stmt.Where("album_slug = ? OR album_filter = ?", find.AlbumSlug, find.AlbumFilter)
		} else if find.AlbumFilter != "" {
			stmt = stmt.Where("album_filter = ?", find.AlbumFilter)
		} else if find.AlbumSlug != "" && find.AlbumSlug != UnknownSlug {
			stmt = stmt.Where("album_slug = ?", find.AlbumSlug)
		} else {
			return nil
		}
	} else if find.AlbumTitle != "" && find.AlbumSlug != "" && find.AlbumSlug != UnknownSlug {
		stmt = stmt.Where("album_slug = ? OR album_title LIKE ?", find.AlbumSlug, find.AlbumTitle)
	} else if find.AlbumSlug != "" && find.AlbumSlug != UnknownSlug {
		stmt = stmt.Where("album_slug = ?", find.AlbumSlug)
	} else if find.AlbumTitle != "" {
		stmt = stmt.Where("album_title LIKE ?", find.AlbumTitle)
	} else {
		return nil
	}

	// Filter by creator if the album has not been published yet.
	if find.CreatedBy != "" {
		stmt = stmt.Where("published_at > ? OR created_by = ?", Now(), find.CreatedBy)
	}

	// Find first matching record.
	if stmt.First(&m).Error != nil {
		return nil
	}

	// Cache result.
	if m.AlbumUID != "" {
		albumCache.SetDefault(m.AlbumUID, m)
	}

	return &m
}

// HasID tests if the album has a valid id and uid.
func (m *Album) HasID() bool {
	if m == nil {
		return false
	}

	return m.ID > 0 && rnd.IsUID(m.AlbumUID, AlbumUID)
}

// Find retrieves the matching record from the database and updates the entity.
func (m *Album) Find() *Album {
	return FindAlbum(*m)
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Album) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.AlbumUID, AlbumUID) {
		return nil
	}

	m.AlbumUID = rnd.GenerateUID(AlbumUID)

	return scope.SetColumn("AlbumUID", m.AlbumUID)
}

// String returns the id or name as string.
func (m *Album) String() string {
	if m == nil {
		return "Album<nil>"
	}

	if m.AlbumSlug != "" && m.AlbumSlug != UnknownSlug {
		return clean.Log(m.AlbumSlug)
	}

	if m.AlbumTitle != "" {
		return clean.Log(m.AlbumTitle)
	}

	if m.AlbumUID != "" {
		return clean.Log(m.AlbumUID)
	}

	return "*Album"
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
	return m.AlbumType == AlbumManual
}

// SetTitle changes the album name.
func (m *Album) SetTitle(title string) *Album {
	title = strings.Trim(title, "_&|{}<>: \n\r\t\\")
	title = strings.ReplaceAll(title, "\"", "“")
	title = txt.Shorten(title, txt.ClipDefault, txt.Ellipsis)

	if title == "" {
		title = m.CreatedAt.Format("January 2006")
	}

	m.AlbumTitle = title

	if m.AlbumType == AlbumManual || m.AlbumSlug == "" {
		if len(m.AlbumTitle) < txt.ClipSlug {
			m.AlbumSlug = txt.Slug(m.AlbumTitle)
		} else {
			m.AlbumSlug = txt.Slug(m.AlbumTitle) + "-" + m.AlbumUID
		}
	}

	if m.AlbumSlug == "" {
		m.AlbumSlug = UnknownSlug
	}

	return m
}

// SetLocation sets a new album location.
func (m *Album) SetLocation(location, state, country string) *Album {
	if location != "" {
		m.AlbumLocation = txt.Shorten(location, txt.ClipDefault, txt.Ellipsis)
	}

	if state != "" || country != "" && country != "zz" {
		m.AlbumCountry = txt.Clip(country, txt.ClipCountry)
		m.AlbumState = txt.Clip(clean.State(state, country), txt.ClipCategory)
	}

	return m
}

// UpdateTitleAndLocation updates title, location, and slug of generated albums if needed.
func (m *Album) UpdateTitleAndLocation(title, location, state, country, slug string) error {
	if !m.HasID() {
		return fmt.Errorf("album does not exist")
	}

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

	if !changed && state == m.AlbumState && (country == m.AlbumCountry || country == "" && m.AlbumCountry == "zz") {
		return nil
	}

	m.SetTitle(title)

	// Skip location?
	if location == "" && state == "" && (country == "" || country == "zz") {
		return m.Updates(Values{
			"album_title": m.AlbumTitle,
			"album_slug":  m.AlbumSlug,
		})
	}

	m.SetLocation(location, state, country)

	return m.Updates(Values{
		"album_title":    m.AlbumTitle,
		"album_location": m.AlbumLocation,
		"album_state":    m.AlbumState,
		"album_country":  m.AlbumCountry,
		"album_slug":     m.AlbumSlug,
	})
}

// UpdateTitleAndState updates the album location.
func (m *Album) UpdateTitleAndState(title, slug, stateName, countryCode string) error {
	if !m.HasID() {
		return fmt.Errorf("album does not exist")
	}

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

	m.SetTitle(title)

	return m.Updates(Values{"album_title": m.AlbumTitle, "album_slug": m.AlbumSlug, "album_location": m.AlbumLocation, "album_country": m.AlbumCountry, "album_state": m.AlbumState})
}

// SaveForm updates the entity using form data and stores it in the database.
func (m *Album) SaveForm(f *form.Album) error {
	if m == nil {
		return errors.New("album must not be nil - you may have found a bug")
	} else if f == nil {
		return fmt.Errorf("form is nil")
	}

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
	if m == nil {
		return errors.New("album must not be nil - you may have found a bug")
	} else if !m.HasID() {
		return fmt.Errorf("album does not exist")
	}

	return UnscopedDb().Model(m).Update(attr, value).Error
}

// Updates multiple columns in the database.
func (m *Album) Updates(values interface{}) error {
	if m == nil {
		return errors.New("album must not be nil - you may have found a bug")
	} else if !m.HasID() {
		return fmt.Errorf("album does not exist")
	}

	return UnscopedDb().Model(m).Updates(values).Error
}

// UpdateFolder updates the path, filter and slug for a folder album.
func (m *Album) UpdateFolder(albumPath, albumFilter string) error {
	if !m.HasID() {
		return fmt.Errorf("album does not exist")
	}

	albumPath = strings.Trim(albumPath, string(os.PathSeparator))
	albumSlug := txt.Slug(albumPath)

	if albumSlug == "" || albumPath == "" || albumFilter == "" || !m.HasID() {
		return fmt.Errorf("folder album must have a path and filter")
	} else if m.AlbumPath == albumPath && m.AlbumFilter == albumFilter && m.AlbumSlug == albumSlug {
		// Nothing changed.
		return nil
	}

	if err := m.Updates(Values{
		"AlbumPath":   albumPath,
		"AlbumFilter": albumFilter,
		"AlbumSlug":   albumSlug,
	}); err != nil {
		return err
	} else if err = UnscopedDb().Exec("UPDATE albums SET album_path = NULL WHERE album_type = ? AND album_path = ? AND id <> ?", AlbumFolder, albumPath, m.ID).Error; err != nil {
		return err
	}

	return nil
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Album) Save() error {
	if err := Db().Save(m).Error; err != nil {
		return err
	} else {
		event.PublishUserEntities("albums", event.EntityUpdated, []*Album{m}, m.CreatedBy)
		return nil
	}
}

// Create inserts a new row to the database.
func (m *Album) Create() error {
	if err := Db().Create(m).Error; err != nil {
		return err
	}

	m.PublishCountChange(1)
	event.PublishUserEntities("albums", event.EntityCreated, []*Album{m}, m.CreatedBy)

	return nil
}

// PublishCountChange publishes an event with the added or removed number of albums.
func (m *Album) PublishCountChange(n int) {
	data := event.Data{"count": n}

	switch m.AlbumType {
	case AlbumManual:
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
	if !m.HasID() {
		return fmt.Errorf("album does not exist")
	}

	if m.Deleted() {
		return nil
	}

	now := Now()

	if err := UnscopedDb().Model(m).UpdateColumns(Values{"updated_at": now, "deleted_at": now}).Error; err != nil {
		return err
	} else {
		m.UpdatedAt = now
		m.DeletedAt = &now
		FlushAlbumCache()
	}

	m.PublishCountChange(-1)
	event.EntitiesDeleted("albums", []string{m.AlbumUID})

	return DeleteShareLinks(m.AlbumUID)
}

// DeletePermanently permanently removes an album from the index.
func (m *Album) DeletePermanently() error {
	if !m.HasID() {
		return fmt.Errorf("album does not exist")
	}

	wasDeleted := m.Deleted()

	if err := UnscopedDb().Delete(m).Error; err != nil {
		return err
	}

	if !wasDeleted {
		m.PublishCountChange(-1)
		event.EntitiesDeleted("albums", []string{m.AlbumUID})
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
	if !m.HasID() {
		return fmt.Errorf("album does not exist")
	}

	if !m.Deleted() {
		return nil
	}

	if err := UnscopedDb().Model(m).Update("DeletedAt", nil).Error; err != nil {
		return err
	}

	m.DeletedAt = nil

	m.PublishCountChange(1)
	event.PublishUserEntities("albums", event.EntityCreated, []*Album{m}, m.CreatedBy)

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
func (m *Album) AddPhotos(photos PhotosInterface) (added PhotoAlbums) {
	if !m.HasID() {
		return added
	}

	// Add album entries.
	for _, photoUid := range photos.UIDs() {
		if !rnd.IsUID(photoUid, PhotoUID) {
			continue
		}

		// Add photo to album.
		entry := PhotoAlbum{AlbumUID: m.AlbumUID, PhotoUID: photoUid, Hidden: false}

		// Save album entry.
		if err := entry.Save(); err != nil {
			log.Errorf("album: %s (add to album %s)", err.Error(), m)
		} else {
			added = append(added, entry)
		}
	}

	// Refresh updated timestamp.
	if err := UpdateAlbum(m.AlbumUID, Values{"updated_at": TimeStamp()}); err != nil {
		log.Errorf("album: %s (update %s)", err.Error(), m)
	}

	return added
}

// RemovePhotos removes photos from an album.
func (m *Album) RemovePhotos(UIDs []string) (removed PhotoAlbums) {
	if !m.HasID() {
		return removed
	}

	for _, uid := range UIDs {
		if !rnd.IsUID(uid, PhotoUID) {
			continue
		}

		entry := PhotoAlbum{AlbumUID: m.AlbumUID, PhotoUID: uid, Hidden: true}

		if err := entry.Save(); err != nil {
			log.Errorf("album: %s (remove from album %s)", err.Error(), m)
		} else {
			removed = append(removed, entry)
		}
	}

	// Refresh updated timestamp.
	if err := UpdateAlbum(m.AlbumUID, Values{"updated_at": TimeStamp()}); err != nil {
		log.Errorf("album: %s (update %s)", err.Error(), m)
	}

	return removed
}

// Links returns all share links for this entity.
func (m *Album) Links() Links {
	return FindLinks("", m.AlbumUID)
}
