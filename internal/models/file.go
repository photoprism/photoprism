package models

import (
	"fmt"
	"strings"

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

func (f *File) DownloadFileName() string {
	if f.Photo == nil {
		return fmt.Sprintf("%s.%s", f.FileHash, f.FileType)
	}

	var name string

	if f.Photo.PhotoTitle != "" {
		name = strings.Title(slug.Make(f.Photo.PhotoTitle))
	} else {
		name = string(f.PhotoID)
	}

	taken := f.Photo.TakenAt.Format("20060102-150405")

	result := fmt.Sprintf("%s-%s.%s", taken, name, f.FileType)

	return result
}
