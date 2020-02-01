package entity

import (
	"database/sql"
	"time"
)

// Account represents a remote service account for uploading, downloading or syncing media files.
type Account struct {
	ID           uint   `gorm:"primary_key"`
	Name         string `gorm:"type:varchar(128);"`
	URL          string `gorm:"type:varbinary(512);"`
	Protocol     string `gorm:"type:varbinary(256);"`
	ApiKey       string `gorm:"type:varbinary(256);"`
	Username     string `gorm:"type:varbinary(256);"`
	Password     string `gorm:"type:varbinary(256);"`
	LastError    string `gorm:"type:varbinary(256);"`
	IgnoreErrors bool
	PushSize     string `gorm:"type:varbinary(16);"`
	PushExif     bool
	PushDelete   bool
	PushSidecar  bool
	SyncPush     bool
	SyncPull     bool
	SyncPaused   int
	SyncInterval int
	SyncRetry    int
	SyncedAt     sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `sql:"index"`
}
