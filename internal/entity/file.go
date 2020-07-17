package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/ulule/deepcopier"
)

type Files []File

// File represents an image or sidecar file that belongs to a photo.
type File struct {
	ID              uint          `gorm:"primary_key" json:"-" yaml:"-"`
	UUID            string        `gorm:"type:varbinary(42);index;" json:"InstanceID,omitempty" yaml:"InstanceID,omitempty"`
	Photo           *Photo        `json:"-" yaml:"-"`
	PhotoID         uint          `gorm:"index;" json:"-" yaml:"-"`
	PhotoUID        string        `gorm:"type:varbinary(42);index;" json:"PhotoUID" yaml:"PhotoUID"`
	FileUID         string        `gorm:"type:varbinary(42);unique_index;" json:"UID" yaml:"UID"`
	FileName        string        `gorm:"type:varbinary(768);unique_index:idx_files_name_root;" json:"Name" yaml:"Name"`
	FileRoot        string        `gorm:"type:varbinary(16);default:'';unique_index:idx_files_name_root;" json:"Root" yaml:"Root,omitempty"`
	OriginalName    string        `gorm:"type:varbinary(768);" json:"OriginalName" yaml:"OriginalName,omitempty"`
	FileHash        string        `gorm:"type:varbinary(128);index" json:"Hash" yaml:"Hash,omitempty"`
	FileModified    time.Time     `json:"Modified" yaml:"Modified,omitempty"`
	FileSize        int64         `json:"Size" yaml:"Size,omitempty"`
	FileCodec       string        `gorm:"type:varbinary(32)" json:"Codec" yaml:"Codec,omitempty"`
	FileType        string        `gorm:"type:varbinary(32)" json:"Type" yaml:"Type,omitempty"`
	FileMime        string        `gorm:"type:varbinary(64)" json:"Mime" yaml:"Mime,omitempty"`
	FilePrimary     bool          `json:"Primary" yaml:"Primary,omitempty"`
	FileSidecar     bool          `json:"Sidecar" yaml:"Sidecar,omitempty"`
	FileMissing     bool          `json:"Missing" yaml:"Missing,omitempty"`
	FileDuplicate   bool          `json:"Duplicate" yaml:"Duplicate,omitempty"`
	FilePortrait    bool          `json:"Portrait" yaml:"Portrait,omitempty"`
	FileVideo       bool          `json:"Video" yaml:"Video,omitempty"`
	FileDuration    time.Duration `json:"Duration" yaml:"Duration,omitempty"`
	FileWidth       int           `json:"Width" yaml:"Width,omitempty"`
	FileHeight      int           `json:"Height" yaml:"Height,omitempty"`
	FileOrientation int           `json:"Orientation" yaml:"Orientation,omitempty"`
	FileProjection  string        `gorm:"type:varbinary(16);" json:"Projection,omitempty" yaml:"Projection,omitempty"`
	FileAspectRatio float32       `gorm:"type:FLOAT;" json:"AspectRatio" yaml:"AspectRatio,omitempty"`
	FileMainColor   string        `gorm:"type:varbinary(16);index;" json:"MainColor" yaml:"MainColor,omitempty"`
	FileColors      string        `gorm:"type:varbinary(9);" json:"Colors" yaml:"Colors,omitempty"`
	FileLuminance   string        `gorm:"type:varbinary(9);" json:"Luminance" yaml:"Luminance,omitempty"`
	FileDiff        uint32        `json:"Diff" yaml:"Diff,omitempty"`
	FileChroma      uint8         `json:"Chroma" yaml:"Chroma,omitempty"`
	FileError       string        `gorm:"type:varbinary(512)" json:"Error" yaml:"Error,omitempty"`
	Share           []FileShare   `json:"-" yaml:"-"`
	Sync            []FileSync    `json:"-" yaml:"-"`
	CreatedAt       time.Time     `json:"CreatedAt" yaml:"-"`
	CreatedIn       int64         `json:"CreatedIn" yaml:"-"`
	UpdatedAt       time.Time     `json:"UpdatedAt" yaml:"-"`
	UpdatedIn       int64         `json:"UpdatedIn" yaml:"-"`
	DeletedAt       *time.Time    `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
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

// PrimaryFile returns the primary file for a photo uid.
func PrimaryFile(photoUID string) (File, error) {
	var file File

	q := Db().Unscoped().First(&file, "file_primary = 1 AND photo_uid = ?", photoUID)

	return file, q.Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *File) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.FileUID, 'f') {
		return nil
	}

	return scope.SetColumn("FileUID", rnd.PPID('f'))
}

// ShareFileName returns a meaningful file name useful for sharing.
func (m *File) ShareFileName() string {
	photo := m.RelatedPhoto()

	if photo == nil {
		return fmt.Sprintf("%s.%s", m.FileHash, m.FileType)
	} else if len(m.FileHash) < 8 {
		return fmt.Sprintf("%s.%s", rnd.UUID(), m.FileType)
	} else if photo.TakenAtLocal.IsZero() || photo.PhotoTitle == "" {
		return fmt.Sprintf("%s.%s", m.FileHash, m.FileType)
	}

	name := strings.Title(slug.MakeLang(photo.PhotoTitle, "en"))
	taken := photo.TakenAtLocal.Format("20060102-150405")
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
		Where("photo_id = ? AND file_missing = 0", m.PhotoID).
		Count(&count).Error; err != nil {
		log.Errorf("file: %s", err.Error())
	}

	return count == 0
}

// Create inserts a new row to the database.
func (m *File) Create() error {
	if m.PhotoID == 0 {
		return fmt.Errorf("file: photo id must not be empty (create)")
	}

	if err := UnscopedDb().Create(m).Error; err != nil {
		log.Errorf("file: %s (create)", err)
		return err
	}

	return nil
}

// Saves the file in the database.
func (m *File) Save() error {
	if m.PhotoID == 0 {
		return fmt.Errorf("file: photo id must not be empty (save %s)", m.FileUID)
	}

	if err := UnscopedDb().Save(m).Error; err != nil {
		log.Errorf("file: %s (save %s)", err, m.FileUID)
		return err
	}

	return nil
}

// UpdateVideoInfos updates related video infos based on this file.
func (m *File) UpdateVideoInfos() error {
	values := FileInfos{}

	if err := deepcopier.Copy(&values).From(m); err != nil {
		return err
	}

	return Db().Model(File{}).Where("photo_id = ? AND file_video = 1", m.PhotoID).Updates(values).Error
}

// Updates a column in the database.
func (m *File) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// RelatedPhoto returns the related photo entity.
func (m *File) RelatedPhoto() *Photo {
	if m.Photo != nil {
		return m.Photo
	}

	photo := Photo{}

	UnscopedDb().Model(m).Related(&photo)

	return &photo
}

// NoJPEG returns true if the file is not a JPEG image file.
func (m *File) NoJPEG() bool {
	return m.FileType != string(fs.TypeJpeg)
}

// Links returns all share links for this entity.
func (m *File) Links() Links {
	return FindLinks("", m.FileUID)
}

// Panorama tests if the file seems to be a panorama image.
func (m *File) Panorama() bool {
	if m.FileSidecar || m.FileWidth <= 1000 || m.FileHeight <= 500 {
		return false
	}

	return m.FileProjection != ProjectionDefault || (m.FileWidth/m.FileHeight) >= 2
}
