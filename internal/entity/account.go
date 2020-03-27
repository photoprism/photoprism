package entity

import (
	"database/sql"
	"time"
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
	AccPush      bool
	AccSync      bool
	RetryLimit   uint
	PushPath     string `gorm:"type:varbinary(256);"`
	PushSize     string `gorm:"type:varbinary(16);"`
	PushExpires  uint
	PushExif     bool
	PushSidecar  bool
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
