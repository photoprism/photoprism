package photoprism

import (
	"github.com/jinzhu/gorm"
)

type File struct {
	gorm.Model
	Photo       *Photo
	PhotoID     uint
	Filename    string
	FileType    string `gorm:"type:varchar(30)"`
	MimeType    string `gorm:"type:varchar(50)"`
	Width       int
	Height      int
	Orientation int
	AspectRatio float64
	Hash        string `gorm:"type:varchar(100);unique_index"`
}
