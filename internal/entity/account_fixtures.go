package entity

import (
	"database/sql"
)

type AccountMap map[string]Account

var AccountFixtures = AccountMap{
	"dummy-webdav": {
		ID:            1000000,
		AccName:       "Test Account",
		AccOwner:      "",
		AccURL:        "http://dummy-webdav/",
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
		ShareExpires:  1,
		SyncPath:      "/Photos",
		SyncStatus:    "refresh",
		SyncInterval:  3600,
		SyncDate:      sql.NullTime{Time: TimeStamp()},
		SyncUpload:    true,
		SyncDownload:  true,
		SyncFilenames: true,
		SyncRaw:       true,
		CreatedAt:     TimeStamp(),
		UpdatedAt:     TimeStamp(),
		DeletedAt:     nil,
	},
	"dummy-webdav2": {
		ID:            1000001,
		AccName:       "Test Account2",
		AccOwner:      "",
		AccURL:        "http://dummy-webdav/",
		AccType:       "webdav",
		AccKey:        "",
		AccUser:       "admin",
		AccPass:       "photoprism",
		AccError:      "",
		AccErrors:     0,
		AccShare:      false,
		AccSync:       false,
		RetryLimit:    3,
		SharePath:     "/Photos",
		ShareSize:     "",
		ShareExpires:  0,
		SyncPath:      "/Photos",
		SyncStatus:    "refresh",
		SyncInterval:  3600,
		SyncDate:      sql.NullTime{Time: TimeStamp()},
		SyncUpload:    true,
		SyncDownload:  true,
		SyncFilenames: true,
		SyncRaw:       true,
		CreatedAt:     TimeStamp(),
		UpdatedAt:     TimeStamp(),
		DeletedAt:     nil,
	},
}

var AccountFixtureWebdavDummy = AccountFixtures["dummy-webdav"]
var AccountFixtureWebdavDummy2 = AccountFixtures["dummy-webdav2"]

// CreateLabelFixtures inserts known entities into the database for testing.
func CreateAccountFixtures() {
	for _, entity := range AccountFixtures {
		Db().Create(&entity)
	}
}
