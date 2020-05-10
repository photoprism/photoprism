package entity

import (
	"database/sql"
	"time"
)

type AccountMap map[string]Account

var AccountFixtures = AccountMap{
	"webdav-dummy": {
		ID:            1000000,
		AccName:       "Test Account",
		AccOwner:      "",
		AccURL:        "http://webdav-dummy/",
		AccType:       "webdav",
		AccKey:        "",
		AccUser:       "admin",
		AccPass:       "photoprism",
		AccError:      "",
		AccErrors:     0,
		AccShare:      true,
		AccSync:       true,
		RetryLimit:    3,
		SharePath:     "/Photos",
		ShareSize:     "",
		ShareExpires:  0,
		SyncPath:      "/Photos",
		SyncStatus:    "",
		SyncInterval:  3600,
		SyncDate:      sql.NullTime{Time: time.Now()},
		SyncUpload:    true,
		SyncDownload:  true,
		SyncFilenames: true,
		SyncRaw:       true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		DeletedAt:     nil,
	},
}

// CreateLabelFixtures inserts known entities into the database for testing.
func CreateAccountFixtures() {
	for _, entity := range AccountFixtures {
		Db().Create(&entity)
	}
}
