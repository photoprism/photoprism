package entity

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/ulule/deepcopier"
)

type DownloadName string

const (
	DownloadNameFile     DownloadName = "file"
	DownloadNameOriginal DownloadName = "original"
	DownloadNameShare    DownloadName = "share"
	DownloadNameDefault               = DownloadNameFile
)

type Files []File

// File represents an image or sidecar file that belongs to a photo.
type File struct {
	ID              uint          `gorm:"primary_key" json:"-" yaml:"-"`
	Photo           *Photo        `json:"-" yaml:"-"`
	PhotoID         uint          `gorm:"index;" json:"-" yaml:"-"`
	PhotoUID        string        `gorm:"type:VARBINARY(42);index;" json:"PhotoUID" yaml:"PhotoUID"`
	InstanceID      string        `gorm:"type:VARBINARY(42);index;" json:"InstanceID,omitempty" yaml:"InstanceID,omitempty"`
	FileUID         string        `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	FileName        string        `gorm:"type:VARBINARY(755);unique_index:idx_files_name_root;" json:"Name" yaml:"Name"`
	FileRoot        string        `gorm:"type:VARBINARY(16);default:'/';unique_index:idx_files_name_root;" json:"Root" yaml:"Root,omitempty"`
	OriginalName    string        `gorm:"type:VARBINARY(755);" json:"OriginalName" yaml:"OriginalName,omitempty"`
	FileHash        string        `gorm:"type:VARBINARY(128);index" json:"Hash" yaml:"Hash,omitempty"`
	FileSize        int64         `json:"Size" yaml:"Size,omitempty"`
	FileCodec       string        `gorm:"type:VARBINARY(32)" json:"Codec" yaml:"Codec,omitempty"`
	FileType        string        `gorm:"type:VARBINARY(32)" json:"Type" yaml:"Type,omitempty"`
	FileMime        string        `gorm:"type:VARBINARY(64)" json:"Mime" yaml:"Mime,omitempty"`
	FilePrimary     bool          `json:"Primary" yaml:"Primary,omitempty"`
	FileSidecar     bool          `json:"Sidecar" yaml:"Sidecar,omitempty"`
	FileMissing     bool          `json:"Missing" yaml:"Missing,omitempty"`
	FilePortrait    bool          `json:"Portrait" yaml:"Portrait,omitempty"`
	FileVideo       bool          `json:"Video" yaml:"Video,omitempty"`
	FileDuration    time.Duration `json:"Duration" yaml:"Duration,omitempty"`
	FileWidth       int           `json:"Width" yaml:"Width,omitempty"`
	FileHeight      int           `json:"Height" yaml:"Height,omitempty"`
	FileOrientation int           `json:"Orientation" yaml:"Orientation,omitempty"`
	FileProjection  string        `gorm:"type:VARBINARY(16);" json:"Projection,omitempty" yaml:"Projection,omitempty"`
	FileAspectRatio float32       `gorm:"type:FLOAT;" json:"AspectRatio" yaml:"AspectRatio,omitempty"`
	FileMainColor   string        `gorm:"type:VARBINARY(16);index;" json:"MainColor" yaml:"MainColor,omitempty"`
	FileColors      string        `gorm:"type:VARBINARY(9);" json:"Colors" yaml:"Colors,omitempty"`
	FileLuminance   string        `gorm:"type:VARBINARY(9);" json:"Luminance" yaml:"Luminance,omitempty"`
	FileDiff        uint32        `json:"Diff" yaml:"Diff,omitempty"`
	FileChroma      uint8         `json:"Chroma" yaml:"Chroma,omitempty"`
	FileError       string        `gorm:"type:VARBINARY(512)" json:"Error" yaml:"Error,omitempty"`
	ModTime         int64         `json:"ModTime" yaml:"-"`
	CreatedAt       time.Time     `json:"CreatedAt" yaml:"-"`
	CreatedIn       int64         `json:"CreatedIn" yaml:"-"`
	UpdatedAt       time.Time     `json:"UpdatedAt" yaml:"-"`
	UpdatedIn       int64         `json:"UpdatedIn" yaml:"-"`
	DeletedAt       *time.Time    `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
	Share           []FileShare   `json:"-" yaml:"-"`
	Sync            []FileSync    `json:"-" yaml:"-"`
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

	res := Db().Unscoped().First(&file, "file_hash = ?", fileHash)

	return file, res.Error
}

// PrimaryFile returns the primary file for a photo uid.
func PrimaryFile(photoUID string) (File, error) {
	var file File

	res := Db().Unscoped().First(&file, "file_primary = 1 AND photo_uid = ?", photoUID)

	return file, res.Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *File) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.FileUID, 'f') {
		return nil
	}

	return scope.SetColumn("FileUID", rnd.PPID('f'))
}

