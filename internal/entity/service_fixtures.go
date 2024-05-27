package entity

import (
	"database/sql"
)

type ServiceMap map[string]Service

var ServiceFixtures = ServiceMap{
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
		SyncDate:      sql.NullTime{Time: Now()},
		SyncUpload:    true,
		SyncDownload:  true,
		SyncFilenames: true,
		SyncRaw:       true,
		CreatedAt:     Now(),
		UpdatedAt:     Now(),
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
		SyncDate:      sql.NullTime{Time: Now()},
		SyncUpload:    true,
		SyncDownload:  true,
		SyncFilenames: true,
		SyncRaw:       true,
		CreatedAt:     Now(),
		UpdatedAt:     Now(),
		DeletedAt:     nil,
	},
}

var ServiceFixtureWebdavDummy = ServiceFixtures["dummy-webdav"]
var ServiceFixtureWebdavDummy2 = ServiceFixtures["dummy-webdav2"]

// CreateServiceFixtures inserts known entities into the database for testing.
func CreateServiceFixtures() {
	for _, entity := range ServiceFixtures {
		Db().Create(&entity)
	}
}
