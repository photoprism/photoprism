package entity

import (
	"fmt"
	"hash/crc32"
	"strconv"
	"time"

	"gorm.io/gorm"

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
	FileSyncRetry      = "retry"
)

// RemoteNameMaxLen is the length in bytes of the longest remote name a file sync or share record holds.
const RemoteNameMaxLen = 1024

// FileSyncUploadPrefix starts the placeholder names that record local files whose upload failed or stalled, with
// their error count, so a sync cycle tries each file once; refresh marks them for the next cycle.
const FileSyncUploadPrefix = fs.PPHiddenPathname + "/sync/upload/"

// fileSyncStalled is the folder of placeholders of stalled uploads without an error count, which refresh removes.
const fileSyncStalled = fs.PPHiddenPathname + "/sync/stalled/"

// FileSyncUnportablePrefix starts the placeholder names that record local files whose names are not portable.
const FileSyncUnportablePrefix = fs.PPHiddenPathname + "/sync/names/"

// fileSyncLongNames is the folder of the placeholder names that record local files whose remote name is too long.
const fileSyncLongNames = fs.PPHiddenPathname + "/sync/long/"

// FileSync tracks the synchronization status for a file on an external service.
type FileSync struct {
	RemoteName string    `gorm:"primaryKey;autoIncrement:false;type:bytes;size:1024" json:"RemoteName" yaml:"RemoteName,omitempty"`
	ServiceID  uint      `gorm:"primaryKey;autoIncrement:false;size:32;index:idx_files_sync_service_status" json:"ServiceID" yaml:"ServiceID,omitempty"`
	FileID     uint      `gorm:"index;size:32;" json:"FileID" yaml:"FileID,omitempty"`
	RemoteDate time.Time `json:"RemoteDate" yaml:"RemoteDate,omitempty"`
	RemoteSize int64     `json:"RemoteSize,omitempty" yaml:"RemoteSize,omitempty"`
	Status     string    `gorm:"type:bytes;size:16;index:idx_files_sync_service_status" json:"Status" yaml:"Status,omitempty"`
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

// ResetFileSyncUnportable removes the placeholders an account recorded for local files whose names are not
// portable once their names are, e.g. after a rename, so that the files are uploaded.
func ResetFileSyncUnportable(serviceID uint) error {
	var rows []struct {
		RemoteName string
		FileName   string
	}

	if err := Db().Table("files_sync").Select("files_sync.remote_name, files.file_name").
		Joins("JOIN files ON files.id = files_sync.file_id").
		Where("files_sync.service_id = ? AND files_sync.remote_name LIKE ?", serviceID, FileSyncUnportablePrefix+"%").
		Scan(&rows).Error; err != nil {
		return err
	}

	names := make([]string, 0, len(rows))

	for _, row := range rows {
		if clean.UnportablePath(row.FileName) == "" {
			names = append(names, row.RemoteName)
		}
	}

	for len(names) > 0 {
		n := min(len(names), 500)

		if err := Db().Where("service_id = ? AND remote_name IN (?)", serviceID, names[:n]).Delete(&FileSync{}).Error; err != nil {
			return err
		}

		names = names[n:]
	}

	return nil
}

// FileSyncUploadName returns the placeholder name that records a failed upload of a local file.
func FileSyncUploadName(fileID uint) string {
	return FileSyncUploadPrefix + strconv.FormatUint(uint64(fileID), 10)
}

// RetryFileSyncUploads marks the failed uploads of an account to be tried again in its next sync cycle, except
// those that failed more often than a positive retry limit allows.
func RetryFileSyncUploads(serviceID uint, retryLimit int) error {
	if err := Db().Where("service_id = ? AND remote_name LIKE ?", serviceID, fileSyncStalled+"%").
		Delete(&FileSync{}).Error; err != nil {
		return err
	}

	// Most refreshes find nothing to retry, so nothing is written then.
	if count, err := CountFailedFileSyncUploads(serviceID); err != nil || count == 0 {
		return err
	}

	stmt := Db().Model(&FileSync{}).Where("service_id = ? AND remote_name LIKE ? AND status = ?",
		serviceID, FileSyncUploadPrefix+"%", FileSyncFailed)

	if retryLimit > 0 {
		stmt = stmt.Where("COALESCE(errors, 0) <= ?", retryLimit)
	}

	return stmt.UpdateColumn("status", FileSyncRetry).Error
}

// FailFileSyncUpload records that uploading a local file failed, so it is not selected again in this sync cycle,
// and counts the error.
func FailFileSyncUpload(serviceID, fileID uint, errMsg string) error {
	name := FileSyncUploadName(fileID)

	if res := Db().Model(&FileSync{}).Where("service_id = ? AND remote_name = ?", serviceID, name).
		UpdateColumns(Values{"file_id": fileID, "status": FileSyncFailed, "error": errMsg,
			"errors": gorm.Expr("COALESCE(errors, 0) + 1")}); res.Error != nil {
		return res.Error
	} else if res.RowsAffected > 0 {
		return nil
	}

	m := NewFileSync(serviceID, name)
	m.FileID, m.Status, m.Error, m.Errors = fileID, FileSyncFailed, errMsg, 1

	return m.Create()
}

// ClearFileSyncUpload removes the placeholder of a failed upload once the file was uploaded.
func ClearFileSyncUpload(serviceID, fileID uint) error {
	return Db().Where("service_id = ? AND remote_name = ?", serviceID, FileSyncUploadName(fileID)).
		Delete(&FileSync{}).Error
}

// CountFailedFileSyncUploads returns the number of local files of an account whose last upload failed, except
// files that are missing or deleted.
func CountFailedFileSyncUploads(serviceID uint) (count int64, err error) {
	err = Db().Model(&FileSync{}).Where("service_id = ? AND remote_name LIKE ? AND status = ?",
		serviceID, FileSyncUploadPrefix+"%", FileSyncFailed).
		Where("NOT EXISTS (SELECT 1 FROM files f WHERE f.id = files_sync.file_id AND (f.file_missing = TRUE OR f.deleted_at IS NOT NULL))").
		Count(&count).Error

	return count, err
}
