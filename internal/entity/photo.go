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

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/react"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	PhotoUID = byte('p')
)

var IndexUpdateInterval = 3 * time.Hour           // 3 Hours
var MetadataUpdateInterval = 24 * 3 * time.Hour   // 3 Days
var MetadataEstimateInterval = 24 * 7 * time.Hour // 7 Days

var photoMutex = sync.Mutex{}

type Photos []Photo

// UIDs returns a slice of photo UIDs.
func (m Photos) UIDs() []string {
	result := make([]string, len(m))

	for i, el := range m {
		result[i] = el.PhotoUID
	}

	return result
}

// MapKey returns a key referencing time and location for indexing.
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
	PhotoDescription string        `gorm:"type:VARCHAR(4096);" json:"Description" yaml:"Description,omitempty"`
	DescriptionSrc   string        `gorm:"type:VARBINARY(8);" json:"DescriptionSrc" yaml:"DescriptionSrc,omitempty"`
	PhotoPath        string        `gorm:"type:VARBINARY(1024);index:idx_photos_path_name;" json:"Path" yaml:"-"`
	PhotoName        string        `gorm:"type:VARBINARY(255);index:idx_photos_path_name;" json:"Name" yaml:"-"`
	OriginalName     string        `gorm:"type:VARBINARY(755);" json:"OriginalName" yaml:"OriginalName,omitempty"`
	PhotoStack       int8          `json:"Stack" yaml:"Stack,omitempty"`
	PhotoFavorite    bool          `json:"Favorite" yaml:"Favorite,omitempty"`
	PhotoPrivate     bool          `json:"Private" yaml:"Private,omitempty"`
	PhotoScan        bool          `json:"Scan" yaml:"Scan,omitempty"`
	PhotoPanorama    bool          `json:"Panorama" yaml:"Panorama,omitempty"`
	TimeZone         string        `gorm:"type:VARBINARY(64);" json:"TimeZone" yaml:"TimeZone,omitempty"`
	PlaceID          string        `gorm:"type:VARBINARY(42);index;default:'zz'" json:"PlaceID" yaml:"-"`
	PlaceSrc         string        `gorm:"type:VARBINARY(8);" json:"PlaceSrc" yaml:"PlaceSrc,omitempty"`
	CellID           string        `gorm:"type:VARBINARY(42);index;default:'zz'" json:"CellID" yaml:"-"`
	CellAccuracy     int           `json:"CellAccuracy" yaml:"CellAccuracy,omitempty"`
	PhotoAltitude    int           `json:"Altitude" yaml:"Altitude,omitempty"`
	PhotoLat         float32       `gorm:"type:FLOAT;index;" json:"Lat" yaml:"Lat,omitempty"`
	PhotoLng         float32       `gorm:"type:FLOAT;index;" json:"Lng" yaml:"Lng,omitempty"`
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
	CreatedAt        time.Time     `yaml:"CreatedAt,omitempty"`
	UpdatedAt        time.Time     `yaml:"UpdatedAt,omitempty"`
	EditedAt         *time.Time    `yaml:"EditedAt,omitempty"`
	PublishedAt      *time.Time    `sql:"index" json:"PublishedAt,omitempty" yaml:"PublishedAt,omitempty"`
	CheckedAt        *time.Time    `sql:"index" yaml:"-"`
	EstimatedAt      *time.Time    `json:"EstimatedAt,omitempty" yaml:"-"`
	DeletedAt        *time.Time    `sql:"index" yaml:"DeletedAt,omitempty"`
}

// TableName returns the entity table name.
func (Photo) TableName() string {
	return "photos"
}

// NewPhoto creates a new photo with default values.
func NewPhoto(stackable bool) Photo {
	return NewUserPhoto(stackable, "")
}

