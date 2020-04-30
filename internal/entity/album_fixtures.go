package entity

import (
	"github.com/jinzhu/gorm"
	"time"
)

var AlbumFixtures = map[string]Album{
	"christmas2030": {
		ID:               1000000,
		CoverUUID:        nil,
		AlbumUUID:        "aq9lxuqxpogaaba7",
		AlbumSlug:        "christmas2030",
		AlbumName:        "Christmas2030",
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
		CoverUUID:        nil,
		AlbumUUID:        "aq9lxuqxpogaaba8",
		AlbumSlug:        "holiday-2030",
		AlbumName:        "Holiday2030",
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
		CoverUUID:        nil,
		AlbumUUID:        "aq9lxuqxpogaaba9",
		AlbumSlug:        "berlin-2019",
		AlbumName:        "Berlin2019",
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
func CreateAlbumFixtures(db *gorm.DB) {
	for _, entity := range AlbumFixtures {
		db.Create(&entity)
	}
}
