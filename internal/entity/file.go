package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/ulule/deepcopier"
)

// File represents an image or sidecar file that belongs to a photo
type File struct {
	ID              uint `gorm:"primary_key"`
	Photo           *Photo
	PhotoID         uint   `gorm:"index;"`
	PhotoUUID       string `gorm:"type:varbinary(36);index;"`
	FileUUID        string `gorm:"type:varbinary(36);unique_index;"`
	FileName        string `gorm:"type:varbinary(768);unique_index"`
	OriginalName    string `gorm:"type:varbinary(768);"`
	FileHash        string `gorm:"type:varbinary(128);index"`
	FileModified    time.Time
	FileSize        int64
	FileCodec       string `gorm:"type:varbinary(32)"`
	FileType        string `gorm:"type:varbinary(32)"`
	FileMime        string `gorm:"type:varbinary(64)"`
	FilePrimary     bool
	FileSidecar     bool
	FileMissing     bool
	FileDuplicate   bool
	FilePortrait    bool
	FileVideo       bool
	FileDuration    time.Duration
	FileWidth       int
	FileHeight      int
	FileOrientation int
	FileAspectRatio float32 `gorm:"type:FLOAT;"`
	FileMainColor   string  `gorm:"type:varbinary(16);index;"`
	FileColors      string  `gorm:"type:varbinary(9);"`
	FileLuminance   string  `gorm:"type:varbinary(9);"`
	FileDiff        uint32
	FileChroma      uint8
	FileNotes       string `gorm:"type:text"`
	FileError       string `gorm:"type:varbinary(512)"`
	Share           []FileShare
	Sync            []FileSync
	Links           []Link `gorm:"foreignkey:ShareUUID;association_foreignkey:FileUUID"`
	CreatedAt       time.Time
	CreatedIn       int64
	UpdatedAt       time.Time
	UpdatedIn       int64
	DeletedAt       *time.Time `sql:"index"`
}

type FileInfos struct {
	FileWidth       int
	FileHeight      int
	FileOrientation int
	FileAspectRatio float32
	FileMainColor   string
	FileColors      string
	FileLuminance   string
	FileDiff        uint32
	FileChroma      uint8
}

// FirstFileByHash gets a file in db from its hash
func FirstFileByHash(fileHash string) (File, error) {
	var file File

	q := Db().Unscoped().First(&file, "file_hash = ?", fileHash)

	return file, q.Error
}

// BeforeCreate computes a random UUID when a new file is created in database
func (m *File) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsPPID(m.FileUUID, 'f') {
		return nil
	}

	return scope.SetColumn("FileUUID", rnd.PPID('f'))
}

// ShareFileName returns a meaningful file name useful for sharing.
func (m *File) ShareFileName() string {
	if m.Photo == nil {
		return fmt.Sprintf("%s.%s", m.FileHash, m.FileType)
	}

	var name string

	if m.Photo.PhotoTitle != "" {
		name = strings.Title(slug.MakeLang(m.Photo.PhotoTitle, "en"))
	} else {
		name = m.PhotoUUID
	}

	taken := m.Photo.TakenAtLocal.Format("20060102-150405")
	token := rnd.Token(3)

	result := fmt.Sprintf("%s-%s-%s.%s", taken, name, token, m.FileType)

	return result
}

// Changed returns true if new and old file size or modified time are different.
func (m File) Changed(fileSize int64, fileModified time.Time) bool {
	if m.DeletedAt != nil {
		return true
	}

	if m.FileSize != fileSize {
		return true
	}

	if m.FileModified.Round(time.Second).Equal(fileModified.Round(time.Second)) {
		return false
	}

	return true
}

// Purge removes a file from the index by marking it as missing.
func (m *File) Purge() error {
	return Db().Unscoped().Model(m).Updates(map[string]interface{}{"file_missing": true, "file_primary": false}).Error
}

// AllFilesMissing returns true, if all files for the photo of this file are missing.
func (m *File) AllFilesMissing() bool {
	count := 0

	if err := Db().Model(&File{}).
		Where("photo_id = ? AND b.file_missing = 0", m.PhotoID).
		Count(&count).Error; err != nil {
		log.Error(err)
	}

	return count == 0
}

// Save stored the file in the database using the default connection.
func (m *File) Save() error {
	if err := Db().Save(m).Error; err != nil {
		return err
	}

	return Db().Model(m).Related(Photo{}).Error
}

// UpdateVideoInfos updates related video infos based on this file.
func (m *File) UpdateVideoInfos() error {
	values := FileInfos{}

	if err := deepcopier.Copy(&values).From(m); err != nil {
		return err
	}

	return Db().Model(File{}).Where("photo_id = ? AND file_video = 1", m.PhotoID).Updates(values).Error
}
