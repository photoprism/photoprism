package entity

import (
	"time"
)

type FileShareMap map[string]FileShare

func (m FileShareMap) Get(name string, fileID, accountID uint, remoteName string) FileShare {
	if result, ok := m[name]; ok {
		return result
	}

	return *NewFileShare(fileID, accountID, remoteName)
}

func (m FileShareMap) Pointer(name string, fileID, accountID uint, remoteName string) *FileShare {
	if result, ok := m[name]; ok {
		return &result
	}

	return NewFileShare(fileID, accountID, remoteName)
}

var FileShareFixtures = FileShareMap{
	"FileShare1": {
		FileID:     1000000,
		AccountID:  1000000,
		RemoteName: "name for remote",
		Status:     FileShareShared,
		Error:      "",
		Errors:     0,
		File:       nil,
		Account:    &AccountFixtureWebdavDummy,
		CreatedAt:  time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	"FileShare2": {
		FileID:     1000000,
		AccountID:  1000001,
		RemoteName: "name for remote",
		Status:     FileShareNew,
		Error:      "",
		Errors:     0,
		File:       nil,
		Account:    &AccountFixtureWebdavDummy2,
		CreatedAt:  time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	},
}

// CreateFileShareFixtures inserts known entities into the database for testing.
func CreateFileShareFixtures() {
	for _, entity := range FileShareFixtures {
		Db().Create(&entity)
	}
}
