package entity

import (
	"time"
)

// FolderMap is the type to hold Folder test data
type FolderMap map[string]Folder

// Get returns an entry from the FolderFixtures, or if not found a Folder with the Path set.
func (m FolderMap) Get(name string) Folder {
	if result, ok := m[name]; ok {
		return result
	}

	return Folder{Path: name}
}

// Pointer returns an entry from the FolderFixtures, or if not found a Folder with the Path set.
func (m FolderMap) Pointer(name string) *Folder {
	if result, ok := m[name]; ok {
		return &result
	}

	return &Folder{Path: name}
}

var deletedAt = time.Date(2026, 8, 19, 2, 16, 12, 0, time.UTC)

// FolderFixtures is the map of test data for Folder
var FolderFixtures = FolderMap{
	"1990": {
		FolderUID:         "dqo63pn35k2d495z",
		Path:              "1990",
		Root:              "/",
		FolderType:        "",
		FolderTitle:       "1990",
		FolderCategory:    "",
		FolderDescription: "",
		FolderOrder:       "name",
		FolderYear:        1990,
		FolderMonth:       7,
		FolderDay:         0,
		FolderCountry:     UnknownID,
		FolderFavorite:    false,
		FolderPrivate:     false,
		FolderIgnore:      false,
		FolderWatch:       false,
		CreatedAt:         time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt:         time.Date(2020, 3, 28, 14, 6, 0, 0, time.UTC),
		ModifiedAt:        time.Date(2020, 3, 20, 14, 6, 0, 0, time.UTC),
		DeletedAt:         nil,
	},
	"1990/04": {
		FolderUID:         "dqo63pn2f87f02xj",
		Path:              "1990/04",
		Root:              "/",
		FolderType:        "",
		FolderTitle:       "April 1990",
		FolderCategory:    "",
		FolderDescription: "",
		FolderOrder:       "name",
		FolderYear:        1990,
		FolderMonth:       4,
		FolderDay:         0,
		FolderCountry:     UnknownID,
		FolderFavorite:    false,
		FolderPrivate:     false,
		FolderIgnore:      false,
		FolderWatch:       false,
		CreatedAt:         time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt:         time.Date(2020, 3, 28, 14, 6, 0, 0, time.UTC),
		ModifiedAt:        time.Date(2020, 3, 20, 14, 6, 0, 0, time.UTC),
		DeletedAt:         nil,
	},
	"2007/12": {
		FolderUID:         "dqo63pn2f87f02oi",
		Path:              "2007/12",
		Root:              "/",
		FolderType:        "",
		FolderTitle:       "December 2007",
		FolderCategory:    "",
		FolderDescription: "",
		FolderOrder:       "name",
		FolderYear:        2007,
		FolderMonth:       12,
		FolderDay:         0,
		FolderCountry:     "de",
		FolderFavorite:    false,
		FolderPrivate:     false,
		FolderIgnore:      false,
		FolderWatch:       false,
		CreatedAt:         time.Date(2007, 12, 25, 2, 6, 51, 0, time.UTC),
		UpdatedAt:         time.Date(2020, 3, 30, 14, 6, 0, 0, time.UTC),
		ModifiedAt:        time.Date(2020, 3, 20, 14, 6, 0, 0, time.UTC),
		DeletedAt:         nil,
	},
	"2156/12": {
		FolderUID:         "dtjzq269o56db218",
		Path:              "2156/12",
		Root:              "/",
		FolderType:        "",
		FolderTitle:       "December 2156",
		FolderCategory:    "",
		FolderDescription: "",
		FolderOrder:       "name",
		FolderYear:        2156,
		FolderMonth:       12,
		FolderDay:         0,
		FolderCountry:     "de",
		FolderFavorite:    false,
		FolderPrivate:     false,
		FolderIgnore:      false,
		FolderWatch:       false,
		CreatedAt:         time.Date(2026, 8, 19, 2, 6, 51, 0, time.UTC),
		UpdatedAt:         time.Date(2026, 8, 19, 2, 6, 59, 0, time.UTC),
		ModifiedAt:        time.Date(2026, 8, 19, 2, 16, 12, 0, time.UTC),
		DeletedAt:         &deletedAt,
	},
}

// CreateFolderFixtures inserts known entities into the database for testing.
func CreateFolderFixtures() {
	for _, entity := range FolderFixtures {
		fixtureDb().Create(&entity)
	}
}
