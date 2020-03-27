package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// File represents an image or sidecar file that belongs to a photo
type File struct {
	ID              uint `gorm:"primary_key"`
	Photo           *Photo
	PhotoID         uint   `gorm:"index;"`
	PhotoUUID       string `gorm:"type:varbinary(36);index;"`
	FileUUID        string `gorm:"type:varbinary(36);unique_index;"`
	FileName        string `gorm:"type:varbinary(600);unique_index"`
	OriginalName    string `gorm:"type:varbinary(600);"`
	FileHash        string `gorm:"type:varbinary(128);index"`
	FileModified    time.Time
	FileSize        int64
	FileType        string `gorm:"type:varbinary(32)"`
	FileMime        string `gorm:"type:varbinary(64)"`
	FilePrimary     bool
	FileSidecar     bool
	FileVideo       bool
	FileMissing     bool
	FileDuplicate   bool
	FilePortrait    bool
	FileWidth       int
	FileHeight      int
	FileOrientation int
	FileAspectRatio float64
	FileMainColor   string `gorm:"type:varbinary(16);index;"`
	FileColors      string `gorm:"type:binary(9);"`
	FileLuminance   string `gorm:"type:binary(9);"`
	FileChroma      uint
	FileNotes       string `gorm:"type:text"`
	FileError       string `gorm:"type:varbinary(512)"`
	FilePush        []FilePush
	FileSync        []FileSync
	CreatedAt       time.Time
	CreatedIn       int64
	UpdatedAt       time.Time
	UpdatedIn       int64
	DeletedAt       *time.Time `sql:"index"`
}

// FirstFileByHash gets a file in db from its hash
func FirstFileByHash(db *gorm.DB, fileHash string) (File, error) {
	var file File

	q := db.Unscoped().First(&file, "file_hash = ?", fileHash)

	return file, q.Error
}

// BeforeCreate computes a random UUID when a new file is created in database
func (m *File) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("FileUUID", rnd.PPID('f'))
}

// DownloadFileName returns a name useful for download links
func (m *File) DownloadFileName() string {
	if m.Photo == nil {
		return fmt.Sprintf("%s.%s", m.FileHash, m.FileType)
	}

	var name string

	if m.Photo.PhotoTitle != "" {
		name = strings.Title(slug.MakeLang(m.Photo.PhotoTitle, "en"))
	} else {
		name = m.PhotoUUID
	}

	taken := m.Photo.TakenAt.Format("20060102-150405")

	result := fmt.Sprintf("%s-%s.%s", taken, name, m.FileType)

	return result
}

// Changed returns true if new and old file size or modified time are different.
func (m File) Changed(fileSize int64, fileModified time.Time) bool {
	if m.FileSize != fileSize {
		return true
	}

	if m.FileModified.Format("2006-01-02 15:04:05") != fileModified.Format("2006-01-02 15:04:05") {
		return true
	}

	return false
}
