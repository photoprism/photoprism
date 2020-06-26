package entity

import (
	"database/sql"
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
		ShareExpires:  1,
		SyncPath:      "/Photos",
		SyncStatus:    "",
		SyncInterval:  3600,
		SyncDate:      sql.NullTime{Time: Timestamp()},
		SyncUpload:    true,
		SyncDownload:  true,
		SyncFilenames: true,
		SyncRaw:       true,
		CreatedAt:     Timestamp(),
		UpdatedAt:     Timestamp(),
		DeletedAt:     nil,
	},
	"webdav-dummy2": {
		ID:            1000001,
		AccName:       "Test Account2",
		AccOwner:      "",
		AccURL:        "http://webdav-dummy/",
		AccType:       "webdav",
		AccKey:        "123",
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
		SyncStatus:    "test",
		SyncInterval:  3600,
		SyncDate:      sql.NullTime{Time: Timestamp()},
		SyncUpload:    true,
		SyncDownload:  true,
		SyncFilenames: true,
		SyncRaw:       true,
		CreatedAt:     Timestamp(),
		UpdatedAt:     Timestamp(),
		DeletedAt:     nil,
	},
}

var AccountFixtureWebdavDummy = AccountFixtures["webdav-dummy"]
var AccountFixtureWebdavDummy2 = AccountFixtures["webdav-dummy2"]

// CreateLabelFixtures inserts known entities into the database for testing.
func CreateAccountFixtures() {
	for _, entity := range AccountFixtures {
		Db().Create(&entity)
	}
}
