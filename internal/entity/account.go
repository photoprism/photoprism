package entity

import (
	"database/sql"
	"sort"
	"time"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/ulule/deepcopier"
)

const (
	AccountSyncStatusRefresh  = "refresh"
	AccountSyncStatusDownload = "download"
	AccountSyncStatusUpload   = "upload"
	AccountSyncStatusSynced   = "synced"
)

type Accounts []Account

// Account represents a remote service account for uploading, downloading or syncing media files.
type Account struct {
	ID            uint   `gorm:"primary_key"`
	AccName       string `gorm:"type:VARCHAR(160);"`
	AccOwner      string `gorm:"type:VARCHAR(160);"`
	AccURL        string `gorm:"type:VARBINARY(512);"`
	AccType       string `gorm:"type:VARBINARY(255);"`
	AccKey        string `gorm:"type:VARBINARY(255);"`
	AccUser       string `gorm:"type:VARBINARY(255);"`
	AccPass       string `gorm:"type:VARBINARY(255);"`
	AccError      string `gorm:"type:VARBINARY(512);"`
	AccErrors     int
	AccShare      bool
	AccSync       bool
	RetryLimit    int
	SharePath     string `gorm:"type:VARBINARY(500);"`
	ShareSize     string `gorm:"type:VARBINARY(16);"`
	ShareExpires  int
	SyncPath      string `gorm:"type:VARBINARY(500);"`
	SyncStatus    string `gorm:"type:VARBINARY(16);"`
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

	err = model.SaveForm(form)

	return model, err
}

// SaveForm saves the entity using form data and stores it in the database.
func (m *Account) SaveForm(form form.Account) error {
	db := Db()

	if err := deepcopier.Copy(m).From(form); err != nil {
		return err
	}

	// TODO: Support for other remote services in addition to WebDAV.
	if m.AccType != remote.ServiceWebDAV {
		m.AccShare = false
		m.AccSync = false
	}

	// Prevent two-way sync, see https://github.com/photoprism/photoprism/issues/1785
	if m.SyncUpload && m.SyncDownload {
		m.SyncUpload = false
	}

	// Set defaults.
	if m.SharePath == "" {
		m.SharePath = "/"
	}

	if m.SyncPath == "" {
		m.SyncPath = "/"
	}

	// Refresh after performing changes.
	if m.AccSync && m.SyncStatus == AccountSyncStatusSynced {
		m.SyncStatus = AccountSyncStatusRefresh
	}

	m.AccName = txt.Clip(m.AccName, txt.ClipName)
	m.AccOwner = txt.Clip(m.AccOwner, txt.ClipName)

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
		result, err = c.Directories("/", true, webdav.SyncTimeout)
	}

	sort.Sort(result)

	return result, err
}

// Updates multiple columns in the database.
func (m *Account) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Update a column in the database.
func (m *Account) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Save updates the existing or inserts a new row.
func (m *Account) Save() error {
	return Db().Save(m).Error
}

// Create inserts a new row to the database.
func (m *Account) Create() error {
	return Db().Create(m).Error
}

// ShareOriginals tests if the unmodified originals should be shared.
func (m *Account) ShareOriginals() bool {
	return m.ShareSize == ""
}
