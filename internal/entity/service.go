package entity

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/service/webdav"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	SyncStatusRefresh  = "refresh"
	SyncStatusDownload = "download"
	SyncStatusUpload   = "upload"
	SyncStatusSynced   = "synced"
)

type Services []Service

// Service represents a remote service, e.g. for uploading, downloading or syncing media files.
//
// Field Descriptions:
// - AccTimeout configures the timeout for requests, options: "", high, medium, low, none.
// - AccErrors holds the number of connection errors since the last reset.
// - AccShare enables manual upload, see SharePath, ShareSize, and ShareExpires.
// - AccSync enables automatic file synchronization, see SyncDownload and SyncUpload.
// - RetryLimit specifies the number of retry attempts, a negative value disables the limit.
type Service struct {
	ID            uint   `gorm:"primary_key"`
	AccName       string `gorm:"type:VARCHAR(160);"`
	AccOwner      string `gorm:"type:VARCHAR(160);"`
	AccURL        string `gorm:"type:VARCHAR(255);"`
	AccType       string `gorm:"type:VARBINARY(255);"`
	AccKey        string `gorm:"type:VARBINARY(255);"`
	AccUser       string `gorm:"type:VARBINARY(255);"`
	AccPass       string `gorm:"type:VARBINARY(255);"`
	AccTimeout    string `gorm:"type:VARBINARY(16);"`
	AccError      string `gorm:"type:VARBINARY(512);"`
	AccErrors     int
	AccShare      bool
	AccSync       bool
	RetryLimit    int
	SharePath     string `gorm:"type:VARBINARY(1024);"`
	ShareSize     string `gorm:"type:VARBINARY(16);"`
	ShareExpires  int
	SyncPath      string `gorm:"type:VARBINARY(1024);"`
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

// TableName returns the entity table name.
func (Service) TableName() string {
	return "services"
}

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

// LogErr updates the service error count and message.
func (m *Service) LogErr(err error) error {
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
	return m.Updates(Service{AccError: m.AccError, AccErrors: m.AccErrors, AccShare: m.AccShare})
}

// ResetErrors resets the service and related file error messages and counters.
func (m *Service) ResetErrors(share, sync bool) error {
	if !share && !sync || Db().NewRecord(m) {
		return nil
	}

	if m.ID == 0 {
		return fmt.Errorf("invalid id")
	}

	if share {
		if err := Db().Model(FileShare{}).Where("service_id = ?", m.ID).Updates(Map{"error": "", "errors": 0}).Error; err != nil {
			return err
		}
	}

	if sync {
		if err := Db().Model(FileSync{}).Where("service_id = ?", m.ID).Updates(Map{"error": "", "errors": 0}).Error; err != nil {
			return err
		}
	}

	m.AccError = ""
	m.AccErrors = 0

	return m.Updates(Map{"acc_error": m.AccError, "acc_errors": m.AccErrors})
}

// SaveForm saves the entity using form data and stores it in the database.
func (m *Service) SaveForm(form form.Service) error {
	db := Db()

	// Copy model values from form.
	if err := deepcopier.Copy(m).From(form); err != nil {
		return err
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

	// Number of remote request retry attempts.
	if m.RetryLimit < -1 {
		m.RetryLimit = -1 // Disabled.
	} else if m.RetryLimit > 999 {
		m.RetryLimit = 999 // 999 retries max.
	}

	// Refresh after performing changes.
	if m.AccSync && m.SyncStatus == SyncStatusSynced {
		m.SyncStatus = SyncStatusRefresh
	}

	// Reset error counters if account already exists.
	if !Db().NewRecord(m) {
		Log("service", "reset errors", m.ResetErrors(m.AccShare, m.AccSync))
	}

	// Ensure account name and owner are not too long.
	m.AccName = txt.Clip(m.AccName, txt.ClipName)
	m.AccOwner = txt.Clip(m.AccOwner, txt.ClipName)

	// Save changes.
	return db.Save(m).Error
}

// Delete deletes the entity from the database.
func (m *Service) Delete() error {
	return Db().Delete(m).Error
}

// Directories returns a list of directories or albums in an account.
func (m *Service) Directories() (result fs.FileInfos, err error) {
	if m.AccType == service.WebDAV {
		var client *webdav.Client
		if client, err = webdav.NewClient(m.AccURL, m.AccUser, m.AccPass, webdav.Timeout(m.AccTimeout)); err != nil {
			return result, err
		}

		result, err = client.Directories("/", true, 0)
	}

	// Sort directory list.
	sort.Sort(result)

	// Update error count and message.
	if logErr := m.LogErr(err); logErr != nil {
		log.Warnf("service: %s", logErr)
	}

	return result, err
}

// Updates multiple columns in the database.
func (m *Service) Updates(values interface{}) error {
	return UnscopedDb().Model(m).UpdateColumns(values).Error
}

// Update a column in the database.
func (m *Service) Update(attr string, value interface{}) error {
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
