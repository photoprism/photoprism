package entity

import "time"

var PhotoAlbumFixtures = map[string]PhotoAlbum{
	"1": {
		PhotoUUID: "1jxf3jfn2k",
		AlbumUUID: "pq9jtdre2lvl0yh7",
		Order:     0,
		CreatedAt: time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 3, 28, 14, 6, 0, 0, time.UTC),
		Photo:     &PhotoFixture19800101_000002_D640C559,
		Album:     &AlbumFixtureHoliday2030,
	},
	"2": {
		PhotoUUID: "pq9jtdre2lvl0y11",
		AlbumUUID: "aq9lxuqxpogaaba9",
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
		Photo:     &PhotoFixturePhoto04,
		Album:     &AlbumFixtureBerlin2019,
	},
}

// CreatePhotoAlbumFixtures inserts known entities into the database for testing.
func CreatePhotoAlbumFixtures() {
	for _, entity := range PhotoAlbumFixtures {
		Db().Create(&entity)
	}
}
