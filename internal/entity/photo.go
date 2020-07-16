package entity

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/ulule/deepcopier"
)

type Photos []Photo

// UIDs returns a slice of photo UIDs.
func (m Photos) UIDs() []string {
	result := make([]string, len(m))

	for i, el := range m {
		result[i] = el.PhotoUID
	}

	return result
}

// Photo represents a photo, all its properties, and link to all its images and sidecar files.
type Photo struct {
	ID               uint         `gorm:"primary_key" yaml:"-"`
	UUID             string       `gorm:"type:varbinary(42);index;" json:"DocumentID,omitempty" yaml:"DocumentID,omitempty"`
	TakenAt          time.Time    `gorm:"type:datetime;index:idx_photos_taken_uid;" json:"TakenAt" yaml:"TakenAt"`
	TakenAtLocal     time.Time    `gorm:"type:datetime;" yaml:"-"`
	TakenSrc         string       `gorm:"type:varbinary(8);" json:"TakenSrc" yaml:"TakenSrc,omitempty"`
	PhotoUID         string       `gorm:"type:varbinary(42);unique_index;index:idx_photos_taken_uid;" json:"UID" yaml:"UID"`
	PhotoType        string       `gorm:"type:varbinary(8);default:'image';" json:"Type" yaml:"Type"`
	TypeSrc          string       `gorm:"type:varbinary(8);" json:"TypeSrc" yaml:"TypeSrc,omitempty"`
	PhotoTitle       string       `gorm:"type:varchar(255);" json:"Title" yaml:"Title"`
	TitleSrc         string       `gorm:"type:varbinary(8);" json:"TitleSrc" yaml:"TitleSrc,omitempty"`
	PhotoDescription string       `gorm:"type:text;" json:"Description" yaml:"Description,omitempty"`
	DescriptionSrc   string       `gorm:"type:varbinary(8);" json:"DescriptionSrc" yaml:"DescriptionSrc,omitempty"`
	PhotoPath        string       `gorm:"type:varbinary(768);index;" json:"Path" yaml:"-"`
	PhotoName        string       `gorm:"type:varbinary(255);" json:"Name" yaml:"-"`
	OriginalName     string       `gorm:"type:varbinary(768);" json:"OriginalName" yaml:"OriginalName,omitempty"`
	PhotoFavorite    bool         `json:"Favorite" yaml:"Favorite,omitempty"`
	PhotoPrivate     bool         `json:"Private" yaml:"Private,omitempty"`
	PhotoScan        bool         `json:"Scan" yaml:"Scan,omitempty"`
	PhotoPanorama    bool         `json:"Panorama" yaml:"Panorama,omitempty"`
	TimeZone         string       `gorm:"type:varbinary(64);" json:"TimeZone" yaml:"-"`
	PlaceID          string       `gorm:"type:varbinary(42);index;default:'zz'" json:"PlaceID" yaml:"-"`
	PlaceSrc         string       `gorm:"type:varbinary(8);" json:"PlaceSrc" yaml:"PlaceSrc,omitempty"`
	CellID           string       `gorm:"type:varbinary(42);index;default:'zz'" json:"CellID" yaml:"-"`
	CellAccuracy     int          `json:"CellAccuracy" yaml:"CellAccuracy,omitempty"`
	PhotoAltitude    int          `json:"Altitude" yaml:"Altitude,omitempty"`
	PhotoLat         float32      `gorm:"type:FLOAT;index;" json:"Lat" yaml:"Lat,omitempty"`
	PhotoLng         float32      `gorm:"type:FLOAT;index;" json:"Lng" yaml:"Lng,omitempty"`
	PhotoCountry     string       `gorm:"type:varbinary(2);index:idx_photos_country_year_month;default:'zz'" json:"Country" yaml:"-"`
	PhotoYear        int          `gorm:"index:idx_photos_country_year_month;" json:"Year" yaml:"Year"`
	PhotoMonth       int          `gorm:"index:idx_photos_country_year_month;" json:"Month" yaml:"Month"`
	PhotoDay         int          `json:"Day" yaml:"Day"`
	PhotoIso         int          `json:"Iso" yaml:"ISO,omitempty"`
	PhotoExposure    string       `gorm:"type:varbinary(64);" json:"Exposure" yaml:"Exposure,omitempty"`
	PhotoFNumber     float32      `gorm:"type:FLOAT;" json:"FNumber" yaml:"FNumber,omitempty"`
	PhotoFocalLength int          `json:"FocalLength" yaml:"FocalLength,omitempty"`
	PhotoQuality     int          `gorm:"type:SMALLINT" json:"Quality" yaml:"-"`
	PhotoResolution  int          `gorm:"type:SMALLINT" json:"Resolution" yaml:"-"`
	CameraID         uint         `gorm:"index:idx_photos_camera_lens;default:1" json:"CameraID" yaml:"-"`
	CameraSerial     string       `gorm:"type:varbinary(255);" json:"CameraSerial" yaml:"CameraSerial,omitempty"`
	CameraSrc        string       `gorm:"type:varbinary(8);" json:"CameraSrc" yaml:"-"`
	LensID           uint         `gorm:"index:idx_photos_camera_lens;default:1" json:"LensID" yaml:"-"`
	Details          *Details     `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Details" yaml:"Details"`
	Camera           *Camera      `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Camera" yaml:"-"`
	Lens             *Lens        `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Lens" yaml:"-"`
	Cell             *Cell        `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Cell" yaml:"-"`
	Place            *Place       `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false" json:"Place" yaml:"-"`
	Keywords         []Keyword    `json:"-" yaml:"-"`
	Albums           []Album      `json:"-" yaml:"-"`
	Files            []File       `yaml:"-"`
	Labels           []PhotoLabel `yaml:"-"`
	CreatedAt        time.Time    `yaml:"CreatedAt,omitempty"`
	UpdatedAt        time.Time    `yaml:"UpdatedAt,omitempty"`
	EditedAt         *time.Time   `yaml:"EditedAt,omitempty"`
	CheckedAt        *time.Time   `sql:"index" yaml:"-"`
	DeletedAt        *time.Time   `sql:"index" yaml:"DeletedAt,omitempty"`
}

