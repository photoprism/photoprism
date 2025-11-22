package entity

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/react"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/tz"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	PhotoUID = byte('p')
)

var IndexUpdateInterval = 3 * time.Hour           // 3 Hours
var MetadataUpdateInterval = 24 * 3 * time.Hour   // 3 Days
var MetadataEstimateInterval = 24 * 7 * time.Hour // 7 Days

var photoMutex = sync.Mutex{}
var labelKeywordsSkipSrc = []string{SrcTitle, SrcCaption, SrcSubject, SrcKeyword}

// MapKey builds a deterministic indexing key from the capture timestamp and spatial cell identifier.
func MapKey(takenAt time.Time, cellId string) string {
	return path.Join(strconv.FormatInt(takenAt.Unix(), 36), cellId)
}

// Photo represents a photo, all its properties, and link to all its images and sidecar files.
type Photo struct {
	ID               uint          `gorm:"primary_key" yaml:"-"`
	UUID             string        `gorm:"type:VARBINARY(64);index;" json:"DocumentID,omitempty" yaml:"DocumentID,omitempty"`
	TakenAt          time.Time     `gorm:"type:DATETIME;index:idx_photos_taken_uid;" json:"TakenAt" yaml:"TakenAt"`
	TakenAtLocal     time.Time     `gorm:"type:DATETIME;" json:"TakenAtLocal" yaml:"TakenAtLocal"`
	TakenSrc         string        `gorm:"type:VARBINARY(8);" json:"TakenSrc" yaml:"TakenSrc,omitempty"`
	PhotoUID         string        `gorm:"type:VARBINARY(42);unique_index;index:idx_photos_taken_uid;" json:"UID" yaml:"UID"`
	PhotoType        string        `gorm:"type:VARBINARY(8);default:'image';" json:"Type" yaml:"Type"`
	TypeSrc          string        `gorm:"type:VARBINARY(8);" json:"TypeSrc" yaml:"TypeSrc,omitempty"`
	PhotoTitle       string        `gorm:"type:VARCHAR(200);" json:"Title" yaml:"Title"`
	TitleSrc         string        `gorm:"type:VARBINARY(8);" json:"TitleSrc" yaml:"TitleSrc,omitempty"`
	PhotoCaption     string        `gorm:"type:VARCHAR(4096);" json:"Caption" yaml:"Caption,omitempty"`
	CaptionSrc       string        `gorm:"type:VARBINARY(8);" json:"CaptionSrc" yaml:"CaptionSrc,omitempty"`
	PhotoDescription string        `gorm:"-" json:"Description,omitempty" yaml:"Description,omitempty"`
	DescriptionSrc   string        `gorm:"-" json:"DescriptionSrc,omitempty" yaml:"DescriptionSrc,omitempty"`
	PhotoPath        string        `gorm:"type:VARBINARY(1024);index:idx_photos_path_name;" json:"Path" yaml:"-"`
	PhotoName        string        `gorm:"type:VARBINARY(255);index:idx_photos_path_name;" json:"Name" yaml:"-"`
	OriginalName     string        `gorm:"type:VARBINARY(755);" json:"OriginalName" yaml:"OriginalName,omitempty"`
	PhotoStack       int8          `json:"Stack" yaml:"Stack,omitempty"`
	PhotoFavorite    bool          `json:"Favorite" yaml:"Favorite,omitempty"`
	PhotoPrivate     bool          `json:"Private" yaml:"Private,omitempty"`
	PhotoScan        bool          `json:"Scan" yaml:"Scan,omitempty"`
	PhotoPanorama    bool          `json:"Panorama" yaml:"Panorama,omitempty"`
	TimeZone         string        `gorm:"type:VARBINARY(64);default:'Local'" json:"TimeZone" yaml:"TimeZone,omitempty"`
	PlaceID          string        `gorm:"type:VARBINARY(42);index;default:'zz'" json:"PlaceID" yaml:"-"`
	PlaceSrc         string        `gorm:"type:VARBINARY(8);" json:"PlaceSrc" yaml:"PlaceSrc,omitempty"`
	CellID           string        `gorm:"type:VARBINARY(42);index;default:'zz'" json:"CellID" yaml:"-"`
	CellAccuracy     int           `json:"CellAccuracy" yaml:"CellAccuracy,omitempty"`
	PhotoAltitude    int           `json:"Altitude" yaml:"Altitude,omitempty"`
	PhotoLat         float64       `gorm:"type:DOUBLE;index;" json:"Lat" yaml:"Lat,omitempty"`
	PhotoLng         float64       `gorm:"type:DOUBLE;index;" json:"Lng" yaml:"Lng,omitempty"`
	PhotoCountry     string        `gorm:"type:VARBINARY(2);index:idx_photos_country_year_month;default:'zz'" json:"Country" yaml:"-"`
	PhotoYear        int           `gorm:"index:idx_photos_ymd;index:idx_photos_country_year_month;" json:"Year" yaml:"Year"`
	PhotoMonth       int           `gorm:"index:idx_photos_ymd;index:idx_photos_country_year_month;" json:"Month" yaml:"Month"`
	PhotoDay         int           `gorm:"index:idx_photos_ymd" json:"Day" yaml:"Day"`
	PhotoIso         int           `json:"Iso" yaml:"ISO,omitempty"`
	PhotoExposure    string        `gorm:"type:VARBINARY(64);" json:"Exposure" yaml:"Exposure,omitempty"`
	PhotoFNumber     float32       `gorm:"type:FLOAT;" json:"FNumber" yaml:"FNumber,omitempty"`
	PhotoFocalLength int           `json:"FocalLength" yaml:"FocalLength,omitempty"`
	PhotoQuality     int           `gorm:"type:SMALLINT" json:"Quality" yaml:"Quality,omitempty"`
	PhotoFaces       int           `json:"Faces,omitempty" yaml:"Faces,omitempty"`
	PhotoResolution  int           `gorm:"type:SMALLINT" json:"Resolution" yaml:"-"`
	PhotoDuration    time.Duration `json:"Duration,omitempty" yaml:"Duration,omitempty"`
	PhotoColor       int16         `json:"Color" yaml:"-"`
	CameraID         uint          `gorm:"index:idx_photos_camera_lens;default:1" json:"CameraID" yaml:"-"`
	CameraSerial     string        `gorm:"type:VARBINARY(160);" json:"CameraSerial" yaml:"CameraSerial,omitempty"`
	CameraSrc        string        `gorm:"type:VARBINARY(8);" json:"CameraSrc" yaml:"-"`
	LensID           uint          `gorm:"index:idx_photos_camera_lens;default:1" json:"LensID" yaml:"-"`
	Details          *Details      `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Details" yaml:"Details"`
	Camera           *Camera       `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Camera" yaml:"-"`
	Lens             *Lens         `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Lens" yaml:"-"`
	Cell             *Cell         `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Cell" yaml:"-"`
	Place            *Place        `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Place" yaml:"-"`
	Keywords         []Keyword     `json:"-" yaml:"-"`
	Albums           []Album       `json:"Albums" yaml:"-"`
	Files            []File        `yaml:"-"`
	Labels           []PhotoLabel  `yaml:"-"`
	CreatedBy        string        `gorm:"type:VARBINARY(42);index" json:"CreatedBy,omitempty" yaml:"CreatedBy,omitempty"`
	CreatedAt        time.Time     `json:"CreatedAt,omitempty" yaml:"CreatedAt,omitempty"`
	UpdatedAt        time.Time     `json:"UpdatedAt,omitempty" yaml:"UpdatedAt,omitempty"`
	EditedAt         *time.Time    `json:"EditedAt,omitempty" yaml:"EditedAt,omitempty"`
	PublishedAt      *time.Time    `sql:"index" json:"PublishedAt,omitempty" yaml:"PublishedAt,omitempty"`
	IndexedAt        *time.Time    `json:"IndexedAt,omitempty" yaml:"-"`
	CheckedAt        *time.Time    `sql:"index" json:"CheckedAt,omitempty" yaml:"-"`
	EstimatedAt      *time.Time    `json:"EstimatedAt,omitempty" yaml:"-"`
	DeletedAt        *time.Time    `sql:"index" json:"DeletedAt,omitempty" yaml:"DeletedAt,omitempty"`
}

