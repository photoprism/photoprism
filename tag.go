package photoprism

import (
	"github.com/jinzhu/gorm"
)

type Tag struct {
	gorm.Model
	Label string
}
