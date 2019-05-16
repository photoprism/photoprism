package models

import (
	"fmt"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
)

// An image or sidecar file that belongs to a photo
type File struct {
	gorm.Model
	Photo            *Photo
	PhotoID          uint
	FilePrimary      bool
	FileMissing      bool
	FileDuplicate    bool
	FileName         string `gorm:"type:varchar(512);index"` // max 3072 bytes / 4 bytes for utf8mb4 = 768 chars
	FileOriginalName string
	FileType         string `gorm:"type:varchar(32)"`
	FileMime         string `gorm:"type:varchar(64)"`
	FileWidth        int
	FileHeight       int
	FileOrientation  int
	FileAspectRatio  float64
	FilePortrait     bool
	FileMainColor    string
	FileColors       string
	FileLuminance    string
	FileChroma       uint
	FileHash         string `gorm:"type:varchar(128);unique_index"`
	FileNotes        string `gorm:"type:text"`
}

func (f *File) DownloadFileName(db *gorm.DB) string {
	var photo Photo

	db.Model(f).Related(&photo)

	name := slug.MakeLang(photo.PhotoTitle, "en")

	result := fmt.Sprintf("%s.%s", name, f.FileType)

	return result
}