// TableName returns the entity table name.
func (Photo) TableName() string {
	return "photos"
}

// NewPhoto returns a Photo with default metadata placeholders and the requested stack flag.
func NewPhoto(stackable bool) Photo {
	return NewUserPhoto(stackable, "")
}

// NewUserPhoto returns a Photo initialized for the given user UID, including default Unknown* references and stack state.
func NewUserPhoto(stackable bool, userUid string) Photo {
	m := Photo{
		PhotoTitle:   UnknownTitle,
		PhotoType:    MediaImage,
		PhotoCountry: UnknownCountry.ID,
		CameraID:     UnknownCamera.ID,
		LensID:       UnknownLens.ID,
		CellID:       UnknownLocation.ID,
		PlaceID:      UnknownPlace.ID,
		TimeZone:     tz.Local,
		Camera:       &UnknownCamera,
		Lens:         &UnknownLens,
		Cell:         &UnknownLocation,
		Place:        &UnknownPlace,
		CreatedBy:    userUid,
	}

	if stackable {
		m.PhotoStack = IsStackable
	} else {
		m.PhotoStack = IsUnstacked
	}

	return m
}

// SavePhotoForm merges a photo form submission into the Photo, normalizes data, refreshes derived metadata, and persists the changes.
// The photo must already exist in the database; after saving, derived counters are updated asynchronously.
func SavePhotoForm(m *Photo, form form.Photo) error {
	if m == nil {
		return fmt.Errorf("photo is nil")
	}

	locChanged := m.PhotoLat != form.PhotoLat || m.PhotoLng != form.PhotoLng || m.PhotoCountry != form.PhotoCountry

	if err := deepcopier.Copy(m).From(form); err != nil {
		return err
	}

	m.NormalizeValues()

	if !m.HasID() {
		return errors.New("cannot save form when photo id is missing")
	}

	// Update time fields.
	// Batch edit (and other callers) treat TakenAtLocal as a naive timestamp that already includes
	// any new Day/Month/Year. Here we normalize it back to UTC using the photo's TimeZone so MariaDB
	// (which lacks TZ support) stores a consistent pair of TakenAt / TakenAtLocal values. See
	// ComputeDateChange for details on why it returns UTC.
	if m.TimeZoneUTC() {
		m.TakenAtLocal = m.TakenAt
	} else {
		m.TakenAt = m.GetTakenAt()
	}

	m.UpdateDateFields()

	details := m.GetDetails()

	if form.Details.PhotoID == m.ID {
		if err := deepcopier.Copy(details).From(form.Details); err != nil {
			return err
		}

		details.Keywords = strings.Join(txt.UniqueWords(txt.Words(details.Keywords)), ", ")
	}

	if locChanged && (m.PlaceSrc == SrcManual || m.PlaceSrc == SrcBatch) {
		locKeywords, labels := m.UpdateLocation()

		m.AddLabels(labels)

		w := txt.UniqueWords(txt.Words(details.Keywords))
		w = append(w, locKeywords...)

		details.Keywords = strings.Join(txt.UniqueWords(w), ", ")
	}

	if err := m.UpdateLabels(); err != nil {
		log.Errorf("photo: %s %s while updating labels", m.String(), err)
	}

	if err := m.GenerateTitle(m.ClassifyLabels()); err != nil {
		log.Info(err)
	}

	if err := m.IndexKeywords(); err != nil {
		log.Errorf("photo: %s %s while indexing keywords", m.String(), err.Error())
	}

	edited := Now()
	m.EditedAt = &edited
	m.PhotoQuality = m.QualityScore()

	if err := m.Save(); err != nil {
		return err
	}

	// Update precalculated photo and file counts.
	UpdateCountsAsync()

	return nil
}

