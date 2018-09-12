package photoprism

import (
	"github.com/jinzhu/gorm"
)

type Album struct {
	gorm.Model
	AlbumName   string
	Photos []Photo `gorm:"many2many:album_photos;"`
}
