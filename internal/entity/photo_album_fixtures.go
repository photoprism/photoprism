package entity

import "time"

var PhotoAlbumFixtures = map[string]PhotoAlbum{
	"1": {
		PhotoUUID: "pt9jtdre2lvl0yh7",
		AlbumUUID: "at9lxuqxpogaaba8",
		Order:     0,
		CreatedAt: time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 3, 28, 14, 6, 0, 0, time.UTC),
		Photo:     PhotoFixtures.Pointer("19800101_000002_D640C559"),
		Album:     &AlbumFixtureHoliday2030,
	},
	"2": {
		PhotoUUID: "pt9jtdre2lvl0y11",
		AlbumUUID: "at9lxuqxpogaaba9",
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
		Photo:     PhotoFixtures.Pointer("Photo04"),
		Album:     &AlbumFixtureBerlin2019,
	},
}

var PhotoAlbumFixture1 = PhotoAlbumFixtures["1"]
var PhotoAlbumFixture2 = PhotoAlbumFixtures["2"]

// CreatePhotoAlbumFixtures inserts known entities into the database for testing.
func CreatePhotoAlbumFixtures() {
	for _, entity := range PhotoAlbumFixtures {
		Db().Create(&entity)
	}
}
