package entity

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/ulule/deepcopier"
)

// Photo represents a photo, all its properties, and link to all its images and sidecar files.
type Photo struct {
	ID               uint        `gorm:"primary_key"`
	PhotoUUID        string      `gorm:"type:varbinary(36);unique_index;index:idx_photos_taken_uuid;"`
	TakenAt          time.Time   `gorm:"type:datetime;index:idx_photos_taken_uuid;" json:"TakenAt"`
	TakenAtLocal     time.Time   `gorm:"type:datetime;"`
	TakenSrc         string      `gorm:"type:varbinary(8);" json:"TakenSrc"`
	PhotoTitle       string      `gorm:"type:varchar(255);" json:"PhotoTitle"`
	TitleSrc         string      `gorm:"type:varbinary(8);" json:"TitleSrc"`
	PhotoPath        string      `gorm:"type:varbinary(768);index;"`
	PhotoName        string      `gorm:"type:varbinary(255);"`
	PhotoQuality     int         `gorm:"type:SMALLINT" json:"PhotoQuality"`
	PhotoResolution  int         `gorm:"type:SMALLINT" json:"PhotoResolution"`
	PhotoFavorite    bool        `json:"PhotoFavorite"`
	PhotoPrivate     bool        `json:"PhotoPrivate"`
	PhotoVideo       bool        `json:"PhotoVideo"`
	PhotoLat         float32     `gorm:"type:FLOAT;index;" json:"PhotoLat"`
	PhotoLng         float32     `gorm:"type:FLOAT;index;" json:"PhotoLng"`
	PhotoAltitude    int         `json:"PhotoAltitude"`
	PhotoIso         int         `json:"PhotoIso"`
	PhotoFocalLength int         `json:"PhotoFocalLength"`
	PhotoFNumber     float32     `gorm:"type:FLOAT;" json:"PhotoFNumber"`
	PhotoExposure    string      `gorm:"type:varbinary(64);" json:"PhotoExposure"`
	CameraID         uint        `gorm:"index:idx_photos_camera_lens;" json:"CameraID"`
	CameraSerial     string      `gorm:"type:varbinary(255);" json:"CameraSerial"`
	CameraSrc        string      `gorm:"type:varbinary(8);" json:"CameraSrc"`
	LensID           uint        `gorm:"index:idx_photos_camera_lens;" json:"LensID"`
	PlaceID          string      `gorm:"type:varbinary(16);index;default:'zz'" json:"PlaceID"`
	LocationID       string      `gorm:"type:varbinary(16);index;" json:"LocationID"`
	LocationSrc      string      `gorm:"type:varbinary(8);" json:"LocationSrc"`
	TimeZone         string      `gorm:"type:varbinary(64);" json:"TimeZone"`
	PhotoCountry     string      `gorm:"type:varbinary(2);index:idx_photos_country_year_month;default:'zz'" json:"PhotoCountry"`
	PhotoYear        int         `gorm:"index:idx_photos_country_year_month;"`
	PhotoMonth       int         `gorm:"index:idx_photos_country_year_month;"`
	Description      Description `json:"Description"`
	DescriptionSrc   string      `gorm:"type:varbinary(8);" json:"DescriptionSrc"`
	Camera           *Camera     `json:"Camera"`
	Lens             *Lens       `json:"Lens"`
	Location         *Location   `json:"Location"`
	Place            *Place      `json:"-"`
	Links            []Link      `gorm:"foreignkey:ShareUUID;association_foreignkey:PhotoUUID"`
	Keywords         []Keyword   `json:"-"`
	Albums           []Album     `json:"-"`
	Files            []File
	Labels           []PhotoLabel
	CreatedAt        time.Time
	UpdatedAt        time.Time
	EditedAt         *time.Time
	DeletedAt        *time.Time `sql:"index"`
}

// SavePhotoForm updates a model using form data and persists it in the database.
func SavePhotoForm(model Photo, form form.Photo, geoApi string) error {
	db := Db()
	locChanged := model.PhotoLat != form.PhotoLat || model.PhotoLng != form.PhotoLng

	if err := deepcopier.Copy(&model).From(form); err != nil {
		return err
	}

	model.UpdateYearMonth()

	if form.Description.PhotoID == model.ID {
		if err := deepcopier.Copy(&model.Description).From(form.Description); err != nil {
			return err
		}

		model.Description.PhotoKeywords = strings.Join(txt.UniqueKeywords(model.Description.PhotoKeywords), ", ")
	}

	if model.HasLatLng() && locChanged && model.LocationSrc == SrcManual {
		locKeywords, labels := model.UpdateLocation(geoApi)

		model.AddLabels(labels)

		w := txt.UniqueKeywords(model.Description.PhotoKeywords)
		w = append(w, locKeywords...)

		model.Description.PhotoKeywords = strings.Join(txt.UniqueWords(w), ", ")
	}

	if err := model.UpdateTitle(model.ClassifyLabels()); err != nil {
		log.Warnf("%s (%s)", err.Error(), model.PhotoUUID)
	}

	if err := model.IndexKeywords(); err != nil {
		log.Warnf("%s (%s)", err.Error(), model.PhotoUUID)
	}

	edited := time.Now().UTC()
	model.EditedAt = &edited
	model.PhotoQuality = model.QualityScore()

	if err := db.Unscoped().Save(&model).Error; err != nil {
		return err
	}

	if err := UpdatePhotoCounts(); err != nil {
		log.Errorf("photo: %s", err)
	}

	return nil
}

