package entity

import (
	"time"
)

var FolderFixtures = map[string]Folder{
	"1990": {
		FolderUID:     "dqo63pn35k2d495z",
		Path:          "1990",
		FolderYear:    1990,
		FolderMonth:   0,
		FolderCountry: "zz",
		CreatedAt:     time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt:     time.Date(2020, 3, 28, 14, 6, 0, 0, time.UTC),
		DeletedAt:     nil,
	},
	"1990/04": {
		FolderUID:     "dqo63pn2f87f02xj",
		Path:          "1990/04",
		FolderYear:    1990,
		FolderMonth:   4,
		FolderCountry: "zz",
		CreatedAt:     time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt:     time.Date(2020, 3, 28, 14, 6, 0, 0, time.UTC),
		DeletedAt:     nil,
	},
	"2007/12": {
		FolderUID:     "dqo63pn2f87f02oi",
		Path:          "2007/12",
		FolderYear:    2007,
		FolderMonth:   12,
		FolderCountry: "de",
		CreatedAt:     time.Date(2007, 12, 25, 2, 6, 51, 0, time.UTC),
		UpdatedAt:     time.Date(2020, 3, 30, 14, 6, 0, 0, time.UTC),
		DeletedAt:     nil,
	},
}

// CreateFolderFixtures inserts known entities into the database for testing.
func CreateFolderFixtures() {
	for _, entity := range FolderFixtures {
		Db().Create(&entity)
	}
}
