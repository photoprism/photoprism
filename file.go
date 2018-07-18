package photoprism

import (
	"github.com/jinzhu/gorm"
)

type File struct {
	gorm.Model
	PhotoID  uint
	Filename string
	Hash     string
	FileType string
	MimeType string
}