// FindPhoto looks up a Photo by UID or numeric ID and preloads key associations used by higher layers.
func FindPhoto(find Photo) *Photo {
	if find.PhotoUID == "" && find.ID == 0 {
		return nil
	}

	m := Photo{}

	// Preload related entities if a matching record is found.
	stmt := UnscopedDb().
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC")
		}).
		Preload("Labels.Label").
		Preload("Camera").
		Preload("Lens").
		Preload("Details").
		Preload("Place").
		Preload("Cell").
		Preload("Cell.Place")

	// Find photo by uid.
	if rnd.IsUID(find.PhotoUID, PhotoUID) {
		if stmt.First(&m, "photo_uid = ?", find.PhotoUID).Error == nil {
			return &m
		}
	}

	// Find photo by id.
	if find.ID > 0 {
		if stmt.First(&m, "id = ?", find.ID).Error == nil {
			return &m
		}
	}

	return nil
}

// PhotoLogString returns a sanitized identifier for logging that prefers
// photo name, falling back to original name, UID, or numeric ID.
func PhotoLogString(photoPath, photoName, originalName, photoUID string, id uint) string {
	if photoName != "" {
		return clean.Log(path.Join(photoPath, photoName))
	} else if originalName != "" {
		return clean.Log(originalName)
	} else if photoUID != "" {
		return "uid " + clean.Log(photoUID)
	} else if id > 0 {
		return fmt.Sprintf("id %d", id)
	}

	return "*Photo"
}

// String returns the id or name as string for logging purposes.
func (m *Photo) String() string {
	if m == nil {
		return "Photo<nil>"
	}

	return PhotoLogString(m.PhotoPath, m.PhotoName, m.OriginalName, m.PhotoUID, m.ID)
}

// FirstOrCreate inserts the Photo if it does not exist and otherwise reloads the persisted row with its associations.
func (m *Photo) FirstOrCreate() *Photo {
	if err := m.Create(); err == nil {
		return m
	} else {
		log.Tracef("photo: %s in %s (create)", err, m.String())
	}

	return FindPhoto(*m)
}

// Create persists a new Photo while holding the package mutex and ensures the related Details record exists.
func (m *Photo) Create() error {
	photoMutex.Lock()
	defer photoMutex.Unlock()

	if err := UnscopedDb().Create(m).Error; err != nil {
		return err
	}

	if err := m.SaveDetails(); err != nil {
		return err
	}

	return nil
}

// Save writes Photo changes, creates missing rows, and re-resolves the primary file relationship.
func (m *Photo) Save() error {
	photoMutex.Lock()
	defer photoMutex.Unlock()

	if err := Save(m, "ID", "PhotoUID"); err != nil {
		return err
	}

	if err := m.SaveDetails(); err != nil {
		return err
	}

	return m.ResolvePrimary()
}

