package entity

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/service/webdav"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Status values a service sync moves through.
const (
	SyncStatusRefresh  = "refresh"
	SyncStatusDownload = "download"
	SyncStatusUpload   = "upload"
	SyncStatusSynced   = "synced"
)

// Services represents a list of remote services.
type Services []Service

// Service represents a remote service, e.g. for uploading, downloading or syncing media files.
//
// Field Descriptions:
// - AccTimeout configures the timeout for requests, options: "", high, medium, low, none.
// - AccErrors holds the number of connection errors since the last reset.
// - AccShare enables manual upload, see SharePath, ShareSize, and ShareExpires.
// - AccSync enables automatic file synchronization, see SyncDownload and SyncUpload.
// - RetryLimit specifies the number of retry attempts, 0 or a negative value disables the limit.
// - SyncYaml controls transferring YAML sidecar files: -1 disabled, 0 default (enabled), 1 enabled.
type Service struct {
	ID            uint         `gorm:"primaryKey;size:32" json:"ID"`
	AccName       string       `gorm:"size:160;" json:"AccName"`
	AccOwner      string       `gorm:"size:160;" json:"AccOwner"`
	AccURL        string       `gorm:"size:255;" json:"AccURL"`
	AccType       string       `gorm:"type:bytes;size:255;" json:"AccType"`
	AccKey        string       `gorm:"type:bytes;size:255;" json:"-"`
	AccUser       string       `gorm:"type:bytes;size:255;" json:"AccUser"`
	AccPass       string       `gorm:"type:bytes;size:255;" json:"-"`
	AccTimeout    string       `gorm:"type:bytes;size:16;" json:"AccTimeout"`
	AccError      string       `gorm:"type:bytes;size:512;" json:"AccError"`
	AccErrors     int          `gorm:"size:32" json:"AccErrors"`
	AccShare      bool         `json:"AccShare"`
	AccSync       bool         `json:"AccSync"`
	RetryLimit    int          `gorm:"size:32" json:"RetryLimit"`
	SharePath     string       `gorm:"type:bytes;size:1024;" json:"SharePath"`
	ShareSize     string       `gorm:"type:bytes;size:16;" json:"ShareSize"`
	ShareExpires  int          `gorm:"size:32" json:"ShareExpires"`
	SyncPath      string       `gorm:"type:bytes;size:1024;" json:"SyncPath"`
	SyncStatus    string       `gorm:"type:bytes;size:16;" json:"SyncStatus"`
	SyncInterval  int          `gorm:"size:32" json:"SyncInterval"`
	SyncDate      sql.NullTime `deepcopier:"skip" json:"SyncDate"`
	SyncUpload    bool         `json:"SyncUpload"`
	SyncDownload  bool         `json:"SyncDownload"`
	SyncFilenames bool         `json:"SyncFilenames"`
	SyncRaw       bool         `json:"SyncRaw"`
	SyncYaml      int          `gorm:"type:int;size:16;default:0;" json:"SyncYaml"`
	CreatedAt     time.Time    `deepcopier:"skip" json:"CreatedAt"`
	UpdatedAt     time.Time    `deepcopier:"skip" json:"UpdatedAt"`
	DeletedAt     DeletedAt    `gorm:"index" deepcopier:"skip" json:"DeletedAt"`
}

// TableName returns the entity table name.
func (Service) TableName() string {
	return "services"
}

// ErrServiceFolderTooLong is returned when a sync or share folder leaves no room for the names of files below it,
// or a folder name in the sync path is longer than local file systems accept.
var ErrServiceFolderTooLong = errors.New("folder name is too long")

// AddService adds a new service account to the database.
func AddService(form form.Service) (model *Service, err error) {
	model = &Service{
		ShareSize:    "",
		ShareExpires: 0,
		RetryLimit:   3,
		AccTimeout:   string(webdav.TimeoutDefault),
		SyncStatus:   SyncStatusRefresh,
	}

	err = model.SaveForm(form)

	return model, err
}

// ResetErrors resets the service and related file error messages and counters, keeping per-file download error counts without a retry limit.
func (m *Service) ResetErrors(share, sync bool) error {
	if !share && !sync || isNewRecord(Db(), m) {
		return nil
	}

	if m.ID == 0 {
		return fmt.Errorf("invalid id")
	}

	if share {
		if err := Db().Model(&FileShare{}).Where("service_id = ?", m.ID).Updates(Values{"error": "", "errors": 0}).Error; err != nil {
			return err
		}
	}

	if sync {
		values := Values{"error": "", "errors": 0}

		// Without a retry limit, the count only orders the download queue, so it is kept.
		if m.RetryLimit <= 0 {
			values = Values{"error": ""}
		}

		if err := Db().Model(&FileSync{}).Where("service_id = ?", m.ID).Updates(values).Error; err != nil {
			return err
		}
	}

	m.AccError = ""
	m.AccErrors = 0

	return m.Updates(Values{"acc_error": m.AccError, "acc_errors": m.AccErrors})
}

