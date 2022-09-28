package entity

import (
	"time"
)

const (
	FileShareNew     = "new"
	FileShareError   = "error"
	FileShareShared  = "shared"
	FileShareRemoved = "removed"
)

// FileShare represents a one-to-many relation between File and Account for pushing files to remote services.
type FileShare struct {
	FileID     uint   `gorm:"primary_key;auto_increment:false"`
	AccountID  uint   `gorm:"primary_key;auto_increment:false"`
	RemoteName string `gorm:"primary_key;auto_increment:false;type:VARBINARY(255)"`
	Status     string `gorm:"type:VARBINARY(16);"`
	Error      string `gorm:"type:VARBINARY(512);"`
	Errors     int
	File       *File
	Account    *Account
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// TableName returns the entity table name.
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
		Errors:     0,
	}

	return result
}

// Updates multiple columns in the database.
func (m *FileShare) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Updates a column in the database.
func (m *FileShare) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Save updates the existing or inserts a new row.
func (m *FileShare) Save() error {
	return Db().Save(m).Error
}

// Create inserts a new row to the database.
func (m *FileShare) Create() error {
	return Db().Create(m).Error
}

// FirstOrCreateFileShare returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateFileShare(m *FileShare) *FileShare {
	result := FileShare{}

	if err := Db().Where("file_id = ? AND account_id = ? AND remote_name = ?", m.FileID, m.AccountID, m.RemoteName).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("file-share: %s", err)
		return nil
	}

	return m
}
