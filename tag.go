package photoprism

import (
	"github.com/jinzhu/gorm"
)

type Tag struct {
	gorm.Model
	TagLabel string `gorm:"type:varchar(100);unique_index"`
}
