package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Photo struct {
	gorm.Model
	TakenAt             time.Time
	PhotoTitle          string
	PhotoTitleChanged   bool
	PhotoDescription    string `gorm:"type:text;"`
	PhotoNotes          string `gorm:"type:text;"`
	PhotoArtist         string
	PhotoColors         string
	PhotoVibrantColor   string
	PhotoMutedColor     string
	PhotoCanonicalName  string
	PhotoPerceptualHash string
	PhotoFavorite       bool
	PhotoLat            float64
	PhotoLong           float64
	PhotoFocalLength    float64
	PhotoAperture       float64
	Camera              *Camera
	CameraID            uint
	Lens                *Lens
	LensID              uint
	Country             *Country
	CountryID           string
	CountryChanged      bool
	Location            *Location
	LocationID          uint
	LocationChanged     bool
	Tags                []*Tag `gorm:"many2many:photo_tags;"`
	Files               []*File
	Albums              []*Album `gorm:"many2many:album_photos;"`
}
