package entity

import (
	"time"
)

const (
	FileSyncNew        = "new"
	FileSyncIgnore     = "ignore"
	FileSyncExists     = "exists"
	FileSyncDownloaded = "downloaded"
	FileSyncUploaded   = "uploaded"
)

// FileSync represents a one-to-many relation between File and Account for syncing with remote services.
type FileSync struct {
	RemoteName string `gorm:"primary_key;auto_increment:false;type:varbinary(255)"`
	AccountID  uint   `gorm:"primary_key;auto_increment:false"`
	FileID     uint   `gorm:"index;"`
	RemoteDate time.Time
	RemoteSize int64
	Status     string `gorm:"type:varbinary(16);"`
	Error      string `gorm:"type:varbinary(512);"`
	Errors     int
	File       *File
	Account    *Account
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
		Status:     FileSyncNew,
	}

	return result
}

// FirstOrCreate returns the matching entity or creates a new one.
func (m *FileSync) FirstOrCreate() *FileSync {
	if err := Db().FirstOrCreate(m, "account_id = ? AND remote_name = ?", m.AccountID, m.RemoteName).Error; err != nil {
		log.Errorf("file sync: %s", err)
	}

	return m
}