// Update a column in the database.
func (m *Photo) Update(attr string, value interface{}) error {
	if m == nil {
		return errors.New("photo must not be nil - you may have found a bug")
	} else if !m.HasID() {
		return errors.New("photo ID must not be empty - you may have found a bug")
	}

	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Updates multiple columns in the database.
func (m *Photo) Updates(values interface{}) error {
	if values == nil {
		return nil
	} else if m == nil {
		return errors.New("photo must not be nil - you may have found a bug")
	} else if !m.HasID() {
		return errors.New("photo ID must not be empty - you may have found a bug")
	}

	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// GetID returns the numeric entity ID.
func (m *Photo) GetID() uint {
	return m.ID
}

// HasID checks if the photo has an id and uid assigned to it.
func (m *Photo) HasID() bool {
	if m == nil {
		return false
	}

	return m.ID > 0 && m.HasUID()
}

// HasUID checks if the photo has a valid UID.
func (m *Photo) HasUID() bool {
	if m == nil {
		return false
	}

	return rnd.IsUID(m.PhotoUID, PhotoUID)
}

// GetUID returns the unique entity id.
func (m *Photo) GetUID() string {
	if m == nil {
		return "<nil>"
	}

	return m.PhotoUID
}

// MediaType returns the current PhotoType as media.Type.
func (m *Photo) MediaType() media.Type {
	return media.Type(m.PhotoType)
}

// ResetMediaType resets the media type and source to the defaults.
func (m *Photo) ResetMediaType(resetSrc string) {
	if m.PhotoType != "" && SrcPriority[m.TypeSrc] > SrcPriority[resetSrc] {
		return
	}

	m.PhotoType = MediaImage
	m.TypeSrc = SrcAuto
}

// ResetDuration sets the video duration to 0.
func (m *Photo) ResetDuration() {
	m.PhotoDuration = 0
}

// HasMediaType checks if the photo has any of the specified media types.
func (m *Photo) HasMediaType(types ...media.Type) bool {
	mediaType := m.MediaType()

	for _, t := range types {
		if mediaType == t {
			return true
		}
	}

	return false
}

// SetMediaType sets a new media type if its priority is higher than that of the current type.
func (m *Photo) SetMediaType(newType media.Type, typeSrc string) {
	// Only allow a new main media type to be set.
	if !newType.IsMain() || newType.Equal(m.PhotoType) {
		return
	}

	// Get current media type.
	currentType := m.MediaType()

	// Do not change the type if the source priority is lower than the current one.
	if SrcPriority[typeSrc] < SrcPriority[m.TypeSrc] && currentType.IsMain() {
		return
	}

	// Do not automatically change a higher priority type to a lower one.
	if SrcPriority[typeSrc] <= SrcPriority[SrcFile] && media.Priority[newType] < media.Priority[currentType] {
		return
	}

	// Set new type and type source.
	m.PhotoType = newType.String()
	m.TypeSrc = typeSrc

	// Write a debug log containing the old and new media type.
	log.Debugf("photo: changed type of %s from %s to %s", m.String(), currentType.String(), newType.String())

	return
}

// Find fetches the matching record.
func (m *Photo) Find() *Photo {
	return FindPhoto(*m)
}

// SaveLabels recalculates derived metadata after label edits, persists the Photo, and schedules count updates.
func (m *Photo) SaveLabels() error {
	if !m.HasID() {
		return errors.New("photo: cannot save to database, id is empty")
	}

	labels := m.ClassifyLabels()

	m.UpdateDateFields()

	if err := m.GenerateTitle(labels); err != nil {
		log.Info(err)
	}

	if err := m.IndexKeywords(); err != nil {
		log.Errorf("photo: %s", err.Error())
	}

	m.PhotoQuality = m.QualityScore()

	if err := m.Save(); err != nil {
		return err
	}

	// Update precalculated photo and file counts.
	UpdateCountsAsync()

	return nil
}

// LabelKeywords converts the photo labels (and their categories) into
// keyword tokens that should be indexable for full‑text search. When the
// relation has not been preloaded yet, it fetches the labels transparently
// so callers always receive the same output.
func (m *Photo) LabelKeywords() (result []string) {
	if m == nil {
		return nil
	}

	if m.Labels == nil {
		m.PreloadLabels()
	}

	for _, l := range m.Labels {
		if l.Label == nil {
			continue
		}

		if l.Uncertainty >= 100 || list.Contains(labelKeywordsSkipSrc, l.LabelSrc) {
			continue
		}

		result = append(result, txt.Keywords(l.Label.LabelName)...)

		for _, c := range l.Label.LabelCategories {
			if c == nil {
				continue
			}
			result = append(result, txt.Keywords(c.LabelName)...)
		}
	}

	return result
}

// ClassifyLabels converts attached PhotoLabel relations into classify.Labels for downstream AI components.
func (m *Photo) ClassifyLabels() classify.Labels {
	result := classify.Labels{}

	for _, l := range m.Labels {
		if l.Label == nil {
			log.Warnf("photo: empty reference while creating classify labels (%d -> %d)", l.PhotoID, l.LabelID)
			continue
		}

		result = append(result, l.ClassifyLabel())
	}

	return result
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Photo) BeforeCreate(scope *gorm.Scope) error {
	if m.TakenAt.IsZero() || m.TakenAtLocal.IsZero() {
		now := Now()

		if err := scope.SetColumn("TakenAt", now); err != nil {
			return err
		}

		if err := scope.SetColumn("TakenAtLocal", now); err != nil {
			return err
		}
	}

	if rnd.IsUnique(m.PhotoUID, PhotoUID) {
		return nil
	}

	m.PhotoUID = rnd.GenerateUID(PhotoUID)

	return scope.SetColumn("PhotoUID", m.PhotoUID)
}

// BeforeSave ensures the existence of TakenAt properties before indexing or updating a photo.
func (m *Photo) BeforeSave(scope *gorm.Scope) error {
	if m.TakenAt.IsZero() || m.TakenAtLocal.IsZero() {
		now := Now()

		if err := scope.SetColumn("TakenAt", now); err != nil {
			return err
		}

		if err := scope.SetColumn("TakenAtLocal", now); err != nil {
			return err
		}
	}

	return nil
}

// RemoveKeyword removes a word from photo keywords.
func (m *Photo) RemoveKeyword(w string) error {
	details := m.GetDetails()

	words := txt.RemoveFromWords(txt.Words(details.Keywords), w)
	details.Keywords = strings.Join(words, ", ")

	return nil
}

// DropKeywords removes the specified keywords from the photo details and then persists them.
func (m *Photo) DropKeywords(remove []string) error {
	if m == nil || len(remove) == 0 {
		return nil
	}

	details := m.GetDetails()

	original := details.Keywords

	words := txt.Words(details.Keywords)

	for _, w := range remove {
		if w != "" {
			words = txt.RemoveFromWords(words, w)
		}
	}

	details.Keywords = strings.Join(words, ", ")

	// No update required.
	if details.Keywords == original {
		return nil
	}

	return details.Updates(Values{"keywords": details.Keywords})
}

// UpdateLabels refreshes automatically generated labels derived from the title, caption, subject metadata, and keywords.
func (m *Photo) UpdateLabels() error {
	if err := m.UpdateTitleLabels(); err != nil {
		return err
	}

	if err := m.UpdateCaptionLabels(); err != nil {
		return err
	}

	if err := m.UpdateSubjectLabels(); err != nil {
		return err
	}

	if err := m.UpdateKeywordLabels(); err != nil {
		return err
	}

	return nil
}

// SubjectNames returns all known subject names.
func (m *Photo) SubjectNames() []string {
	if f, err := m.PrimaryFile(); err == nil {
		return f.SubjectNames()
	}

	return nil
}

// SubjectKeywords returns keywords for all known subject names.
func (m *Photo) SubjectKeywords() []string {
	return txt.Words(strings.Join(m.SubjectNames(), " "))
}

// UpdateSubjectLabels updates the labels assigned based on photo subject metadata.
func (m *Photo) UpdateSubjectLabels() error {
	details := m.GetDetails()

	if details == nil {
		return nil
	} else if details.Subject == "" {
		return nil
	} else if SrcPriority[details.SubjectSrc] < SrcPriority[SrcMeta] {
		return nil
	}

	keywords := txt.UniqueKeywords(details.Subject)

	var labelIds []uint

	for _, w := range keywords {
		if label, err := FindLabel(w, true); err == nil {
			if label.Skip() {
				continue
			}

			labelIds = append(labelIds, label.ID)
			FirstOrCreatePhotoLabel(NewPhotoLabel(m.ID, label.ID, 20, classify.SrcSubject))
		}
	}

	return Db().Where("label_src = ? AND photo_id = ? AND label_id NOT IN (?)", classify.SrcSubject, m.ID, labelIds).Delete(&PhotoLabel{}).Error
}

// UpdateKeywordLabels updates the labels assigned based on photo keyword metadata.
func (m *Photo) UpdateKeywordLabels() error {
	details := m.GetDetails()

	if details == nil {
		return nil
	} else if details.Keywords == "" {
		return nil
	}

	keywords := txt.UniqueKeywords(details.Keywords)

	var labelIds []uint

	for _, w := range keywords {
		if label, err := FindLabel(w, true); err == nil {
			if label.Skip() {
				continue
			}

			labelIds = append(labelIds, label.ID)
			FirstOrCreatePhotoLabel(NewPhotoLabel(m.ID, label.ID, 25, classify.SrcKeyword))
		}
	}

	return Db().Where("label_src = ? AND photo_id = ? AND label_id NOT IN (?)", classify.SrcKeyword, m.ID, labelIds).Delete(&PhotoLabel{}).Error
}

// IndexKeywords synchronizes the photo-keyword join table based on normalized keywords from titles, captions, and metadata.
func (m *Photo) IndexKeywords() error {
	db := UnscopedDb()
	details := m.GetDetails()

	var keywordIds []uint
	var keywords []string

	// Extract keywords from title, caption, and other sources.
	keywords = append(keywords, txt.Keywords(m.GetTitle())...)
	keywords = append(keywords, txt.Keywords(m.GetCaption())...)
	keywords = append(keywords, m.SubjectKeywords()...)
	keywords = append(keywords, txt.Words(details.Keywords)...)
	keywords = append(keywords, m.LabelKeywords()...)
	keywords = append(keywords, txt.Keywords(details.Subject)...)
	keywords = append(keywords, txt.Keywords(details.Artist)...)

	keywords = txt.UniqueWords(keywords)

	for _, w := range keywords {
		kw := FirstOrCreateKeyword(NewKeyword(w))

		if kw == nil {
			log.Errorf("index keyword must not be nil - you may have found a bug")
			continue
		}

		if kw.Skip {
			continue
		}

		keywordIds = append(keywordIds, kw.ID)

		FirstOrCreatePhotoKeyword(NewPhotoKeyword(m.ID, kw.ID))
	}

	return db.Where("photo_id = ? AND keyword_id NOT IN (?)", m.ID, keywordIds).Delete(&PhotoKeyword{}).Error
}

// PreloadFiles loads the non-deleted file records associated with the photo.
func (m *Photo) PreloadFiles() *Photo {
	q := Db().
		Table("files").
		Select("files.*").
		Where("files.photo_id = ? AND files.deleted_at IS NULL", m.ID).
		Order("files.file_name DESC")

	Log("photo", "preload files", q.Scan(&m.Files).Error)

	return m
}

// PreloadKeywords loads keyword entities linked to the photo.
func (m *Photo) PreloadKeywords() *Photo {
	q := Db().NewScope(nil).DB().
		Table("keywords").
		Select(`keywords.*`).
		Joins("JOIN photos_keywords pk ON pk.keyword_id = keywords.id AND pk.photo_id = ?", m.ID).
		Order("keywords.keyword ASC")

	Log("photo", "preload files", q.Scan(&m.Keywords).Error)

	return m
}

// PreloadAlbums loads albums related to the photo using the standard visibility filters.
func (m *Photo) PreloadAlbums() *Photo {
	q := Db().NewScope(nil).DB().
		Table("albums").
		Select(`albums.*`).
		Joins("JOIN photos_albums pa ON pa.album_uid = albums.album_uid AND pa.photo_uid = ? AND pa.hidden = 0", m.PhotoUID).
		Where("albums.deleted_at IS NULL").
		Order("albums.album_title ASC")

	Log("photo", "preload albums", q.Scan(&m.Albums).Error)

	return m
}

// PreloadLabels loads labels related to the photo from the database. It is a
// no-op when the Photo pointer is nil or the record has not been persisted yet
// so call sites can invoke it defensively before reading `m.Labels`.
func (m *Photo) PreloadLabels() *Photo {
	if m == nil {
		return m
	} else if !m.HasID() {
		return m
	}

	Log("photo", "preload labels", Db().Model(PhotoLabel{}).Preload("Label").Where("photo_id = ?", m.ID).
		Order("photos_labels.uncertainty ASC, photos_labels.label_id DESC").Find(&m.Labels).Error)

	return m
}

// PreloadMany loads the primary supporting associations (files, keywords, albums).
func (m *Photo) PreloadMany() *Photo {
	m.PreloadFiles()
	m.PreloadKeywords()
	m.PreloadAlbums()

	return m
}

// NormalizeValues updates the model values with the values from deprecated fields, if any.
func (m *Photo) NormalizeValues() (normalized bool) {
	if m.PhotoCaption == "" && m.PhotoDescription != "" {
		m.PhotoCaption = m.PhotoDescription
		m.CaptionSrc = m.DescriptionSrc
		m.PhotoDescription = ""
		m.DescriptionSrc = ""
		normalized = true
	}

	if timeZone := tz.Name(m.TimeZone); timeZone != m.TimeZone {
		m.TimeZone = timeZone
		normalized = true
	}

	return normalized
}

// NoCameraSerial reports whether the photo has no camera serial assigned.
func (m *Photo) NoCameraSerial() bool {
	return m.CameraSerial == ""
}

// UnknownCamera tests whether the camera reference is the placeholder entry.
func (m *Photo) UnknownCamera() bool {
	return m.CameraID == 0 || m.CameraID == UnknownCamera.ID
}

// UnknownLens tests whether the lens reference is the placeholder entry.
func (m *Photo) UnknownLens() bool {
	return m.LensID == 0 || m.LensID == UnknownLens.ID
}

// GetDetails loads or lazily creates the Details record backing optional photo metadata.
func (m *Photo) GetDetails() *Details {
	if m.Details != nil {
		m.Details.PhotoID = m.ID
		return m.Details
	} else if !m.HasID() {
		m.Details = &Details{}
		return m.Details
	}

	m.Details = &Details{PhotoID: m.ID}

	if details := FirstOrCreateDetails(m.Details); details != nil {
		m.Details = details
	}

	return m.Details
}

// SaveDetails writes photo details to the database.
func (m *Photo) SaveDetails() error {
	if err := m.GetDetails().Save(); err == nil {
		return nil
	} else if details := FirstOrCreateDetails(m.GetDetails()); details != nil {
		m.Details = details
		return nil
	} else {
		log.Errorf("photo: %s (save details for %d)", err, m.ID)
		return err
	}
}

// ShouldGenerateLabels reports whether automatic vision labels should be generated for the photo.
// It allows regeneration when forced, when no labels exist, or when only manual/high-uncertainty
// labels are present so low-confidence results do not block improved predictions.
func (m *Photo) ShouldGenerateLabels(force bool) bool {
	// Return true if force is set or there are no labels yet.
	if len(m.Labels) == 0 || force {
		return true
	}

	// Check if any of the existing labels were generated using a vision model.
	for _, l := range m.Labels {
		if l.Uncertainty >= 100 {
			continue
		}

		if SrcGenerated[l.LabelSrc] > 0 {
			return false
		} else if l.LabelSrc == SrcCaption && SrcGenerated[m.CaptionSrc] > 0 {
			return false
		}
	}

	return true
}

// AddLabels ensures classify labels exist as Label entities and attaches them to the photo.
// Labels are skipped when they have no usable title or carry 0% probability so that UpdateClassify
// never receives invalid input from upstream detectors.
func (m *Photo) AddLabels(labels classify.Labels) {
	for _, classifyLabel := range labels {
		title := classifyLabel.Title()

		if title == "" || txt.Slug(title) == "" {
			log.Debugf("index: skipping blank label (%s)", m)
			continue
		}

		if classifyLabel.Uncertainty >= 100 {
			log.Debugf("index: skipping label %s with zero probability (%s)", title, m)
			continue
		}

		labelEntity := FirstOrCreateLabel(NewLabel(title, classifyLabel.Priority))

		if labelEntity == nil {
			log.Errorf("index: label %s could not be created (%s)", clean.Log(title), m)
			continue
		}

		if labelEntity.Deleted() {
			log.Debugf("index: skipping deleted label %s (%s)", clean.Log(title), m)
			continue
		}

		if err := labelEntity.UpdateClassify(classifyLabel); err != nil {
			log.Errorf("index: failed to update label %s (%s)", clean.Log(title), err)
		}

		labelSrc := classifyLabel.Source

		if labelSrc == SrcAuto {
			labelSrc = SrcImage
		} else {
			labelSrc = clean.ShortTypeLower(labelSrc)
		}

		template := NewPhotoLabel(m.ID, labelEntity.ID, classifyLabel.Uncertainty, labelSrc)
		template.Topicality = classifyLabel.Topicality
		score := 0

		if classifyLabel.NSFWConfidence > 0 {
			score = classifyLabel.NSFWConfidence
		}

		if classifyLabel.NSFW && score == 0 {
			score = 100
		}

		if score > 100 {
			score = 100
		}

		template.NSFW = score
		photoLabel := FirstOrCreatePhotoLabel(template)

		if photoLabel == nil {
			log.Errorf("index: photo-label %d must not be nil - you may have found a bug (%s)", labelEntity.ID, m)
			continue
		}

		if photoLabel.HasID() {
			updates := Values{}

			if photoLabel.Uncertainty > classifyLabel.Uncertainty && photoLabel.Uncertainty < 100 {
				updates["Uncertainty"] = classifyLabel.Uncertainty
				updates["LabelSrc"] = labelSrc
			}

			if classifyLabel.Topicality > 0 && photoLabel.Topicality != classifyLabel.Topicality {
				updates["Topicality"] = classifyLabel.Topicality
			}

			if classifyLabel.NSFWConfidence > 0 || classifyLabel.NSFW {
				nsfwScore := 0
				if classifyLabel.NSFWConfidence > 0 {
					nsfwScore = classifyLabel.NSFWConfidence
				}
				if classifyLabel.NSFW && nsfwScore == 0 {
					nsfwScore = 100
				}
				if nsfwScore > 100 {
					nsfwScore = 100
				}
				if photoLabel.NSFW != nsfwScore {
					updates["NSFW"] = nsfwScore
				}
			}

			if len(updates) > 0 {
				if err := photoLabel.Updates(updates); err != nil {
					log.Errorf("index: %s", err)
				}
			}
		}
	}

	Db().Set("gorm:auto_preload", true).Model(m).Related(&m.Labels)
}

// SetCamera updates the camera reference if the source priority allows the change.
func (m *Photo) SetCamera(camera *Camera, source string) {
	if camera == nil {
		log.Warnf("photo: %s failed to update camera from source %s", m.String(), SrcString(source))
		return
	}

	if camera.Unknown() {
		return
	}

	if SrcPriority[source] < SrcPriority[m.CameraSrc] && !m.UnknownCamera() {
		return
	}

	m.CameraID = camera.ID
	m.Camera = camera
	m.CameraSrc = source

	if !m.PhotoScan && m.Camera.Scanner() {
		m.PhotoScan = true
	}
}

// SetLens updates the lens reference when the source outranks the existing metadata.
func (m *Photo) SetLens(lens *Lens, source string) {
	if lens == nil {
		log.Warnf("photo: %s failed to update lens from source %s", m.String(), SrcString(source))
		return
	}

	if lens.Unknown() {
		return
	}

	if SrcPriority[source] < SrcPriority[m.CameraSrc] && !m.UnknownLens() {
		return
	}

	m.LensID = lens.ID
	m.Lens = lens
}

// SetExposure updates the photo exposure details.
func (m *Photo) SetExposure(focalLength int, fNumber float32, iso int, exposure, source string) {
	hasPriority := SrcPriority[source] >= SrcPriority[m.CameraSrc]

	// Set focal length.
	if focalLength > 0 && focalLength <= 128000 && (hasPriority || m.PhotoFocalLength <= 0) {
		m.PhotoFocalLength = focalLength
	}

	// Set F number.
	if fNumber > 0 && fNumber <= 256 && (hasPriority || m.PhotoFNumber <= 0) {
		m.PhotoFNumber = fNumber
	}

	// Set ISO number.
	if iso > 0 && iso <= 128000 && (hasPriority || m.PhotoIso <= 0) {
		m.PhotoIso = iso
	}

	// Set exposure time.
	if exposure != "" && (hasPriority || m.PhotoExposure == "") {
		m.PhotoExposure = exposure
	}
}

// AllFilesMissing reports whether all files for this photo are marked missing.
func (m *Photo) AllFilesMissing() bool {
	count := 0

	if err := Db().Model(&File{}).
		Where("photo_id = ? AND file_missing = 0", m.ID).
		Count(&count).Error; err != nil {
		log.Error(err)
	}

	return count == 0
}

// AllFiles returns all files of this photo.
func (m *Photo) AllFiles() (files Files) {
	if err := UnscopedDb().Where("photo_id = ?", m.ID).Find(&files).Error; err != nil {
		log.Error(err)
	}

	return files
}

// Archive removes the photo from albums and flags it as archived (soft delete).
func (m *Photo) Archive() error {
	if !m.HasID() {
		return fmt.Errorf("photo has no id")
	} else if m.DeletedAt != nil {
		return nil
	}

	deletedAt := Now()

	if err := Db().Model(&PhotoAlbum{}).Where("photo_uid = ?", m.PhotoUID).UpdateColumn("hidden", true).Error; err != nil {
		return err
	} else if err = m.Update("deleted_at", deletedAt); err != nil {
		return err
	}

	m.DeletedAt = &deletedAt

	return nil
}

// Restore removes the photo from the archive (reverses soft delete).
func (m *Photo) Restore() error {
	if !m.HasID() {
		return fmt.Errorf("photo has no id")
	} else if m.DeletedAt == nil {
		return nil
	}

	if err := m.Update("deleted_at", gorm.Expr("NULL")); err != nil {
		return err
	}

	m.DeletedAt = nil

	return nil
}

// Delete deletes the photo from the index.
func (m *Photo) Delete(permanently bool) (files Files, err error) {
	if !m.HasID() {
		return files, fmt.Errorf("invalid photo id %d / uid %s", m.ID, clean.Log(m.PhotoUID))
	}

	if permanently {
		return m.DeletePermanently()
	}

	files = m.AllFiles()

	for _, file := range files {
		if err = file.Delete(false); err != nil {
			log.Errorf("index: %s (remove file)", err)
		}
	}

	m.DeletedAt = TimeStamp()
	m.PhotoQuality = -1

	return files, m.Updates(Values{"deleted_at": *m.DeletedAt, "photo_quality": m.PhotoQuality})
}

// DeletePermanently permanently removes a photo from the index.
func (m *Photo) DeletePermanently() (files Files, err error) {
	if m.ID < 1 || m.PhotoUID == "" {
		return files, fmt.Errorf("invalid photo id %d / uid %s", m.ID, clean.Log(m.PhotoUID))
	}

	files = m.AllFiles()

	for _, file := range files {
		if logErr := file.DeletePermanently(); logErr != nil {
			log.Errorf("index: %s (remove file)", logErr)
		}
	}

	if logErr := UnscopedDb().Delete(Details{}, "photo_id = ?", m.ID).Error; logErr != nil {
		log.Errorf("index: %s (remove details)", logErr)
	}

	if logErr := UnscopedDb().Delete(PhotoKeyword{}, "photo_id = ?", m.ID).Error; logErr != nil {
		log.Errorf("index: %s (remove keywords)", logErr)
	}

	if logErr := UnscopedDb().Delete(PhotoLabel{}, "photo_id = ?", m.ID).Error; logErr != nil {
		log.Errorf("index: %s (remove labels)", logErr)
	}

	if logErr := UnscopedDb().Delete(PhotoAlbum{}, "photo_uid = ?", m.PhotoUID).Error; logErr != nil {
		log.Errorf("index: %s (remove albums)", logErr)
	}

	return files, UnscopedDb().Delete(m).Error
}

// React adds or updates a user reaction.
func (m *Photo) React(user *User, reaction react.Emoji) error {
	if user == nil {
		return fmt.Errorf("unknown user")
	}

	if reaction.Unknown() {
		return m.UnReact(user)
	}

	return NewReaction(m.PhotoUID, user.GetUID()).React(reaction).Save()
}

// UnReact deletes a previous user reaction, if any.
func (m *Photo) UnReact(user *User) error {
	if user == nil {
		return fmt.Errorf("unknown user")
	}

	if r := FindReaction(m.PhotoUID, user.GetUID()); r != nil {
		return r.Delete()
	}

	return nil
}

// SetFavorite updates the favorite flag of a photo.
func (m *Photo) SetFavorite(favorite bool) error {
	changed := m.PhotoFavorite != favorite
	m.PhotoFavorite = favorite
	m.PhotoQuality = m.QualityScore()

	if err := m.Updates(Values{"photo_favorite": m.PhotoFavorite, "photo_quality": m.PhotoQuality}); err != nil {
		return err
	}

	// Update counters if changed and not deleted.
	if changed && m.PhotoPrivate == false && m.DeletedAt == nil {
		if favorite {
			event.Publish("count.favorites", event.Data{
				"count": 1,
			})
		} else {
			event.Publish("count.favorites", event.Data{
				"count": -1,
			})
		}
	}

	return nil
}

// SetStack updates the stack flag of a photo.
func (m *Photo) SetStack(stack int8) {
	if m.PhotoStack != stack {
		m.PhotoStack = stack
		Log("photo", "update stack flag", m.Update("photo_stack", m.PhotoStack))
	}
}

// Approved checks if the photo is not in review.
func (m *Photo) Approved() bool {
	if !m.HasID() {
		return false
	} else if m.PhotoQuality >= 3 || m.PhotoType != MediaImage || m.EditedAt != nil {
		return true
	}

	return false
}

// Approve approves the photo if it is in review.
func (m *Photo) Approve() error {
	if !m.HasID() {
		return fmt.Errorf("photo has no id")
	} else if m.PhotoQuality >= 3 {
		// Nothing to do.
		return nil
	}

	// Restore photo if archived.
	if err := m.Restore(); err != nil {
		return err
	}

	edited := Now()
	m.EditedAt = &edited
	m.PhotoQuality = m.QualityScore()

	if err := Db().Model(m).Updates(Photo{EditedAt: m.EditedAt, PhotoQuality: m.PhotoQuality}).Error; err != nil {
		return err
	}

	// Update precalculated photo and file counts.
	UpdateCountsAsync()

	event.Publish("count.review", event.Data{
		"count": -1,
	})

	return nil
}

// Links returns all share links for this entity.
func (m *Photo) Links() Links {
	return FindLinks("", m.PhotoUID)
}

// PrimaryFile returns the primary file for this photo.
func (m *Photo) PrimaryFile() (*File, error) {
	return PrimaryFile(m.PhotoUID)
}

// SetPrimary sets a new primary file.
func (m *Photo) SetPrimary(fileUid string) (err error) {
	if m.PhotoUID == "" {
		return fmt.Errorf("photo uid is empty")
	}

	var files []string

	if fileUid != "" {
		// Do nothing.
	} else if err = Db().Model(File{}).
		Where("photo_uid = ? AND file_type IN (?) AND file_missing = 0 AND file_error = ''", m.PhotoUID, media.PreviewExpr).
		Order("file_width DESC, file_hdr DESC").Limit(1).
		Pluck("file_uid", &files).Error; err != nil {
		return err
	} else if len(files) == 0 {
		return fmt.Errorf("found no preview image for %s", clean.Log(m.PhotoUID))
	} else {
		fileUid = files[0]
	}

	if fileUid == "" {
		return fmt.Errorf("file uid is empty")
	}

	if err = Db().Model(File{}).
		Where("photo_uid = ? AND file_uid <> ?", m.PhotoUID, fileUid).
		UpdateColumn("file_primary", 0).Error; err != nil {
		return err
	} else if err = Db().Model(File{}).Where("photo_uid = ? AND file_uid = ?", m.PhotoUID, fileUid).
		UpdateColumn("file_primary", 1).Error; err != nil {
		return err
	} else if m.PhotoQuality < 0 {
		m.PhotoQuality = 0
		err = m.UpdateQuality()
	}

	// Regenerate file search index.
	File{PhotoID: m.ID, PhotoUID: m.PhotoUID}.RegenerateIndex()

	return nil
}

// MapKey returns a key referencing time and location for indexing.
func (m *Photo) MapKey() string {
	return MapKey(m.TakenAt, m.CellID)
}

// SetCameraSerial updates the camera serial number.
func (m *Photo) SetCameraSerial(s string) {
	if s = txt.Clip(s, txt.ClipDefault); m.NoCameraSerial() && s != "" {
		m.CameraSerial = s
	}
}

// FaceCount returns the current number of faces on the primary picture.
func (m *Photo) FaceCount() int {
	if f, err := m.PrimaryFile(); err != nil {
		return 0
	} else {
		return f.ValidFaceCount()
	}
}

// Indexed returns the immutable timestamp recorded when the photo completed indexing.
// It automatically initializes the timestamp when missing so workers can rely on it even if CheckedAt resets.
func (m *Photo) Indexed() *time.Time {
	if m == nil {
		return nil
	} else if m.IndexedAt == nil {
		m.IndexedAt = TimeStamp()
	} else if m.IndexedAt.IsZero() {
		m.IndexedAt = TimeStamp()
	}

	return m.IndexedAt
}

// IsNewlyIndexed reports whether the photo still awaits its first indexing timestamp while not being deleted.
func (m *Photo) IsNewlyIndexed() bool {
	if m == nil {
		return false
	} else if m.IndexedAt == nil {
		return !m.IsDeleted()
	} else if m.IndexedAt.IsZero() {
		return !m.IsDeleted()
	}

	return false
}

// IsDeleted returns true if the photo was deleted.
func (m *Photo) IsDeleted() bool {
	if m == nil {
		return true
	} else if m.DeletedAt == nil {
		return false
	} else if m.DeletedAt.IsZero() {
		return false
	}

	return true
}
