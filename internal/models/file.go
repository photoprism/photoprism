package models

import (
	"github.com/jinzhu/gorm"
)

type File struct {
	gorm.Model
	Photo           *Photo
	PhotoID         uint
	FilePrimary     bool
	FileName        string
	FileType        string `gorm:"type:varchar(30)"`
	FileMime        string `gorm:"type:varchar(50)"`
	FileWidth       int
	FileHeight      int
	FileOrientation int
	FileAspectRatio float64
	FileHash        string `gorm:"type:varchar(100);unique_index"`
	FileNotes       string `gorm:"type:text;"`
}
