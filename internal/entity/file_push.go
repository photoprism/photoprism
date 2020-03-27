package entity

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/mutex"
)

// FilePush represents a one-to-many relation between File and Account for pushing files to remote services.
type FilePush struct {
	FileID     uint   `gorm:"primary_key;auto_increment:false"`
	AccountID  uint   `gorm:"primary_key;auto_increment:false"`
	RemoteName string `gorm:"primary_key;auto_increment:false;type:varbinary(256)"`
	PushStatus string `gorm:"type:varbinary(16);"`
	PushError  string `gorm:"type:varbinary(512);"`
	RetryCount uint
	File       *File
	Account    *Account
	RemoveAt   sql.NullTime
	PushedAt   sql.NullTime
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TableName returns the entity database table name.
func (FilePush) TableName() string {
	return "files_push"
}

// NewFilePush creates a new entity.
func NewFilePush(fileID, accountID uint, remoteName string) *FilePush {
	result := &FilePush{
		FileID:     fileID,
		AccountID:  accountID,
		RemoteName: remoteName,
	}

	return result
}

// FirstOrCreate returns the matching entity or creates a new one.
func (m *FilePush) FirstOrCreate(db *gorm.DB) *FilePush {
	mutex.Db.Lock()
	defer mutex.Db.Unlock()

	if err := db.FirstOrCreate(m, "file_id = ? AND account_id = ? AND remote_name = ?", m.FileID, m.AccountID, m.RemoteName).Error; err != nil {
		log.Errorf("file push: %s", err)
	}

	return m
}
