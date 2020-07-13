package entity

import (
	"time"
)

type AlbumMap map[string]Album

func (m AlbumMap) Get(name string) Album {
	if result, ok := m[name]; ok {
		return result
	}

	return *NewAlbum(name, AlbumDefault)
}

func (m AlbumMap) Pointer(name string) *Album {
	if result, ok := m[name]; ok {
		return &result
	}

	return NewAlbum(name, AlbumDefault)
}

var AlbumFixtures = AlbumMap{
	"christmas2030": {
		ID:               1000000,
		CoverUID:         "",
		AlbumUID:         "at9lxuqxpogaaba7",
		AlbumSlug:        "christmas2030",
		AlbumType:        AlbumDefault,
		AlbumTitle:       "Christmas2030",
		AlbumDescription: "Wonderful christmas",
		AlbumNotes:       "",
		AlbumOrder:       "oldest",
		AlbumTemplate:    "",
		AlbumFavorite:    false,
		CreatedAt:        time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"holiday-2030": {
		ID:               1000001,
		CoverUID:         "",
		AlbumUID:         "at9lxuqxpogaaba8",
		AlbumSlug:        "holiday-2030",
		AlbumType:        AlbumDefault,
		AlbumTitle:       "Holiday2030",
		AlbumDescription: "Wonderful christmas",
		AlbumNotes:       "",
		AlbumOrder:       "newest",
		AlbumTemplate:    "",
		AlbumFavorite:    true,
		CreatedAt:        time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"berlin-2019": {
		ID:               1000002,
		CoverUID:         "",
		AlbumUID:         "at9lxuqxpogaaba9",
		AlbumSlug:        "berlin-2019",
		AlbumType:        AlbumDefault,
		AlbumTitle:       "Berlin2019",
		AlbumDescription: "Wonderful christmas",
		AlbumNotes:       "",
		AlbumOrder:       "oldest",
		AlbumTemplate:    "",
		AlbumFavorite:    false,
		CreatedAt:        time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"april-1990": {
		ID:               1000003,
		CoverUID:         "",
		AlbumUID:         "at1lxuqipogaaba1",
		AlbumSlug:        "april-1990",
		AlbumType:        AlbumFolder,
		AlbumTitle:       "April 1990",
		AlbumDescription: "Spring is the time of year when many things change.",
		AlbumNotes:       "Thunderstorms cause most of the severe spring weather.",
		AlbumOrder:       "oldest",
		AlbumTemplate:    "",
		AlbumFilter:      "path:\"1990/04\"",
		AlbumFavorite:    false,
		CreatedAt:        time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"import": {
		ID:               1000004,
		CoverUID:         "",
		AlbumUID:         "at6axuzitogaaiax",
		AlbumSlug:        "import",
		AlbumType:        AlbumDefault,
		AlbumTitle:       "Import Album",
		AlbumDescription: "",
		AlbumNotes:       "",
		AlbumOrder:       "name",
		AlbumTemplate:    "",
		AlbumFilter:      "",
		AlbumFavorite:    false,
		CreatedAt:        time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"emptyMoment": {
		ID:               1000005,
		CoverUID:         "",
		AlbumUID:         "at7axuzitogaaiax",
		AlbumSlug:        "empty-moment",
		AlbumType:        AlbumMoment,
		AlbumTitle:       "Empty Moment",
		AlbumCategory:    "Fun",
		AlbumLocation:    "Favorite Park",
		AlbumDescription: "",
		AlbumNotes:       "",
		AlbumOrder:       "name",
		AlbumTemplate:    "",
		AlbumFilter:      "",
		AlbumFavorite:    false,
		CreatedAt:        time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
}

// CreateAlbumFixtures inserts known entities into the database for testing.
func CreateAlbumFixtures() {
	for _, entity := range AlbumFixtures {
		Db().Create(&entity)
	}
}
