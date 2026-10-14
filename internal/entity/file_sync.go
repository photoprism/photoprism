package entity

import (
	"fmt"
	"hash/crc32"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Status values a file sync moves through.
const (
	FileSyncNew        = "new"
	FileSyncIgnore     = "ignore"
	FileSyncExists     = "exists"
	FileSyncFailed     = "failed"
	FileSyncDownloaded = "downloaded"
	FileSyncUploaded   = "uploaded"
)

// RemoteNameMaxLen is the length in bytes of the longest remote name a file sync or share record holds.
const RemoteNameMaxLen = 1024

// FileSyncStalledPrefix starts the placeholder names that record local files whose upload stalled, so that the
// next runs upload other files first; refresh removes them, so they are uploaded again in the next sync cycle.
const FileSyncStalledPrefix = fs.PPHiddenPathname + "/sync/stalled/"

// fileSyncLongNames is the folder of the placeholder names that record local files whose remote name is too long.
const fileSyncLongNames = fs.PPHiddenPathname + "/sync/long/"

// FileSync tracks the synchronization status for a file on an external service.
type FileSync struct {
	RemoteName string    `gorm:"primaryKey;autoIncrement:false;type:bytes;size:1024" json:"RemoteName" yaml:"RemoteName,omitempty"`
	ServiceID  uint      `gorm:"primaryKey;autoIncrement:false;size:32" json:"ServiceID" yaml:"ServiceID,omitempty"`
	FileID     uint      `gorm:"index;size:32;" json:"FileID" yaml:"FileID,omitempty"`
	RemoteDate time.Time `json:"RemoteDate" yaml:"RemoteDate,omitempty"`
	RemoteSize int64     `json:"RemoteSize,omitempty" yaml:"RemoteSize,omitempty"`
	Status     string    `gorm:"type:bytes;size:16;" json:"Status" yaml:"Status,omitempty"`
	Error      string    `gorm:"type:bytes;size:512;" json:"Error,omitempty" yaml:"Error,omitempty"`
	Errors     int       `gorm:"size:32" json:"Errors,omitempty" yaml:"Errors,omitempty"`
	File       *File     `gorm:"-:migration" json:"File,omitempty" yaml:"-"`
	Account    *Service  `gorm:"foreignKey:ServiceID;-:migration" json:"Account,omitempty" yaml:"-"`
	CreatedAt  time.Time `json:"CreatedAt" yaml:"CreatedAt"`
	UpdatedAt  time.Time `json:"UpdatedAt" yaml:"UpdatedAt"`
}

// TableName returns the entity table name.
func (FileSync) TableName() string {
	return "files_sync"
}

// NewFileSync creates a new sync record with status preset to "new".
func NewFileSync(accountID uint, remoteName string) *FileSync {
	result := &FileSync{
		ServiceID:  accountID,
		RemoteName: remoteName,
		Status:     FileSyncNew,
	}

	return result
}

// Updates mutates multiple columns on the existing row.
func (m *FileSync) Updates(values any) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Update mutates a single column on the existing row.
func (m *FileSync) Update(attr string, value any) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *FileSync) Save() error {
	return Db().Save(m).Error
}

// Create inserts a new row to the database.
func (m *FileSync) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreateFileSync returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateFileSync(m *FileSync) *FileSync {
	result := FileSync{}

	if err := Db().Where("service_id = ? AND remote_name = ?", m.ServiceID, m.RemoteName).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("file-sync: %s", clean.Error(err))
		return nil
	}

	return m
}

// RemoteNameTooLong reports whether a remote name is longer than a file sync or share record holds.
func RemoteNameTooLong(name string) bool {
	return len(name) > RemoteNameMaxLen
}

// FileSyncLongNamePrefix returns the prefix of the placeholder names that record local files whose remote name
// below syncPath is too long. It includes a checksum of syncPath, so the records of other paths can be told apart.
func FileSyncLongNamePrefix(syncPath string) string {
	return fmt.Sprintf("%s%08x/", fileSyncLongNames, crc32.ChecksumIEEE([]byte(syncPath)))
}

// ResetFileSyncLongNames removes the placeholders an account recorded for a sync path other than syncPath,
// so that the files are checked again.
func ResetFileSyncLongNames(serviceID uint, syncPath string) error {
	return Db().Where("service_id = ? AND remote_name LIKE ? AND remote_name NOT LIKE ?", serviceID,
		fileSyncLongNames+"%", FileSyncLongNamePrefix(syncPath)+"%").Delete(&FileSync{}).Error
}

// ResetFileSyncStalled removes the placeholders of uploads that stalled for an account, so they are tried again.
func ResetFileSyncStalled(serviceID uint) error {
	return Db().Where("service_id = ? AND remote_name LIKE ?", serviceID, FileSyncStalledPrefix+"%").Delete(&FileSync{}).Error
}