// NewPhoto creates a photo entity.
func NewPhoto() Photo {
	return Photo{
		PhotoType:    TypeImage,
		PhotoCountry: UnknownCountry.ID,
		CameraID:     UnknownCamera.ID,
		LensID:       UnknownLens.ID,
		CellID:       UnknownLocation.ID,
		PlaceID:      UnknownPlace.ID,
		Camera:       &UnknownCamera,
		Lens:         &UnknownLens,
		Cell:         &UnknownLocation,
		Place:        &UnknownPlace,
	}
}

// SavePhotoForm saves a model in the database using form data.
func SavePhotoForm(model Photo, form form.Photo, geoApi string) error {
	locChanged := model.PhotoLat != form.PhotoLat || model.PhotoLng != form.PhotoLng || model.PhotoCountry != form.PhotoCountry

	if err := deepcopier.Copy(&model).From(form); err != nil {
		return err
	}

	if !model.HasID() {
		return errors.New("photo: can't save form, id is empty")
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
		locKeywords, labels := model.UpdateLocation(geoApi)

		model.AddLabels(labels)

		w := txt.UniqueWords(txt.Words(details.Keywords))
		w = append(w, locKeywords...)

		details.Keywords = strings.Join(txt.UniqueWords(w), ", ")
	}

	if err := model.SyncKeywordLabels(); err != nil {
		log.Errorf("photo: %s", err)
	}

	if err := model.UpdateTitle(model.ClassifyLabels()); err != nil {
		log.Info(err)
	}

	if err := model.IndexKeywords(); err != nil {
		log.Errorf("photo: %s", err.Error())
	}

	edited := Timestamp()
	model.EditedAt = &edited
	model.PhotoQuality = model.QualityScore()

	if err := model.Save(); err != nil {
		return err
	}

	if err := UpdatePhotoCounts(); err != nil {
		log.Errorf("photo: %s", err)
	}

	return nil
}

