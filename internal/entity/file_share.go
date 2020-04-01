package entity

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
)

// FileShare represents a one-to-many relation between File and Account for pushing files to remote services.
type FileShare struct {
	FileID     uint   `gorm:"primary_key;auto_increment:false"`
	AccountID  uint   `gorm:"primary_key;auto_increment:false"`
	RemoteName string `gorm:"primary_key;auto_increment:false;type:varbinary(256)"`
	Status     string `gorm:"type:varbinary(16);"`
	Error      string `gorm:"type:varbinary(512);"`
	File       *File
	Account    *Account
	SharedAt   sql.NullTime
	ExpiresAt  sql.NullTime
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TableName returns the entity database table name.
func (FileShare) TableName() string {
	return "files_share"
}

// NewFileShare creates a new entity.
func NewFileShare(fileID, accountID uint, remoteName string) *FileShare {
	result := &FileShare{
		FileID:     fileID,
		AccountID:  accountID,
		RemoteName: remoteName,
		Status:     "new",
		Error:      "",
	}

	return result
}

// FirstOrCreate returns the matching entity or creates a new one.
func (m *FileShare) FirstOrCreate(db *gorm.DB) *FileShare {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "file_id = ? AND account_id = ? AND remote_name = ?", m.FileID, m.AccountID, m.RemoteName).Error; err != nil {
		log.Errorf("file push: %s", err)
	}

	return m
}
