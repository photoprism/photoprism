package entity

import (
	"time"
)

type FileSyncMap map[string]FileSync

func (m FileSyncMap) Get(name string, accountID uint, remoteName string) FileSync {
	if result, ok := m[name]; ok {
		return result
	}

	return *NewFileSync(accountID, remoteName)
}

func (m FileSyncMap) Pointer(name string, accountID uint, remoteName string) *FileSync {
	if result, ok := m[name]; ok {
		return &result
	}

	return NewFileSync(accountID, remoteName)
}

var FileSyncFixtures = FileSyncMap{
	"FileSync1": {
		FileID:     1000000,
		ServiceID:  1000000,
		RemoteName: "/20200706-092527-Landscape-München-2020.jpg",
		Status:     "uploaded",
		Error:      "",
		Errors:     0,
		File:       &FileFixturesExampleJPG,
		Account:    &ServiceFixtureWebdavDummy,
		RemoteDate: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		RemoteSize: int64(888),
		CreatedAt:  time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	"FileSync2": {
		FileID:     1000000,
		ServiceID:  1000001,
		RemoteName: "/20200706-092527-Landscape-Hamburg-2020.jpg",
		Status:     "downloaded",
		Error:      "",
		Errors:     0,
		File:       &FileFixturesExampleJPG,
		Account:    &ServiceFixtureWebdavDummy2,
		RemoteDate: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		RemoteSize: int64(160),
		CreatedAt:  time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	"FileSync3": {
		FileID:     1000000,
		ServiceID:  1000000,
		RemoteName: "/20200706-092527-People-2020.jpg",
		Status:     "new",
		Error:      "",
		Errors:     0,
		File:       &FileFixturesExampleJPG,
		Account:    &ServiceFixtureWebdavDummy,
		RemoteDate: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		RemoteSize: int64(860),
		CreatedAt:  time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	},
}

// CreateFileSyncFixtures inserts known entities into the database for testing.
func CreateFileSyncFixtures() {
	for _, entity := range FileSyncFixtures {
		Db().Create(&entity)
	}
}
