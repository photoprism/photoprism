package entity

import (
	"sort"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/util"
)

// A photo can have multiple images and sidecar files
type Photo struct {
	ID                uint      `gorm:"primary_key"`
	TakenAt           time.Time `gorm:"type:datetime;index:idx_photos_taken_uuid;"`
	PhotoUUID         string    `gorm:"type:varbinary(36);unique_index;index:idx_photos_taken_uuid;"`
	PhotoPath         string    `gorm:"type:varbinary(400);index;"`
	PhotoName         string    `gorm:"type:varbinary(200);"`
	PhotoTitle        string    `json:"PhotoTitle"`
	PhotoTitleChanged bool
	PhotoDescription  string  `gorm:"type:text;"`
	PhotoNotes        string  `gorm:"type:text;"`
	PhotoArtist       string  `json:"PhotoArtist"`
	PhotoFavorite     bool    `json:"PhotoFavorite"`
	PhotoPrivate      bool    `json:"PhotoPrivate"`
	PhotoNSFW         bool    `json:"PhotoNSFW"`
	PhotoStory        bool    `json:"PhotoStory"`
	PhotoLat          float64 `gorm:"index;"`
	PhotoLng          float64 `gorm:"index;"`
	PhotoAltitude     int
	PhotoFocalLength  int
	PhotoIso          int
	PhotoFNumber      float64
	PhotoExposure     string `gorm:"type:varbinary(16);"`
	PhotoViews        uint
	Camera            *Camera
	CameraID          uint `gorm:"index:idx_photos_camera_lens;"`
	Lens              *Lens
	LensID            uint `gorm:"index:idx_photos_camera_lens;"`
	CountryChanged    bool
	Location          *Location
	LocationID        string `gorm:"type:varbinary(16);index;"`
	Place             *Place
	PlaceID           string `gorm:"type:varbinary(16);index;"`
	LocationChanged   bool
	LocationEstimated bool
	PhotoCountry      string    `gorm:"index:idx_photos_country_year_month;"`
	PhotoYear         int       `gorm:"index:idx_photos_country_year_month;"`
	PhotoMonth        int       `gorm:"index:idx_photos_country_year_month;"`
	TakenAtLocal      time.Time `gorm:"type:datetime;"`
	TakenAtChanged    bool
	TimeZone          string `gorm:"type:varbinary(64);"`
	Files             []File
	Labels            []Label
	Keywords          []Keyword
	Albums            []Album
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time `sql:"index"`
}

func (m *Photo) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("PhotoUUID", ID('p')); err != nil {
		return err
	}

	return nil
}

func (m *Photo) IndexKeywords(keywords []string, db *gorm.DB) {
	var keywordIds []uint

	// Index title and description
	keywords = append(keywords, util.Keywords(m.PhotoTitle)...)
	keywords = append(keywords, util.Keywords(m.PhotoDescription)...)
	last := ""

	sort.Strings(keywords)

	for _, w := range keywords {
		if len(w) < 3 || w == last {
			continue
		}

		last = w
		kw := NewKeyword(w).FirstOrCreate(db)

		if kw.Skip {
			continue
		}

		keywordIds = append(keywordIds, kw.ID)

		NewPhotoKeyword(m.ID, kw.ID).FirstOrCreate(db)
	}

	db.Where("photo_id = ? AND keyword_id NOT IN (?)", m.ID, keywordIds).Delete(&PhotoKeyword{})
}

func (m *Photo) PreloadFiles(db *gorm.DB) {
	q := db.NewScope(nil).DB().
		Table("files").
		Select(`files.*`).
		Where("files.photo_id = ?", m.ID).
		Order("files.file_primary DESC")

	logError(q.Scan(&m.Files))
}

func (m *Photo) PreloadLabels(db *gorm.DB) {
	q := db.NewScope(nil).DB().
		Table("labels").
		Select(`labels.*`).
		Joins("JOIN photos_labels ON photos_labels.label_id = labels.id AND photos_labels.photo_id = ?", m.ID).
		Where("labels.deleted_at IS NULL").
		Order("labels.label_name ASC")

	logError(q.Scan(&m.Labels))
}

func (m *Photo) PreloadKeywords(db *gorm.DB) {
	q := db.NewScope(nil).DB().
		Table("keywords").
		Select(`keywords.*`).
		Joins("JOIN photos_keywords ON photos_keywords.keyword_id = keywords.id AND photos_keywords.photo_id = ?", m.ID).
		Order("keywords.keyword ASC")

	logError(q.Scan(&m.Keywords))
}

func (m *Photo) PreloadAlbums(db *gorm.DB) {
	q := db.NewScope(nil).DB().
		Table("albums").
		Select(`albums.*`).
		Joins("JOIN photos_albums ON photos_albums.album_uuid = albums.album_uuid AND photos_albums.photo_uuid = ?", m.PhotoUUID).
		Where("albums.deleted_at IS NULL").
		Order("albums.album_name ASC")

	logError(q.Scan(&m.Albums))
}

func (m *Photo) PreloadMany(db *gorm.DB) {
	m.PreloadFiles(db)
	m.PreloadLabels(db)
	m.PreloadKeywords(db)
	m.PreloadAlbums(db)
}

func (m *Photo) NoLocation() bool {
	return m.LocationID == ""
}

func (m *Photo) HasLocation() bool {
	return m.LocationID != ""
}

func (m *Photo) NoPlace() bool {
	return len(m.PlaceID) < 2
}

func (m *Photo) HasPlace() bool {
	return len(m.PlaceID) >= 2
}

func (m *Photo) NoTitle() bool {
	return m.PhotoTitle == ""
}

func (m *Photo) HasTitle() bool {
	return m.PhotoTitle != ""
}