// DownloadName returns the download file name.
func (m *File) DownloadName(n DownloadName, seq int) string {
	switch n {
	case DownloadNameFile:
		return m.Base(seq)
	case DownloadNameOriginal:
		return m.OriginalBase(seq)
	default:
		return m.ShareBase(seq)
	}
}

// Base returns the file name without path.
func (m *File) Base(seq int) string {
	if m.FileName == "" {
		return m.ShareBase(seq)
	}

	base := filepath.Base(m.FileName)

	if seq > 0 {
		return fmt.Sprintf("%s (%d)%s", fs.StripExt(base), seq, filepath.Ext(base))
	}

	return base
}

// OriginalBase returns the original file name without path.
func (m *File) OriginalBase(seq int) string {
	if m.OriginalName == "" {
		return m.Base(seq)
	}

	base := filepath.Base(m.OriginalName)

	if seq > 0 {
		return fmt.Sprintf("%s (%d)%s", fs.StripExt(base), seq, filepath.Ext(base))
	}

	return base
}

// ShareBase returns a meaningful file name for sharing.
func (m *File) ShareBase(seq int) string {
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

	if seq > 0 {
		return fmt.Sprintf("%s-%s (%d).%s", taken, name, seq, m.FileType)
	}

	return fmt.Sprintf("%s-%s.%s", taken, name, m.FileType)
}

// Changed returns true if new and old file size or modified time are different.
func (m File) Changed(fileSize int64, modTime time.Time) bool {
	// File size has changed.
	if m.FileSize != fileSize {
		return true
	}

	// Modification time has changed.
	if m.ModTime == modTime.Unix() {
		return false
	}

	return true
}

// Missing returns true if this file is current missing or marked as deleted.
func (m File) Missing() bool {
	return m.FileMissing || m.DeletedAt != nil
}

// Delete permanently deletes the entity from the database.
func (m *File) DeletePermanently() error {
	Db().Unscoped().Delete(FileShare{}, "file_id = ?", m.ID)
	Db().Unscoped().Delete(FileSync{}, "file_id = ?", m.ID)

	return Db().Unscoped().Delete(m).Error
}

// Delete deletes the entity from the database.
func (m *File) Delete(permanently bool) error {
	if permanently {
		return m.DeletePermanently()
	}

	Db().Delete(File{}, "id = ?", m.ID)

	return Db().Delete(m).Error
}

// Purge removes a file from the index by marking it as missing.
func (m *File) Purge() error {
	deletedAt := Timestamp()
	m.FileMissing = true
	m.FilePrimary = false
	m.DeletedAt = &deletedAt
	return UnscopedDb().Exec("UPDATE files SET file_missing = 1, file_primary = 0, deleted_at = ? WHERE id = ?", &deletedAt, m.ID).Error
}

// Found restores a previously purged file.
func (m *File) Found() error {
	m.FileMissing = false
	m.DeletedAt = nil
	return UnscopedDb().Exec("UPDATE files SET file_missing = 0, deleted_at = NULL WHERE id = ?", m.ID).Error
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

// ResolvePrimary ensures there is only one primary file for a photo..
func (m *File) ResolvePrimary() error {
	if m.FilePrimary {
		return UnscopedDb().Exec("UPDATE `files` SET file_primary = (id = ?) WHERE photo_id = ?", m.ID, m.PhotoID).Error
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

	return m.ResolvePrimary()
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

// Updates multiple columns in the database.
func (m *File) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Rename updates the name and path of this file.
func (m *File) Rename(fileName, rootName, filePath, fileBase string) error {
	log.Debugf("file: renaming %s to %s", txt.Quote(m.FileName), txt.Quote(fileName))

	// Update database row.
	if err := m.Updates(map[string]interface{}{
		"FileName":    fileName,
		"FileRoot":    rootName,
		"FileMissing": false,
		"DeletedAt":   nil,
	}); err != nil {
		return err
	}

	m.FileName = fileName
	m.FileRoot = rootName
	m.FileMissing = false
	m.DeletedAt = nil

	// Update photo path and name if possible.
	if p := m.RelatedPhoto(); p != nil {
		return p.Updates(map[string]interface{}{
			"PhotoPath": filePath,
			"PhotoName": fileBase,
		})
	}

	return nil
}

// Undelete removes the missing flag from this file.
func (m *File) Undelete() error {
	if !m.Missing() {
		return nil
	}

	// Update database row.
	err := m.Updates(map[string]interface{}{
		"FileMissing": false,
		"DeletedAt":   nil,
	})

	if err != nil {
		return err
	}

	log.Debugf("file: removed missing flag from %s", txt.Quote(m.FileName))

	m.FileMissing = false
	m.DeletedAt = nil

	return nil
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
	return m.FileType != string(fs.FormatJpeg)
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
