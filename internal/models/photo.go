package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/util"
	uuid "github.com/satori/go.uuid"
)

// A photo can have multiple images and sidecar files
type Photo struct {
	Model
	PhotoUUID         string `gorm:"unique_index;"`
	PhotoToken        string `gorm:"type:varchar(64);"`
	PhotoPath         string `gorm:"type:varchar(128);index;"`
	PhotoName         string
	PhotoTitle        string
	PhotoTitleChanged bool
	PhotoDescription  string `gorm:"type:text;"`
	PhotoNotes        string `gorm:"type:text;"`
	PhotoArtist       string
	PhotoFavorite     bool
	PhotoPrivate      bool
	PhotoNSFW         bool
	PhotoStory        bool
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
	TakenAt           time.Time `gorm:"index;"`
	TakenAtLocal      time.Time
	TakenAtChanged    bool
	TimeZone          string
	Labels            []*PhotoLabel
	Albums            []*PhotoAlbum
	Files             []*File
}

func (m *Photo) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("PhotoUUID", uuid.NewV4().String()); err != nil {
		return err
	}

	if err := scope.SetColumn("PhotoToken", util.RandomToken(4)); err != nil {
		return err
	}

	return nil
}
