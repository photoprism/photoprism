package migrate

import (
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

var versionOnce = sync.Once{}
var versionMutex = sync.Mutex{}

// Version represents the application version.
type Version struct {
	ID         uint       `gorm:"primary_key" yaml:"-"`
	Version    string     `gorm:"size:255;unique_index:idx_version_edition;" json:"Version" yaml:"Version,omitempty"`
	Edition    string     `gorm:"size:255;unique_index:idx_version_edition;" json:"Edition" yaml:"Edition,omitempty"`
	Error      string     `gorm:"size:255;" json:"Error" yaml:"Error,omitempty"`
	CreatedAt  time.Time  `yaml:"CreatedAt,omitempty"`
	UpdatedAt  time.Time  `yaml:"UpdatedAt,omitempty"`
	MigratedAt *time.Time `json:"MigratedAt" yaml:"MigratedAt,omitempty"`
}

// TableName returns the entity database table name.
func (Version) TableName() string {
	return "versions"
}

var UnknownVersion = Version{
	Version: "0.0.0",
	Edition: "dev",
}

// NeedsMigration tests if the Version is not yet installed.
func (m *Version) NeedsMigration() bool {
	if m == nil {
		return true
	} else if m.MigratedAt == nil || m.CreatedAt.IsZero() {
		return true
	} else if m.Unknown() {
		return true
	}

	return m.MigratedAt.IsZero()
}

// Migrated flags the version as installed and migrated.
func (m *Version) Migrated(db *gorm.DB) error {
	if err := m.CreateTable(db); err != nil {
		return err
	} else if m.Unknown() {
		return nil
	}

	timeStamp := time.Now().UTC().Truncate(time.Second)
	m.MigratedAt = &timeStamp
	m.Error = ""

	return db.Model(m).Updates(Map{"MigratedAt": m.MigratedAt, "Error": m.Error}).Error
}

// NewVersion creates a Version entity from a model name and a make name.
func NewVersion(version, edition string) *Version {
	result := &Version{
		Version: txt.Clip(version, txt.ClipLongName),
		Edition: txt.Clip(edition, txt.ClipLongName),
	}

	return result
}

// Create inserts a new row to the database.
func (m *Version) Create(db *gorm.DB) error {
	if err := m.CreateTable(db); err != nil {
		return err
	}

	versionMutex.Lock()
	defer versionMutex.Unlock()

	return db.Create(m).Error
}

// Save saved the record in the database.
func (m *Version) Save(db *gorm.DB) error {
	if err := m.CreateTable(db); err != nil {
		return err
	}

	versionMutex.Lock()
	defer versionMutex.Unlock()

	return db.Save(m).Error
}

// Find fetches an existing record from the database.
func (m *Version) Find(db *gorm.DB) *Version {
	if err := m.CreateTable(db); err != nil {
		return nil
	} else if m.Version == "" {
		return nil
	}

	result := Version{}

	if err := db.Where("version = ? AND edition = ?", m.Version, m.Edition).First(&result).Error; err != nil {
		return nil
	}

	return &result
}

// FirstOrCreateVersion returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateVersion(db *gorm.DB, m *Version) *Version {
	if m == nil {
		return &UnknownVersion
	} else if err := m.CreateTable(db); err != nil {
		return &UnknownVersion
	} else if m.Version == "" {
		return &UnknownVersion
	}

	// Find existing version or create new record.
	if found := m.Find(db); found != nil {
		return found
	} else if err := m.Create(db); err != nil {
		log.Errorf("version: %s (create)", err)
	} else {
		return m
	}

	return &UnknownVersion
}

// String returns an identifier that can be used in logs.
func (m *Version) String() string {
	return clean.Log(m.Version + "-" + m.Edition)
}

// Unknown checks if the version is unknown.
func (m *Version) Unknown() bool {
	if m == nil {
		return true
	} else if m.Version == "" {
		return true
	}

	return m.Version == UnknownVersion.Version
}

// CreateTable creates the versions database table if needed.
func (m *Version) CreateTable(db *gorm.DB) (err error) {
	if db == nil {
		return fmt.Errorf("db is nil")
	} else if m == nil {
		return fmt.Errorf("version is nil")
	}

	versionOnce.Do(func() {
		err = db.AutoMigrate(&Version{}).Error
	})

	return err
}
