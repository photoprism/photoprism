package entity

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
)

// FileSync represents a one-to-many relation between File and Account for syncing with remote services.
type FileSync struct {
	FileID     uint   `gorm:"primary_key;auto_increment:false"`
	AccountID  uint   `gorm:"primary_key;auto_increment:false"`
	RemoteName string `gorm:"type:varbinary(256);"`
	SyncStatus string `gorm:"type:varbinary(16);"`
	SyncError  string `gorm:"type:varbinary(512);"`
	RetryCount uint
	File       *File
	Account    *Account
	SyncedAt   sql.NullTime
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TableName returns the entity database table name.
func (FileSync) TableName() string {
	return "files_sync"
}

// NewFileSync creates a new entity.
func NewFileSync(fileID, accountID uint) *FileSync {
	result := &FileSync{
		FileID:    fileID,
		AccountID: accountID,
	}

	return result
}

// FirstOrCreate returns the matching entity or creates a new one.
func (m *FileSync) FirstOrCreate(db *gorm.DB) *FileSync {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "file_id = ? AND account_id = ?", m.FileID, m.AccountID).Error; err != nil {
		log.Errorf("file sync: %s", err)
	}

	return m
}