// SaveForm saves the entity using form data and stores it in the database.
func (m *Service) SaveForm(form form.Service) error {
	db := Db()
	stored, isNew := *m, isNewRecord(db, m)

	// Copy model values from form.
	if err := deepcopier.Copy(m).From(form); err != nil {
		return err
	}

	// The account error is written by the sync worker, so a request may only clear it.
	if form.AccError != "" {
		m.AccError = stored.AccError
	}

	// Remove control characters, except from the password, whose bytes are the credential.
	for _, s := range []*string{&m.AccName, &m.AccOwner, &m.AccURL, &m.AccType, &m.AccKey, &m.AccUser,
		&m.AccTimeout, &m.SharePath, &m.ShareSize, &m.SyncPath} {
		*s = clean.StripControl(*s)
	}

	// TODO: Support for other remote services in addition to WebDAV.
	if m.AccType != service.WebDAV {
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

	// Remote names of files below the folders must fit their records, and downloaded file and folder names local
	// file systems. A stored folder is checked only when a save changes it or turns on what it is used for.
	storedShare, storedSync := serviceFolder(stored.SharePath), serviceFolder(stored.SyncPath)
	shareChanged := isNew || m.SharePath != storedShare || m.AccShare && !stored.AccShare
	syncChanged := isNew || m.SyncPath != storedSync || m.AccSync && !stored.AccSync
	downloadChanged := syncChanged || m.SyncDownload && !stored.SyncDownload

	if shareChanged && len(m.SharePath) >= RemoteNameMaxLen || syncChanged && len(m.SyncPath) >= RemoteNameMaxLen ||
		downloadChanged && m.SyncDownload && fs.NameTooLong(m.SyncPath) {
		return ErrServiceFolderTooLong
	}

	// Number of remote request retry attempts.
	if m.RetryLimit < -1 || m.RetryLimit == 0 {
		m.RetryLimit = -1 // Disabled.
	} else if m.RetryLimit > 999 {
		m.RetryLimit = 999 // 999 retries max.
	}

	// Automatic sync interval in seconds, where 0 means never.
	if m.SyncInterval < 0 {
		m.SyncInterval = 0
	} else if m.SyncInterval > 31536000 {
		m.SyncInterval = 31536000 // One year max.
	}

	// Refresh after performing changes, and in any stage when the folder or a direction changed, or raw and video
	// files were turned on, also while sync is off.
	if m.AccSync && m.SyncStatus == SyncStatusSynced || !isNew && (m.SyncPath != storedSync ||
		m.SyncDownload != stored.SyncDownload || m.SyncUpload != stored.SyncUpload || m.SyncRaw && !stored.SyncRaw) {
		m.SyncStatus = SyncStatusRefresh
	}

	// Reset error counters if account already exists.
	if !isNew {
		Log("service", "reset errors", m.ResetErrors(m.AccShare, m.AccSync))
	}

	// Ensure account name and owner are not too long.
	m.AccName = txt.Clip(m.AccName, txt.ClipName)
	m.AccOwner = txt.Clip(m.AccOwner, txt.ClipName)

	// Limit the YAML sidecar option to disabled, default, and enabled.
	if m.SyncYaml < -1 {
		m.SyncYaml = -1
	} else if m.SyncYaml > 1 {
		m.SyncYaml = 1
	}

	// Save changes.
	return db.Save(m).Error
}

// serviceFolder returns a stored share or sync folder as a save normalizes it, so an unchanged folder compares equal.
func serviceFolder(s string) string {
	if s = clean.StripControl(s); s == "" {
		return "/"
	}

	return s
}

// SyncYamlEnabled reports whether YAML sidecar files are transferred, which is the default.
func (m *Service) SyncYamlEnabled() bool {
	return m.SyncYaml >= 0
}

// Delete deletes the entity from the database.
func (m *Service) Delete() error {
	return Db().Delete(m).Error
}

// Directories returns a list of directories or albums in an account.
func (m *Service) Directories(servicesCIDR string) (result fs.FileInfos, err error) {
	if m.AccType == service.WebDAV {
		var client *webdav.Client
		if client, err = webdav.NewClient(m.AccURL, m.AccUser, m.AccPass, webdav.Timeout(m.AccTimeout), servicesCIDR); err != nil {
			return result, err
		}

		result, err = client.Directories("/", true, 0)
	}

	// Sort directory list.
	sort.Sort(result)

	return result, err
}

// Updates multiple columns in the database.
func (m *Service) Updates(values any) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Update a column in the database.
func (m *Service) Update(attr string, value any) error {
	return UnscopedDb().Model(m).UpdateColumn(attr, value).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Service) Save() error {
	return Db().Save(m).Error
}

// Create inserts a new row to the database.
func (m *Service) Create() error {
	return Db().Create(m).Error
}

// ShareOriginals tests if the unmodified originals should be shared.
func (m *Service) ShareOriginals() bool {
	return m.ShareSize == ""
}
