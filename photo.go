package photoprism

import (
	"github.com/jinzhu/gorm"
)

type Photo struct {
	gorm.Model
	Title          string
	Description    string  `gorm:"type:text;"`
	CanonicalName  string
	PerceptualHash string
	Tags           []Tag   `gorm:"many2many:photo_tags;"`
	Files          []File
	Albums         []Album `gorm:"many2many:album_photos;"`
	Author         string
	CameraModel    string
	LocationName   string
	Lat            float64
	Long           float64
	Liked          bool
	Private        bool
	Deleted        bool
}
