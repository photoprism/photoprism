package entity

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/remote"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
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
	AccTimeout    string `gorm:"type:VARBINARY(16);"` // Request timeout: default, high, medium, low, none
	AccError      string `gorm:"type:VARBINARY(512);"`
	AccErrors     int    // Number of general account errors, there are counters for individual files too.
	AccShare      bool   // Manual upload enabled, see SharePath, ShareSize, and ShareExpires.
	AccSync       bool   // Background sync enabled, see SyncDownload and SyncUpload.
	RetryLimit    int    // Number of remote request retry attempts.
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
		AccTimeout:   string(webdav.TimeoutDefault),
		SyncStatus:   AccountSyncStatusRefresh,
	}

	err = model.SaveForm(form)

	return model, err
}

// LogError updates the account error count and message.
func (m *Account) LogError(err error) error {
	if err == nil {
		return m.ResetErrors(true, true)
	}

	// Update error message and increase count.
	m.AccError = err.Error()
	m.AccErrors++

	// Disable sharing when retry limit is reached.
	if m.RetryLimit > 0 && m.AccErrors > m.RetryLimit {
		m.AccShare = false
	}

	// Update fields in database.
	return m.Updates(Account{AccError: m.AccError, AccErrors: m.AccErrors, AccShare: m.AccShare})
}

// ResetErrors resets the account and related file error messages and counters.
func (m *Account) ResetErrors(share, sync bool) error {
	if !share && !sync {
		return nil
	}

	if m.ID == 0 {
		return fmt.Errorf("invalid account id")
	}

	if share {
		if err := Db().Model(FileShare{}).Where("account_id = ?", m.ID).Updates(Values{"error": "", "errors": 0}).Error; err != nil {
			return err
		}
	}

	if sync {
		if err := Db().Model(FileSync{}).Where("account_id = ?", m.ID).Updates(Values{"error": "", "errors": 0}).Error; err != nil {
			return err
		}
	}

	m.AccError = ""
	m.AccErrors = 0

	return m.Updates(Values{"acc_error": m.AccError, "acc_errors": m.AccErrors})
}

// SaveForm saves the entity using form data and stores it in the database.
func (m *Account) SaveForm(form form.Account) error {
	db := Db()

	// Copy model values from form.
	if err := deepcopier.Copy(m).From(form); err != nil {
		return err
	}

	// TODO: Support for other remote services in addition to WebDAV.
	if m.AccType != remote.ServiceWebDAV {
		m.AccShare = false // Disable manual upload.
		m.AccSync = false  // Disable background sync.
	}

	// Prevent two-way sync, see https://github.com/photoprism/photoprism/issues/1785
	if m.SyncUpload && m.SyncDownload {
		m.SyncUpload = false
	}

	// Set default manual upload folder if empty.
	if m.SharePath == "" {
		m.SharePath = "/"
	}

	// Set default background sync folder if empty.
	if m.SyncPath == "" {
		m.SyncPath = "/"
	}

	// Number of remote request retry attempts.
	if m.RetryLimit < -1 {
		m.RetryLimit = -1 // Disabled.
	} else if m.RetryLimit > 999 {
		m.RetryLimit = 999 // 999 retries max.
	}

	// Refresh after performing changes.
	if m.AccSync && m.SyncStatus == AccountSyncStatusSynced {
		m.SyncStatus = AccountSyncStatusRefresh
	}

	// Reset share/sync errors.
	if err := m.ResetErrors(m.AccShare, m.AccSync); err != nil {
		log.Debugf("account: %s", err)
		log.Errorf("account: failed to reset errors")
	}

	// Ensure account name and owner are not too long.
	m.AccName = txt.Clip(m.AccName, txt.ClipName)
	m.AccOwner = txt.Clip(m.AccOwner, txt.ClipName)

	// Save changes.
	return db.Save(m).Error
}

// Delete deletes the entity from the database.
func (m *Account) Delete() error {
	return Db().Delete(m).Error
}

// Directories returns a list of directories or albums in an account.
func (m *Account) Directories() (result fs.FileInfos, err error) {
	if m.AccType == remote.ServiceWebDAV {
		client := webdav.New(m.AccURL, m.AccUser, m.AccPass, webdav.Timeout(m.AccTimeout))
		result, err = client.Directories("/", true, 0)
	}

	// Sort directory list.
	sort.Sort(result)

	// Update error count and message.
	if err := m.LogError(err); err != nil {
		log.Warnf("account: %s", err)
	}

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
