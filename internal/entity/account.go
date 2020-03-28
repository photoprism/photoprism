package entity

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/ulule/deepcopier"
)

// Account represents a remote service account for uploading, downloading or syncing media files.
type Account struct {
	ID           uint   `gorm:"primary_key"`
	AccName      string `gorm:"type:varchar(128);"`
	AccOwner     string `gorm:"type:varchar(128);"`
	AccURL       string `gorm:"type:varbinary(512);"`
	AccType      string `gorm:"type:varbinary(256);"`
	AccKey       string `gorm:"type:varbinary(256);"`
	AccUser      string `gorm:"type:varbinary(256);"`
	AccPass      string `gorm:"type:varbinary(256);"`
	AccError     string `gorm:"type:varbinary(512);"`
	AccShare     bool
	AccSync      bool
	RetryLimit   uint
	SharePath    string `gorm:"type:varbinary(256);"`
	ShareSize    string `gorm:"type:varbinary(16);"`
	ShareExpires uint
	ShareExif    bool
	ShareSidecar bool
	SyncPath     string `gorm:"type:varbinary(256);"`
	SyncInterval uint
	SyncUpload   bool
	SyncDownload bool
	SyncDelete   bool
	SyncRaw      bool
	SyncVideo    bool
	SyncSidecar  bool
	SyncStart    sql.NullTime
	SyncedAt     sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `sql:"index"`
}

// CreateAccount creates a new account entity in the database.
func CreateAccount(form form.Account, db *gorm.DB) (model *Account, err error) {
	model = &Account{}

	if err := deepcopier.Copy(&model).From(form); err != nil {
		return model, err
	}

	err = db.Save(&model).Error

	return model, err
}

// Save updates the entity using form data and stores it in the database.
func (m *Account) Save(form form.Account, db *gorm.DB) error {
	if err := deepcopier.Copy(m).From(form); err != nil {
		return err
	}

	return db.Save(m).Error
}

// Delete deletes the entity from the database.
func (m *Account) Delete(db *gorm.DB) error {
	return db.Delete(m).Error
}
