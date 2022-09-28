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
	RemoteName string `gorm:"primary_key;auto_increment:false;type:VARBINARY(255)"`
	AccountID  uint   `gorm:"primary_key;auto_increment:false"`
	FileID     uint   `gorm:"index;"`
	RemoteDate time.Time
	RemoteSize int64
	Status     string `gorm:"type:VARBINARY(16);"`
	Error      string `gorm:"type:VARBINARY(512);"`
	Errors     int
	File       *File
	Account    *Account
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TableName returns the entity table name.
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

// Updates multiple columns in the database.
func (m *FileSync) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Update a column in the database.
func (m *FileSync) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Save updates the existing or inserts a new row.
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

	if err := Db().Where("account_id = ? AND remote_name = ?", m.AccountID, m.RemoteName).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("file-sync: %s", err)
		return nil
	}

	return m
}
