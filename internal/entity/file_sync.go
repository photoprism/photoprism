package entity

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
)

// FileSync represents a one-to-many relation between File and Account for syncing with remote services.
type FileSync struct {
	FileID     uint   `gorm:"index;"`
	AccountID  uint   `gorm:"primary_key;auto_increment:false"`
	RemoteName string `gorm:"type:varbinary(256);primary_key;auto_increment:false"`
	RemoteDate time.Time
	RemoteSize int64
	Status     string `gorm:"type:varbinary(16);"`
	Error      string `gorm:"type:varbinary(512);"`
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
func NewFileSync(accountID uint, remoteName string) *FileSync {
	result := &FileSync{
		AccountID:  accountID,
		RemoteName: remoteName,
		Status:     "new",
	}

	return result
}

// FirstOrCreate returns the matching entity or creates a new one.
func (m *FileSync) FirstOrCreate(db *gorm.DB) *FileSync {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "account_id = ? AND remote_name = ?", m.AccountID, m.RemoteName).Error; err != nil {
		log.Errorf("file sync: %s", err)
	}

	return m
}
