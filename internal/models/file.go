package models

import (
	"github.com/jinzhu/gorm"
)

// An image or sidecar file that belongs to a photo
type File struct {
	gorm.Model
	Photo              *Photo
	PhotoID            uint
	FilePrimary        bool
	FileMissing        bool
	FileName           string `gorm:"type:varchar(2048);index"`
	FileType           string `gorm:"type:varchar(32)"`
	FileMime           string `gorm:"type:varchar(64)"`
	FileWidth          int
	FileHeight         int
	FileOrientation    int
	FileAspectRatio    float64
	FileHash           string `gorm:"type:varchar(128);unique_index"`
	FilePerceptualHash string

	FileNotes string `gorm:"type:text;"`
}