// String returns the id or name as string.
func (m *Photo) String() string {
	if m.PhotoUID == "" {
		if m.PhotoName != "" {
			return txt.Quote(m.PhotoName)
		} else if m.OriginalName != "" {
			return txt.Quote(m.OriginalName)
		}

		return "(unknown)"
	}

	return "uid " + txt.Quote(m.PhotoUID)
}

// FirstOrCreate fetches an existing row from the database or inserts a new one.
func (m *Photo) FirstOrCreate() error {
	if err := m.Create(); err == nil {
		return nil
	} else if err := m.Find(); err != nil {
		return fmt.Errorf("photo: %s (first or create %s)", err, m.String())
	}

	return nil
}

// Create inserts a new photo to the database.
func (m *Photo) Create() error {
	if err := UnscopedDb().Create(m).Error; err != nil {
		return err
	}

	if err := m.SaveDetails(); err != nil {
		return err
	}

	return nil
}

// Save updates an existing photo or inserts a new one.
func (m *Photo) Save() error {
	if err := UnscopedDb().Save(m).Error; err == nil {
		// Nothing to do.
	} else if !strings.Contains(strings.ToLower(err.Error()), "lock") {
		log.Errorf("photo: %s (save %s)", err, m.PhotoUID)
		return err
	} else if err := UnscopedDb().Save(m).Error; err != nil {
		log.Errorf("photo: %s (save %s after deadlock)", err, m.PhotoUID)
		return err
	}

	if err := m.SaveDetails(); err != nil {
		return err
	}

	return nil
}

// Find returns a photo from the database.
func (m *Photo) Find() error {
	if m.PhotoUID == "" && m.ID == 0 {
		return fmt.Errorf("photo: id and uid must not be empty (find)")
	}

	q := UnscopedDb().
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

	if rnd.IsPPID(m.PhotoUID, 'p') {
		if err := q.First(m, "photo_uid = ?", m.PhotoUID).Error; err != nil {
			return err
		}
	} else if err := q.First(m, "id = ?", m.ID).Error; err != nil {
		return err
	}

	return nil
}