// NewUserPhoto creates a photo owned by a user.
func NewUserPhoto(stackable bool, userUid string) Photo {
	m := Photo{
		PhotoTitle:   UnknownTitle,
		PhotoType:    MediaImage,
		PhotoCountry: UnknownCountry.ID,
		CameraID:     UnknownCamera.ID,
		LensID:       UnknownLens.ID,
		CellID:       UnknownLocation.ID,
		PlaceID:      UnknownPlace.ID,
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

// SavePhotoForm saves a model in the database using form data.
func SavePhotoForm(model Photo, form form.Photo) error {
	locChanged := model.PhotoLat != form.PhotoLat || model.PhotoLng != form.PhotoLng || model.PhotoCountry != form.PhotoCountry

	if err := deepcopier.Copy(&model).From(form); err != nil {
		return err
	}

	if !model.HasID() {
		return errors.New("cannot save form when photo id is missing")
	}

	// Update time fields.
	if model.TimeZoneUTC() {
		model.TakenAtLocal = model.TakenAt
	} else {
		model.TakenAt = model.GetTakenAt()
	}

	model.UpdateDateFields()

	details := model.GetDetails()

	if form.Details.PhotoID == model.ID {
		if err := deepcopier.Copy(details).From(form.Details); err != nil {
			return err
		}

		details.Keywords = strings.Join(txt.UniqueWords(txt.Words(details.Keywords)), ", ")
	}

	if locChanged && model.PlaceSrc == SrcManual {
		locKeywords, labels := model.UpdateLocation()

		model.AddLabels(labels)

		w := txt.UniqueWords(txt.Words(details.Keywords))
		w = append(w, locKeywords...)

		details.Keywords = strings.Join(txt.UniqueWords(w), ", ")
	}

	if err := model.SyncKeywordLabels(); err != nil {
		log.Errorf("photo: %s %s while syncing keywords and labels", model.String(), err)
	}

	if err := model.UpdateTitle(model.ClassifyLabels()); err != nil {
		log.Info(err)
	}

	if err := model.IndexKeywords(); err != nil {
		log.Errorf("photo: %s %s while indexing keywords", model.String(), err.Error())
	}

	edited := TimeStamp()
	model.EditedAt = &edited
	model.PhotoQuality = model.QualityScore()

	if err := model.Save(); err != nil {
		return err
	}

	// Update precalculated photo and file counts.
	if err := UpdateCounts(); err != nil {
		log.Warnf("index: %s (update counts)", err)
	}

	return nil
}

// String returns the id or name as string.
func (m *Photo) String() string {
	if m.PhotoName != "" {
		return clean.Log(path.Join(m.PhotoPath, m.PhotoName))
	} else if m.OriginalName != "" {
		return clean.Log(m.OriginalName)
	} else if m.PhotoUID != "" {
		return "uid " + clean.Log(m.PhotoUID)
	} else if m.ID > 0 {
		return fmt.Sprintf("id %d", m.ID)
	}

	return "(unknown)"
}

// FirstOrCreate fetches an existing row from the database or inserts a new one.
func (m *Photo) FirstOrCreate() *Photo {
	if err := m.Create(); err == nil {
		return m
	} else {
		log.Tracef("photo: %s in %s (create)", err, m.String())
	}

	return FindPhoto(*m)
}

// Create inserts a new photo to the database.
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

// Save updates the record in the database or inserts a new record if it does not already exist.
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

// FindPhoto fetches the matching record or returns null if it was not found.
func FindPhoto(find Photo) *Photo {
	if find.PhotoUID == "" && find.ID == 0 {
		return nil
	}

	m := Photo{}

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

	// Search for UID.
	if rnd.IsUID(find.PhotoUID, PhotoUID) {
		if stmt.First(&m, "photo_uid = ?", find.PhotoUID).Error == nil {
			return &m
		}
	}

	// Search for ID.
	if find.ID > 0 {
		if stmt.First(&m, "id = ?", find.ID).Error == nil {
			return &m
		}
	}

	return nil
}

// Find fetches the matching record.
func (m *Photo) Find() *Photo {
	return FindPhoto(*m)
}

// SaveLabels updates the photo after labels have changed.
func (m *Photo) SaveLabels() error {
	if !m.HasID() {
		return errors.New("photo: cannot save to database, id is empty")
	}

	labels := m.ClassifyLabels()

	m.UpdateDateFields()

	if err := m.UpdateTitle(labels); err != nil {
		log.Info(err)
	}

	details := m.GetDetails()

	w := txt.UniqueWords(txt.Words(details.Keywords))
	w = append(w, labels.Keywords()...)
	details.Keywords = strings.Join(txt.UniqueWords(w), ", ")

	if err := m.IndexKeywords(); err != nil {
		log.Errorf("photo: %s", err.Error())
	}

	m.PhotoQuality = m.QualityScore()

	if err := m.Save(); err != nil {
		return err
	}

	// Update precalculated photo and file counts.
	if err := UpdateCounts(); err != nil {
		log.Warnf("index: %s (update counts)", err)
	}

	return nil
}

// ClassifyLabels returns all associated labels as classify.Labels
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
		now := TimeStamp()

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

// BeforeSave ensures the existence of TakenAt properties before indexing or updating a photo
func (m *Photo) BeforeSave(scope *gorm.Scope) error {
	if m.TakenAt.IsZero() || m.TakenAtLocal.IsZero() {
		now := TimeStamp()

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

// SyncKeywordLabels maintains the label / photo relationship for existing labels and keywords.
func (m *Photo) SyncKeywordLabels() error {
	details := m.GetDetails()
	keywords := txt.UniqueKeywords(details.Keywords)

	var labelIds []uint

	for _, w := range keywords {
		if label := FindLabel(w); label != nil {
			if label.Deleted() {
				continue
			}

			labelIds = append(labelIds, label.ID)
			FirstOrCreatePhotoLabel(NewPhotoLabel(m.ID, label.ID, 25, classify.SrcKeyword))
		}
	}

	return Db().Where("label_src = ? AND photo_id = ? AND label_id NOT IN (?)", classify.SrcKeyword, m.ID, labelIds).Delete(&PhotoLabel{}).Error
}

// IndexKeywords adds given keywords to the photo entry
func (m *Photo) IndexKeywords() error {
	db := UnscopedDb()
	details := m.GetDetails()

	var keywordIds []uint
	var keywords []string

	// Add title, description and other keywords
	keywords = append(keywords, txt.Keywords(m.PhotoTitle)...)
	keywords = append(keywords, txt.Keywords(m.PhotoDescription)...)
	keywords = append(keywords, m.SubjectKeywords()...)
	keywords = append(keywords, txt.Words(details.Keywords)...)
	keywords = append(keywords, txt.Keywords(details.Subject)...)
	keywords = append(keywords, txt.Keywords(details.Artist)...)

	keywords = txt.UniqueWords(keywords)

	for _, w := range keywords {
		kw := FirstOrCreateKeyword(NewKeyword(w))

		if kw == nil {
			log.Errorf("index keyword should not be nil - possible bug")
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

// PreloadFiles prepares gorm scope to retrieve photo file
func (m *Photo) PreloadFiles() {
	q := Db().
		Table("files").
		Select("files.*").
		Where("files.photo_id = ? AND files.deleted_at IS NULL", m.ID).
		Order("files.file_name DESC")

	Log("photo", "preload files", q.Scan(&m.Files).Error)
}

// PreloadKeywords prepares gorm scope to retrieve photo keywords
func (m *Photo) PreloadKeywords() {
	q := Db().NewScope(nil).DB().
		Table("keywords").
		Select(`keywords.*`).
		Joins("JOIN photos_keywords pk ON pk.keyword_id = keywords.id AND pk.photo_id = ?", m.ID).
		Order("keywords.keyword ASC")

	Log("photo", "preload files", q.Scan(&m.Keywords).Error)
}

// PreloadAlbums prepares gorm scope to retrieve photo albums
func (m *Photo) PreloadAlbums() {
	q := Db().NewScope(nil).DB().
		Table("albums").
		Select(`albums.*`).
		Joins("JOIN photos_albums pa ON pa.album_uid = albums.album_uid AND pa.photo_uid = ? AND pa.hidden = 0", m.PhotoUID).
		Where("albums.deleted_at IS NULL").
		Order("albums.album_title ASC")

	Log("photo", "preload albums", q.Scan(&m.Albums).Error)
}

// PreloadMany prepares gorm scope to retrieve photo file, albums and keywords
func (m *Photo) PreloadMany() {
	m.PreloadFiles()
	m.PreloadKeywords()
	m.PreloadAlbums()
}

// HasID tests if the photo has a database id and uid.
func (m *Photo) HasID() bool {
	return m.ID > 0 && m.PhotoUID != ""
}

// NoCameraSerial checks if the photo has no CameraSerial
func (m *Photo) NoCameraSerial() bool {
	return m.CameraSerial == ""
}

// UnknownCamera test if the camera is unknown.
func (m *Photo) UnknownCamera() bool {
	return m.CameraID == 0 || m.CameraID == UnknownCamera.ID
}

// UnknownLens test if the lens is unknown.
func (m *Photo) UnknownLens() bool {
	return m.LensID == 0 || m.LensID == UnknownLens.ID
}

// HasDescription checks if the photo has a description.
func (m *Photo) HasDescription() bool {
	return m.PhotoDescription != ""
}

// GetDetails returns the photo description details.
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

// AddLabels updates the entity with additional or updated label information.
func (m *Photo) AddLabels(labels classify.Labels) {
	for _, classifyLabel := range labels {
		labelEntity := FirstOrCreateLabel(NewLabel(classifyLabel.Title(), classifyLabel.Priority))

		if labelEntity == nil {
			log.Errorf("index: label %s should not be nil - possible bug (%s)", clean.Log(classifyLabel.Title()), m)
			continue
		}

		if labelEntity.Deleted() {
			log.Debugf("index: skipping deleted label %s (%s)", clean.Log(classifyLabel.Title()), m)
			continue
		}

		if err := labelEntity.UpdateClassify(classifyLabel); err != nil {
			log.Errorf("index: failed updating label %s (%s)", clean.Log(classifyLabel.Title()), err)
		}

		photoLabel := FirstOrCreatePhotoLabel(NewPhotoLabel(m.ID, labelEntity.ID, classifyLabel.Uncertainty, classifyLabel.Source))

		if photoLabel == nil {
			log.Errorf("index: photo-label %d should not be nil - possible bug (%s)", labelEntity.ID, m)
			continue
		}

		if photoLabel.Uncertainty > classifyLabel.Uncertainty && photoLabel.Uncertainty < 100 {
			if err := photoLabel.Updates(map[string]interface{}{
				"Uncertainty": classifyLabel.Uncertainty,
				"LabelSrc":    classifyLabel.Source,
			}); err != nil {
				log.Errorf("index: %s", err)
			}
		}
	}

	Db().Set("gorm:auto_preload", true).Model(m).Related(&m.Labels)
}

// SetDescription changes the photo description if not empty and from the same source.
func (m *Photo) SetDescription(desc, source string) {
	newDesc := txt.Clip(desc, txt.ClipLongText)

	if newDesc == "" {
		return
	}

	if (SrcPriority[source] < SrcPriority[m.DescriptionSrc]) && m.HasDescription() {
		return
	}

	m.PhotoDescription = newDesc
	m.DescriptionSrc = source
}

// SetCamera updates the camera.
func (m *Photo) SetCamera(camera *Camera, source string) {
	if camera == nil {
		log.Warnf("photo: %s failed updating camera from source %s", m.String(), SrcString(source))
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

// SetLens updates the lens.
func (m *Photo) SetLens(lens *Lens, source string) {
	if lens == nil {
		log.Warnf("photo: %s failed updating lens from source %s", m.String(), SrcString(source))
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

	if focalLength > 0 && (hasPriority || m.PhotoFocalLength <= 0) {
		m.PhotoFocalLength = focalLength
	}

	if fNumber > 0 && (hasPriority || m.PhotoFNumber <= 0) {
		m.PhotoFNumber = fNumber
	}

	if iso > 0 && (hasPriority || m.PhotoIso <= 0) {
		m.PhotoIso = iso
	}

	if exposure != "" && (hasPriority || m.PhotoExposure == "") {
		m.PhotoExposure = exposure
	}
}

// AllFilesMissing returns true, if all files for this photo are missing.
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
	deletedAt := TimeStamp()

	if err := Db().Model(&PhotoAlbum{}).Where("photo_uid = ?", m.PhotoUID).UpdateColumn("hidden", true).Error; err != nil {
		return err
	} else if err := m.Update("deleted_at", deletedAt); err != nil {
		return err
	}

	m.DeletedAt = &deletedAt

	return nil
}

// Restore removes the archive flag (undo soft delete).
func (m *Photo) Restore() error {
	if err := m.Update("deleted_at", gorm.Expr("NULL")); err != nil {
		return err
	}

	m.DeletedAt = nil

	return nil
}

// Delete deletes the photo from the index.
func (m *Photo) Delete(permanently bool) (files Files, err error) {
	if m.ID < 1 || m.PhotoUID == "" {
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

	return files, m.Updates(map[string]interface{}{"DeletedAt": TimeStamp(), "PhotoQuality": -1})
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

// NoDescription returns true if the photo has no description.
func (m *Photo) NoDescription() bool {
	return m.PhotoDescription == ""
}

// Update a column in the database.
func (m *Photo) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Updates multiple columns in the database.
func (m *Photo) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// React adds or updates a user reaction.
func (m *Photo) React(user *User, reaction react.Emoji) error {
	if user == nil {
		return fmt.Errorf("unknown user")
	}

	if reaction.Unknown() {
		return m.UnReact(user)
	}

	return NewReaction(m.PhotoUID, user.UID()).React(reaction).Save()
}

// UnReact deletes a previous user reaction, if any.
func (m *Photo) UnReact(user *User) error {
	if user == nil {
		return fmt.Errorf("unknown user")
	}

	if r := FindReaction(m.PhotoUID, user.UID()); r != nil {
		return r.Delete()
	}

	return nil
}

// SetFavorite updates the favorite flag of a photo.
func (m *Photo) SetFavorite(favorite bool) error {
	changed := m.PhotoFavorite != favorite
	m.PhotoFavorite = favorite
	m.PhotoQuality = m.QualityScore()

	if err := m.Updates(map[string]interface{}{"PhotoFavorite": m.PhotoFavorite, "PhotoQuality": m.PhotoQuality}); err != nil {
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
		Log("photo", "update stack flag", m.Update("PhotoStack", m.PhotoStack))
	}
}

// Approve approves a photo in review.
func (m *Photo) Approve() error {
	if m.PhotoQuality >= 3 {
		// Nothing to do.
		return nil
	}

	edited := TimeStamp()
	m.EditedAt = &edited
	m.PhotoQuality = m.QualityScore()

	if err := Db().Model(m).Updates(Photo{EditedAt: m.EditedAt, PhotoQuality: m.PhotoQuality}).Error; err != nil {
		return err
	}

	// Update precalculated photo and file counts.
	if err := UpdateCounts(); err != nil {
		log.Warnf("index: %s (update counts)", err)
	}

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
