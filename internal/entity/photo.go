package entity

import (
	"sort"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/util"
)

// A photo can have multiple images and sidecar files
type Photo struct {
	Model
	PhotoUUID         string `gorm:"unique_index;"`
	PhotoToken        string `gorm:"type:varchar(64);"`
	PhotoPath         string `gorm:"type:varchar(128);index;"`
	PhotoName         string
	PhotoTitle        string `json:"PhotoTitle"`
	PhotoTitleChanged bool
	PhotoDescription  string  `gorm:"type:text;"`
	PhotoNotes        string  `gorm:"type:text;"`
	PhotoArtist       string  `json:"PhotoArtist"`
	PhotoFavorite     bool    `json:"PhotoFavorite"`
	PhotoPrivate      bool    `json:"PhotoPrivate"`
	PhotoNSFW         bool    `json:"PhotoNSFW"`
	PhotoStory        bool    `json:"PhotoStory"`
	PhotoLat          float64 `gorm:"index;"`
	PhotoLong         float64 `gorm:"index;"`
	PhotoAltitude     int
	PhotoFocalLength  int
	PhotoIso          int
	PhotoFNumber      float64
	PhotoExposure     string
	PhotoViews        uint
	Camera            *Camera
	CameraID          uint `gorm:"index;"`
	Lens              *Lens
	LensID            uint `gorm:"index;"`
	Country           *Country
	CountryID         string `gorm:"index;"`
	CountryChanged    bool
	Location          *Location
	LocationID        uint
	LocationChanged   bool
	LocationEstimated bool
	TakenAt           time.Time `gorm:"type:datetime;index;"`
	TakenAtLocal      time.Time `gorm:"type:datetime;"`
	TakenAtChanged    bool
	TimeZone          string
	Files             []File
	Labels            []Label
	Keywords          []Keyword
	Albums            []Album
}

func (m *Photo) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("PhotoUUID", util.UUID()); err != nil {
		return err
	}

	if err := scope.SetColumn("PhotoToken", util.RandomToken(4)); err != nil {
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