// Save the photo to the database.
func (m *Photo) SaveLabels() error {
	if !m.HasID() {
		return errors.New("photo: can't save to database, id is empty")
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

	if err := UpdatePhotoCounts(); err != nil {
		log.Errorf("photo: %s", err)
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
		now := Timestamp()

		if err := scope.SetColumn("TakenAt", now); err != nil {
			return err
		}

		if err := scope.SetColumn("TakenAtLocal", now); err != nil {
			return err
		}
	}

	if rnd.IsUID(m.PhotoUID, 'p') {
		return nil
	}

	return scope.SetColumn("PhotoUID", rnd.PPID('p'))
}

// BeforeSave ensures the existence of TakenAt properties before indexing or updating a photo
func (m *Photo) BeforeSave(scope *gorm.Scope) error {
	if m.TakenAt.IsZero() || m.TakenAtLocal.IsZero() {
		now := Timestamp()

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
	keywords = append(keywords, txt.Keywords(details.Keywords)...)
	keywords = append(keywords, txt.Keywords(details.Subject)...)
	keywords = append(keywords, txt.Keywords(details.Artist)...)

	keywords = txt.UniqueWords(keywords)

	for _, w := range keywords {
		kw := FirstOrCreateKeyword(NewKeyword(w))

		if kw == nil {
			log.Errorf("index keyword should not be nil - bug?")
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
	q := Db().NewScope(nil).DB().
		Table("files").
		Select(`files.*`).
		Where("files.photo_id = ?", m.ID).
		Order("files.file_name DESC")

	logError(q.Scan(&m.Files))
}

/* func (m *Photo) PreloadLabels() {
	q := Db().NewScope(nil).DB().
		Table("labels").
		Select(`labels.*`).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = ?", m.ID).
		Where("labels.deleted_at IS NULL").
		Order("labels.label_name ASC")

	logError(q.Scan(&m.Labels))
} */

// PreloadKeywords prepares gorm scope to retrieve photo keywords
func (m *Photo) PreloadKeywords() {
	q := Db().NewScope(nil).DB().
		Table("keywords").
		Select(`keywords.*`).
		Joins("JOIN photos_keywords ON photos_keywords.keyword_id = keywords.id AND photos_keywords.photo_id = ?", m.ID).
		Order("keywords.keyword ASC")

	logError(q.Scan(&m.Keywords))
}

// PreloadAlbums prepares gorm scope to retrieve photo albums
func (m *Photo) PreloadAlbums() {
	q := Db().NewScope(nil).DB().
		Table("albums").
		Select(`albums.*`).
		Joins("JOIN photos_albums ON photos_albums.album_uid = albums.album_uid AND photos_albums.photo_uid = ?", m.PhotoUID).
		Where("albums.deleted_at IS NULL").
		Order("albums.album_title ASC")

	logError(q.Scan(&m.Albums))
}

// PreloadMany prepares gorm scope to retrieve photo file, albums and keywords
func (m *Photo) PreloadMany() {
	m.PreloadFiles()
	// m.PreloadLabels()
	m.PreloadKeywords()
	m.PreloadAlbums()
}

// HasID checks if the photo has a database id and uid.
func (m *Photo) HasID() bool {
	return m.ID > 0 && m.PhotoUID != ""
}

// UnknownLocation checks if the photo has an unknown location.
func (m *Photo) UnknownLocation() bool {
	return m.CellID == "" || m.CellID == UnknownLocation.ID
}

// HasLocation checks if the photo has a known location.
func (m *Photo) HasLocation() bool {
	return !m.UnknownLocation()
}

// LocationLoaded checks if the photo has a known location that is currently loaded.
func (m *Photo) LocationLoaded() bool {
	if m.Cell == nil {
		return false
	}

	if m.Cell.Place == nil {
		return false
	}

	return !m.Cell.Unknown() && m.Cell.ID == m.CellID
}

// LoadLocation loads the photo location from the database if not done already.
func (m *Photo) LoadLocation() error {
	if m.LocationLoaded() {
		return nil
	}

	if m.UnknownLocation() {
		return fmt.Errorf("photo: unknown location (%s)", m)
	}

	var location Cell

	err := Db().Preload("Place").First(&location, "id = ?", m.CellID).Error

	if err != nil {
		return err
	}

	if location.Place == nil {
		location.Place = &UnknownPlace
		location.PlaceID = UnknownPlace.ID
	}

	m.Cell = &location

	return nil
}

// PlaceLoaded checks if the photo has a known place that is currently loaded.
func (m *Photo) PlaceLoaded() bool {
	if m.Place == nil {
		return false
	}

	return !m.Place.Unknown() && m.Place.ID == m.PlaceID
}

// LoadPlace loads the photo place from the database if not done already.
func (m *Photo) LoadPlace() error {
	if m.PlaceLoaded() {
		return nil
	}

	if m.UnknownPlace() {
		return fmt.Errorf("photo: unknown place (%s)", m)
	}

	var place Place

	err := Db().First(&place, "id = ?", m.PlaceID).Error

	if err != nil {
		return err
	}

	m.Place = &place

	return nil
}

// HasLatLng checks if the photo has a latitude and longitude.
func (m *Photo) HasLatLng() bool {
	return m.PhotoLat != 0.0 || m.PhotoLng != 0.0
}

// NoLatLng checks if latitude and longitude are missing.
func (m *Photo) NoLatLng() bool {
	return !m.HasLatLng()
}

// UnknownPlace checks if the photo has an unknown place.
func (m *Photo) UnknownPlace() bool {
	return m.PlaceID == "" || m.PlaceID == UnknownPlace.ID
}

// HasPlace checks if the photo has a known place.
func (m *Photo) HasPlace() bool {
	return !m.UnknownPlace()
}

// HasCountry checks if the photo has a known country.
func (m *Photo) HasCountry() bool {
	return !m.UnknownCountry()
}

// UnknownCountry checks if the photo has an unknown country.
func (m *Photo) UnknownCountry() bool {
	return m.CountryCode() == UnknownCountry.ID
}

// NoTitle checks if the photo has no Title
func (m *Photo) NoTitle() bool {
	return m.PhotoTitle == ""
}

// NoCameraSerial checks if the photo has no CameraSerial
func (m *Photo) NoCameraSerial() bool {
	return m.CameraSerial == ""
}

// HasTitle checks if the photo has a title.
func (m *Photo) HasTitle() bool {
	return m.PhotoTitle != ""
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

// FileTitle returns a photo title based on the file name and/or path.
func (m *Photo) FileTitle() string {
	if !fs.IsGenerated(m.PhotoName) {
		if title := txt.FileTitle(m.PhotoName); title != "" {
			return title
		}
	}

	if m.OriginalName != "" && !fs.IsGenerated(m.OriginalName) {
		if title := txt.FileTitle(m.OriginalName); title != "" {
			return title
		} else if title := txt.FileTitle(path.Dir(m.OriginalName)); title != "" {
			return title
		}
	}

	if m.PhotoPath != "" {
		return txt.FileTitle(m.PhotoPath)
	}

	return ""
}

// UpdateTitle updated the photo title based on location and labels.
func (m *Photo) UpdateTitle(labels classify.Labels) error {
	if m.TitleSrc != SrcAuto && m.HasTitle() {
		return fmt.Errorf("photo: won't update title, %s was modified", m.PhotoUID)
	}

	var knownLocation bool

	oldTitle := m.PhotoTitle
	fileTitle := m.FileTitle()

	if m.LocationLoaded() {
		knownLocation = true
		loc := m.Cell

		// TODO: User defined title format
		if title := labels.Title(loc.Name()); title != "" {
			log.Debugf("photo: using label %s to create title for %s", txt.Quote(title), m.PhotoUID)
			if loc.NoCity() || loc.LongCity() || loc.CityContains(title) {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", txt.Title(title), loc.CountryName(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", txt.Title(title), loc.City(), m.TakenAt.Format("2006")), SrcAuto)
			}
		} else if loc.Name() != "" && loc.City() != "" {
			if len(loc.Name()) > 45 {
				m.SetTitle(txt.Title(loc.Name()), SrcAuto)
			} else if len(loc.Name()) > 20 || len(loc.City()) > 16 || strings.Contains(loc.Name(), loc.City()) {
				m.SetTitle(fmt.Sprintf("%s / %s", loc.Name(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", loc.Name(), loc.City(), m.TakenAt.Format("2006")), SrcAuto)
			}
		} else if loc.City() != "" && loc.CountryName() != "" {
			if len(loc.City()) > 20 {
				m.SetTitle(fmt.Sprintf("%s / %s", loc.City(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", loc.City(), loc.CountryName(), m.TakenAt.Format("2006")), SrcAuto)
			}
		}
	} else if m.PlaceLoaded() {
		knownLocation = true

		if title := labels.Title(fileTitle); title != "" {
			log.Debugf("photo: using label %s to create title for %s", txt.Quote(title), m.PhotoUID)
			if m.Place.NoCity() || m.Place.LongCity() || m.Place.CityContains(title) {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", txt.Title(title), m.Place.CountryName(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", txt.Title(title), m.Place.City(), m.TakenAt.Format("2006")), SrcAuto)
			}
		} else if m.Place.City() != "" && m.Place.CountryName() != "" {
			if len(m.Place.City()) > 20 {
				m.SetTitle(fmt.Sprintf("%s / %s", m.Place.City(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", m.Place.City(), m.Place.CountryName(), m.TakenAt.Format("2006")), SrcAuto)
			}
		}
	}

	if !knownLocation || m.NoTitle() {
		if fileTitle == "" && len(labels) > 0 && labels[0].Priority >= -1 && labels[0].Uncertainty <= 85 && labels[0].Name != "" {
			if m.TakenSrc != SrcAuto {
				m.SetTitle(fmt.Sprintf("%s / %s", txt.Title(labels[0].Name), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(txt.Title(labels[0].Name), SrcAuto)
			}
		} else if fileTitle != "" && len(fileTitle) <= 20 && !m.TakenAtLocal.IsZero() && m.TakenSrc != SrcAuto {
			m.SetTitle(fmt.Sprintf("%s / %s", fileTitle, m.TakenAtLocal.Format("2006")), SrcAuto)
		} else if fileTitle != "" {
			m.SetTitle(fileTitle, SrcAuto)
		} else {
			if m.TakenSrc != SrcAuto {
				m.SetTitle(fmt.Sprintf("%s / %s", TitleUnknown, m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(TitleUnknown, SrcAuto)
			}
		}
	}

	if m.PhotoTitle != oldTitle {
		log.Debugf("photo: changed title of %s to %s", m.PhotoUID, txt.Quote(m.PhotoTitle))
	}

	return nil
}

// AddLabels updates the entity with additional or updated label information.
func (m *Photo) AddLabels(labels classify.Labels) {
	for _, classifyLabel := range labels {
		labelEntity := FirstOrCreateLabel(NewLabel(classifyLabel.Title(), classifyLabel.Priority))

		if labelEntity == nil {
			log.Errorf("index: label %s should not be nil - bug? (%s)", txt.Quote(classifyLabel.Title()), m)
			continue
		}

		if labelEntity.Deleted() {
			log.Debugf("index: skipping deleted label %s (%s)", txt.Quote(classifyLabel.Title()), m)
			continue
		}

		if err := labelEntity.UpdateClassify(classifyLabel); err != nil {
			log.Errorf("index: %s", err)
		}

		photoLabel := FirstOrCreatePhotoLabel(NewPhotoLabel(m.ID, labelEntity.ID, classifyLabel.Uncertainty, classifyLabel.Source))

		if photoLabel == nil {
			log.Errorf("index: photo-label %d should not be nil - bug? (%s)", labelEntity.ID, m)
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

// SetTitle changes the photo title and clips it to 300 characters.
func (m *Photo) SetTitle(title, source string) {
	newTitle := txt.Clip(title, txt.ClipDefault)

	if newTitle == "" {
		return
	}

	if m.TitleSrc != SrcAuto && m.TitleSrc != source && source != SrcManual && m.HasTitle() {
		return
	}

	m.PhotoTitle = newTitle
	m.TitleSrc = source
}

// SetDescription changes the photo description if not empty and from the same source.
func (m *Photo) SetDescription(desc, source string) {
	newDesc := txt.Clip(desc, txt.ClipDescription)

	if newDesc == "" {
		return
	}

	if m.DescriptionSrc != SrcAuto && m.DescriptionSrc != source && source != SrcManual && m.PhotoDescription != "" {
		return
	}

	m.PhotoDescription = newDesc
	m.DescriptionSrc = source
}

// SetTakenAt changes the photo date if not empty and from the same source.
func (m *Photo) SetTakenAt(taken, local time.Time, zone, source string) {
	if taken.IsZero() || taken.Year() < 1000 {
		return
	}

	if m.TakenSrc != SrcAuto && m.TakenSrc != source && source != SrcManual {
		return
	}

	m.TakenAt = taken.Round(time.Second).UTC()
	m.TakenSrc = source

	if local.IsZero() || local.Year() < 1000 {
		m.TakenAtLocal = m.TakenAt
	} else {
		m.TakenAtLocal = local.Round(time.Second)
	}

	if zone != "" {
		m.TimeZone = zone
	}

	m.UpdateDateFields()
}

// UpdateDateFields updates internal date fields.
func (m *Photo) UpdateDateFields() {
	if m.TakenAt.IsZero() || m.TakenAt.Year() < 1000 {
		return
	}

	if m.TakenAtLocal.IsZero() || m.TakenAtLocal.Year() < 1000 {
		m.TakenAtLocal = m.TakenAt
	}

	if m.TakenSrc == SrcAuto {
		m.PhotoYear = YearUnknown
		m.PhotoMonth = MonthUnknown
		m.PhotoDay = DayUnknown
	} else if m.TakenSrc != SrcManual {
		m.PhotoYear = m.TakenAtLocal.Year()
		m.PhotoMonth = int(m.TakenAtLocal.Month())
		m.PhotoDay = m.TakenAtLocal.Day()
	}
}

// SetCoordinates changes the photo lat, lng and altitude if not empty and from the same source.
func (m *Photo) SetCoordinates(lat, lng float32, altitude int, source string) {
	if lat == 0.0 && lng == 0.0 {
		return
	}

	if m.PlaceSrc != SrcAuto && m.PlaceSrc != source && source != SrcManual {
		return
	}

	m.PhotoLat = lat
	m.PhotoLng = lng
	m.PhotoAltitude = altitude
	m.PlaceSrc = source
}

// AllFilesMissing returns true, if all files for this photo are missing.
func (m *Photo) AllFilesMissing() bool {
	count := 0

	if err := Db().Model(&File{}).
		Where("photo_id = ? AND b.file_missing = 0", m.ID).
		Count(&count).Error; err != nil {
		log.Error(err)
	}

	return count == 0
}

// Delete deletes the entity from the database.
func (m *Photo) Delete(permanently bool) error {
	if permanently {
		return m.DeletePermanently()
	}

	Db().Delete(File{}, "photo_id = ?", m.ID)

	return Db().Delete(m).Error
}

// Delete permanently deletes the entity from the database.
func (m *Photo) DeletePermanently() error {
	Db().Unscoped().Delete(File{}, "photo_id = ?", m.ID)
	Db().Unscoped().Delete(Details{}, "photo_id = ?", m.ID)
	Db().Unscoped().Delete(PhotoKeyword{}, "photo_id = ?", m.ID)
	Db().Unscoped().Delete(PhotoLabel{}, "photo_id = ?", m.ID)
	Db().Unscoped().Delete(PhotoAlbum{}, "photo_uid = ?", m.PhotoUID)

	return Db().Unscoped().Delete(m).Error
}

// NoDescription returns true if the photo has no description.
func (m *Photo) NoDescription() bool {
	return m.PhotoDescription == ""
}

// Updates a column in the database.
func (m *Photo) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Updates multiple columns in the database.
func (m *Photo) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// SetFavorite updates the favorite status of a photo.
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

// Approve approves a photo in review.
func (m *Photo) Approve() error {
	if m.PhotoQuality >= 3 {
		// Nothing to do.
		return nil
	}

	edited := Timestamp()
	m.EditedAt = &edited
	m.PhotoQuality = m.QualityScore()

	if err := Db().Model(m).Updates(Photo{EditedAt: m.EditedAt, PhotoQuality: m.PhotoQuality}).Error; err != nil {
		return err
	}

	if err := UpdatePhotoCounts(); err != nil {
		log.Errorf("photo: %s", err)
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
func (m *Photo) PrimaryFile() (File, error) {
	return PrimaryFile(m.PhotoUID)
}