// Save stored the entity in the database.
func (m *Photo) Save() error {
	db := Db()
	labels := m.ClassifyLabels()

	m.UpdateYearMonth()

	if err := m.UpdateTitle(labels); err != nil {
		log.Warnf("%s (%s)", err.Error(), m.PhotoUUID)
	}

	if m.DescriptionLoaded() {
		w := txt.UniqueKeywords(m.Description.PhotoKeywords)
		w = append(w, labels.Keywords()...)
		m.Description.PhotoKeywords = strings.Join(txt.UniqueWords(w), ", ")
	}

	if err := m.IndexKeywords(); err != nil {
		log.Error(err)
	}

	m.PhotoQuality = m.QualityScore()

	if err := db.Unscoped().Save(m).Error; err != nil {
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

// BeforeCreate computes a unique UUID, and set a default takenAt before indexing a new photo
func (m *Photo) BeforeCreate(scope *gorm.Scope) error {
	if m.TakenAt.IsZero() || m.TakenAtLocal.IsZero() {
		now := time.Now()

		if err := scope.SetColumn("TakenAt", now); err != nil {
			return err
		}

		if err := scope.SetColumn("TakenAtLocal", now); err != nil {
			return err
		}
	}

	if rnd.IsPPID(m.PhotoUUID, 'p') {
		return nil
	}

	return scope.SetColumn("PhotoUUID", rnd.PPID('p'))
}

// BeforeSave ensures the existence of TakenAt properties before indexing or updating a photo
func (m *Photo) BeforeSave(scope *gorm.Scope) error {
	if m.TakenAt.IsZero() || m.TakenAtLocal.IsZero() {
		now := time.Now()

		if err := scope.SetColumn("TakenAt", now); err != nil {
			return err
		}

		if err := scope.SetColumn("TakenAtLocal", now); err != nil {
			return err
		}
	}

	return nil
}

// IndexKeywords adds given keywords to the photo entry
func (m *Photo) IndexKeywords() error {
	if !m.DescriptionLoaded() {
		return fmt.Errorf("photo: can't index keywords, description not loaded (%s)", m.PhotoUUID)
	}

	db := Db()

	var keywordIds []uint
	var keywords []string

	// Add title, description and other keywords
	keywords = append(keywords, txt.Keywords(m.PhotoTitle)...)
	keywords = append(keywords, txt.Keywords(m.Description.PhotoDescription)...)
	keywords = append(keywords, txt.Keywords(m.Description.PhotoKeywords)...)
	keywords = append(keywords, txt.Keywords(m.Description.PhotoSubject)...)
	keywords = append(keywords, txt.Keywords(m.Description.PhotoArtist)...)

	keywords = txt.UniqueWords(keywords)

	for _, w := range keywords {
		kw := NewKeyword(w).FirstOrCreate()

		if kw.Skip {
			continue
		}

		keywordIds = append(keywordIds, kw.ID)

		NewPhotoKeyword(m.ID, kw.ID).FirstOrCreate()
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
		Joins("JOIN photos_albums ON photos_albums.album_uuid = albums.album_uuid AND photos_albums.photo_uuid = ?", m.PhotoUUID).
		Where("albums.deleted_at IS NULL").
		Order("albums.album_name ASC")

	logError(q.Scan(&m.Albums))
}

// PreloadMany prepares gorm scope to retrieve photo file, albums and keywords
func (m *Photo) PreloadMany() {
	m.PreloadFiles()
	// m.PreloadLabels()
	m.PreloadKeywords()
	m.PreloadAlbums()
}

// NoLocation checks if the photo has no location
func (m *Photo) NoLocation() bool {
	return m.LocationID == ""
}

// HasLocation checks if the photo has a location
func (m *Photo) HasLocation() bool {
	return m.LocationID != ""
}

// HasLatLng checks if the photo has a latitude and longitude.
func (m *Photo) HasLatLng() bool {
	return m.PhotoLat != 0 && m.PhotoLng != 0
}

// NoLatLng checks if latitude and longitude are missing.
func (m *Photo) NoLatLng() bool {
	return !m.HasLatLng()
}

// NoPlace checks if the photo has no Place
func (m *Photo) NoPlace() bool {
	return m.PlaceID == "" || m.PlaceID == UnknownPlace.ID
}

// HasPlace checks if the photo has a Place
func (m *Photo) HasPlace() bool {
	return !m.NoPlace()
}

// NoTitle checks if the photo has no Title
func (m *Photo) NoTitle() bool {
	return m.PhotoTitle == ""
}

// NoCameraSerial checks if the photo has no CameraSerial
func (m *Photo) NoCameraSerial() bool {
	return m.CameraSerial == ""
}

// HasTitle checks if the photo has a  Title
func (m *Photo) HasTitle() bool {
	return m.PhotoTitle != ""
}

// DescriptionLoaded returns true if photo description exists.
func (m *Photo) DescriptionLoaded() bool {
	return m.Description.PhotoID == m.ID
}

// UpdateTitle updated the photo title based on location and labels.
func (m *Photo) UpdateTitle(labels classify.Labels) error {
	if m.TitleSrc != SrcAuto && m.HasTitle() {
		return errors.New("photo: won't update title, was modified")
	}

	hasLocation := m.Location != nil && m.Location.Place != nil

	if hasLocation {
		loc := m.Location

		if title := labels.Title(loc.Name()); title != "" { // TODO: User defined title format
			log.Infof("photo: using label %s to create photo title", txt.Quote(title))
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
	}

	if !hasLocation || m.NoTitle() {
		if len(labels) > 0 && labels[0].Priority >= -1 && labels[0].Uncertainty <= 85 && labels[0].Name != "" {
			if m.TakenSrc != SrcAuto {
				m.SetTitle(fmt.Sprintf("%s / %s", txt.Title(labels[0].Name), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(txt.Title(labels[0].Name), SrcAuto)
			}
		} else if !m.TakenAtLocal.IsZero() && m.TakenSrc != SrcAuto {
			m.SetTitle(fmt.Sprintf("%s / %s", TitleUnknown, m.TakenAtLocal.Format("2006")), SrcAuto)
		} else {
			m.SetTitle(TitleUnknown, SrcAuto)
		}

		log.Infof("photo: changed photo title to %s", txt.Quote(m.PhotoTitle))
	} else {
		log.Infof("photo: new title is %s", txt.Quote(m.PhotoTitle))
	}

	return nil
}

// AddLabels updates the entity with additional or updated label information.
func (m *Photo) AddLabels(labels classify.Labels) {
	// TODO: Update classify labels from database
	for _, label := range labels {
		lm := NewLabel(label.Title(), label.Priority).FirstOrCreate()

		if lm.New {
			event.EntitiesCreated("labels", []*Label{lm})

			if label.Priority >= 0 {
				event.Publish("count.labels", event.Data{
					"count": 1,
				})
			}
		}

		if err := lm.Update(label); err != nil {
			log.Errorf("index: %s", err)
		}

		plm := NewPhotoLabel(m.ID, lm.ID, label.Uncertainty, label.Source).FirstOrCreate()

		if plm.Uncertainty > label.Uncertainty && plm.Uncertainty < 100 {
			plm.Uncertainty = label.Uncertainty
			plm.LabelSrc = label.Source
			if err := Db().Save(&plm).Error; err != nil {
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

	if m.DescriptionSrc != SrcAuto && m.DescriptionSrc != source && source != SrcManual && m.Description.PhotoDescription != "" {
		return
	}

	m.Description.PhotoDescription = newDesc
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

	m.UpdateYearMonth()
}

// UpdateYearMonth updates internal date fields.
func (m *Photo) UpdateYearMonth() {
	if m.TakenAt.IsZero() || m.TakenAt.Year() < 1000 {
		return
	}

	if m.TakenAtLocal.IsZero() || m.TakenAtLocal.Year() < 1000 {
		m.TakenAtLocal = m.TakenAt
	}

	if m.TakenSrc == SrcAuto {
		m.PhotoYear = YearUnknown
		m.PhotoMonth = MonthUnknown
	} else {
		m.PhotoYear = m.TakenAtLocal.Year()
		m.PhotoMonth = int(m.TakenAtLocal.Month())
	}
}

// SetCoordinates changes the photo lat, lng and altitude if not empty and from the same source.
func (m *Photo) SetCoordinates(lat, lng float32, altitude int, source string) {
	if lat == 0 && lng == 0 {
		return
	}

	if m.LocationSrc != SrcAuto && m.LocationSrc != source && source != SrcManual {
		return
	}

	m.PhotoLat = lat
	m.PhotoLng = lng
	m.PhotoAltitude = altitude
	m.LocationSrc = source
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
	Db().Unscoped().Delete(PhotoKeyword{}, "photo_id = ?", m.ID)
	Db().Unscoped().Delete(PhotoLabel{}, "photo_id = ?", m.ID)
	Db().Unscoped().Delete(PhotoAlbum{}, "photo_uuid = ?", m.PhotoUUID)

	return Db().Unscoped().Delete(m).Error
}
