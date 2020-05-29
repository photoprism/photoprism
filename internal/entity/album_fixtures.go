package entity

import (
	"time"
)

type AlbumMap map[string]Album

func (m AlbumMap) Get(name string) Album {
	if result, ok := m[name]; ok {
		return result
	}

	return *NewAlbum(name, TypeAlbum)
}

func (m AlbumMap) Pointer(name string) *Album {
	if result, ok := m[name]; ok {
		return &result
	}

	return NewAlbum(name, TypeAlbum)
}

var AlbumFixtures = AlbumMap{
	"christmas2030": {
		ID:               1000000,
		CoverUID:         "",
		AlbumUID:         "at9lxuqxpogaaba7",
		AlbumSlug:        "christmas2030",
		AlbumType:        TypeAlbum,
		AlbumTitle:       "Christmas2030",
		AlbumDescription: "Wonderful christmas",
		AlbumNotes:       "",
		AlbumOrder:       "oldest",
		AlbumTemplate:    "",
		AlbumFavorite:    false,
		Links:            []Link{},
		CreatedAt:        time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"holiday-2030": {
		ID:               1000001,
		CoverUID:         "",
		AlbumUID:         "at9lxuqxpogaaba8",
		AlbumSlug:        "holiday-2030",
		AlbumType:        TypeAlbum,
		AlbumTitle:       "Holiday2030",
		AlbumDescription: "Wonderful christmas",
		AlbumNotes:       "",
		AlbumOrder:       "newest",
		AlbumTemplate:    "",
		AlbumFavorite:    true,
		Links:            []Link{},
		CreatedAt:        time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"berlin-2019": {
		ID:               1000002,
		CoverUID:         "",
		AlbumUID:         "at9lxuqxpogaaba9",
		AlbumSlug:        "berlin-2019",
		AlbumType:        TypeAlbum,
		AlbumTitle:       "Berlin2019",
		AlbumDescription: "Wonderful christmas",
		AlbumNotes:       "",
		AlbumOrder:       "oldest",
		AlbumTemplate:    "",
		AlbumFavorite:    false,
		Links:            []Link{},
		CreatedAt:        time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC),
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
