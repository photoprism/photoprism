package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Photo struct {
	gorm.Model
	TakenAt             time.Time
	PhotoTitle          string
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
	Country             *Country
	CountryID           uint
	Location            *Location
	LocationID          uint
	Tags                []*Tag `gorm:"many2many:photo_tags;"`
	Files               []*File
	Albums              []*Album `gorm:"many2many:album_photos;"`
	Camera              *Camera
	CameraID            uint
}
