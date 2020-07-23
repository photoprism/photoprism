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
		AccountID:  1000000,
		RemoteName: "name for remote sync",
		Status:     "test",
		Error:      "",
		Errors:     0,
		File:       &FileFixturesExampleJPG,
		Account:    &AccountFixtureWebdavDummy,
		RemoteDate: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		RemoteSize: int64(888),
		CreatedAt:  time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	"FileSync2": {
		FileID:     1000000,
		AccountID:  1000001,
		RemoteName: "name for remote sync",
		Status:     "test",
		Error:      "",
		Errors:     0,
		File:       &FileFixturesExampleJPG,
		Account:    &AccountFixtureWebdavDummy2,
		RemoteDate: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		RemoteSize: int64(60),
		CreatedAt:  time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	"FileSync3": {
		FileID:     1000000,
		AccountID:  1000000,
		RemoteName: "name for remote sync",
		Status:     "new",
		Error:      "",
		Errors:     0,
		File:       &FileFixturesExampleJPG,
		Account:    &AccountFixtureWebdavDummy,
		RemoteDate: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		RemoteSize: int64(60),
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
