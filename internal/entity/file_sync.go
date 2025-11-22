package entity

import (
	"time"
)

const (
	FileSyncNew        = "new"
	FileSyncIgnore     = "ignore"
	FileSyncExists     = "exists"
	FileSyncFailed     = "failed"
	FileSyncDownloaded = "downloaded"
	FileSyncUploaded   = "uploaded"
)

// FileSync tracks the synchronization status for a file on an external service.
type FileSync struct {
	RemoteName string    `gorm:"primary_key;auto_increment:false;type:VARBINARY(255)" json:"RemoteName" yaml:"RemoteName,omitempty"`
	ServiceID  uint      `gorm:"primary_key;auto_increment:false" json:"ServiceID" yaml:"ServiceID,omitempty"`
	FileID     uint      `gorm:"index;" json:"FileID" yaml:"FileID,omitempty"`
	RemoteDate time.Time `json:"RemoteDate,omitempty" yaml:"RemoteDate,omitempty"`
	RemoteSize int64     `json:"RemoteSize,omitempty" yaml:"RemoteSize,omitempty"`
	Status     string    `gorm:"type:VARBINARY(16);" json:"Status" yaml:"Status,omitempty"`
	Error      string    `gorm:"type:VARBINARY(512);" json:"Error,omitempty" yaml:"Error,omitempty"`
	Errors     int       `json:"Errors,omitempty" yaml:"Errors,omitempty"`
	File       *File     `json:"File,omitempty" yaml:"-"`
	Account    *Service  `json:"Account,omitempty" yaml:"-"`
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
func (m *FileSync) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Update mutates a single column on the existing row.
func (m *FileSync) Update(attr string, value interface{}) error {
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
		log.Errorf("file-sync: %s", err)
		return nil
	}

	return m
}
