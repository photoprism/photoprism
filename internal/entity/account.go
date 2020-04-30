package entity

import (
	"database/sql"
	"sort"
	"time"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/ulule/deepcopier"
)

const (
	AccountSyncStatusRefresh  = "refresh"
	AccountSyncStatusDownload = "download"
	AccountSyncStatusUpload   = "upload"
	AccountSyncStatusSynced   = "synced"
)

// Account represents a remote service account for uploading, downloading or syncing media files.
type Account struct {
	ID            uint   `gorm:"primary_key"`
	AccName       string `gorm:"type:varchar(255);"`
	AccOwner      string `gorm:"type:varchar(255);"`
	AccURL        string `gorm:"type:varbinary(512);"`
	AccType       string `gorm:"type:varbinary(255);"`
	AccKey        string `gorm:"type:varbinary(255);"`
	AccUser       string `gorm:"type:varbinary(255);"`
	AccPass       string `gorm:"type:varbinary(255);"`
	AccError      string `gorm:"type:varbinary(512);"`
	AccErrors     int
	AccShare      bool
	AccSync       bool
	RetryLimit    int
	SharePath     string `gorm:"type:varbinary(255);"`
	ShareSize     string `gorm:"type:varbinary(16);"`
	ShareExpires  int
	SyncPath      string `gorm:"type:varbinary(255);"`
	SyncStatus    string `gorm:"type:varbinary(16);"`
	SyncInterval  int
	SyncDate      sql.NullTime `deepcopier:"skip"`
	SyncUpload    bool
	SyncDownload  bool
	SyncFilenames bool
	SyncRaw       bool
	CreatedAt     time.Time  `deepcopier:"skip"`
	UpdatedAt     time.Time  `deepcopier:"skip"`
	DeletedAt     *time.Time `deepcopier:"skip" sql:"index"`
}

// CreateAccount creates a new account entity in the database.
func CreateAccount(form form.Account) (model *Account, err error) {
	model = &Account{
		ShareSize:    "",
		ShareExpires: 0,
		RetryLimit:   3,
		SyncStatus:   AccountSyncStatusRefresh,
	}

	err = model.Save(form)

	return model, err
}

// Save updates the entity using form data and stores it in the database.
func (m *Account) Save(form form.Account) error {
	db := Db()

	if err := deepcopier.Copy(m).From(form); err != nil {
		return err
	}

	if m.AccType != string(remote.ServiceWebDAV) {
		// TODO: Only WebDAV supported at the moment
		m.AccShare = false
		m.AccSync = false
	}

	// Set defaults
	if m.SharePath == "" {
		m.SharePath = "/"
	}

	if m.SyncPath == "" {
		m.SyncPath = "/"
	}

	// Refresh after performing changes
	if m.AccSync && m.SyncStatus == AccountSyncStatusSynced {
		m.SyncStatus = AccountSyncStatusRefresh
	}

	return db.Save(m).Error
}

// Delete deletes the entity from the database.
func (m *Account) Delete() error {
	return Db().Delete(m).Error
}

// Directories returns a list of directories or albums in an account.
func (m *Account) Directories() (result fs.FileInfos, err error) {
	if m.AccType == remote.ServiceWebDAV {
		c := webdav.New(m.AccURL, m.AccUser, m.AccPass)
		result, err = c.Directories("/", true)
	}

	sort.Sort(result)

	return result, err
}
