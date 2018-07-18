package photoprism

import (
	"github.com/jinzhu/gorm"
)

type Album struct {
	gorm.Model
	Name   string
	Photos []Photo `gorm:"many2many:album_photos;"`
}
