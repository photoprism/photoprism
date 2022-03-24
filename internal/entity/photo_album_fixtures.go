package entity

import "time"

type PhotoAlbumMap map[string]PhotoAlbum

func (m PhotoAlbumMap) Get(name, photoUID, albumUID string) PhotoAlbum {
	if result, ok := m[name]; ok {
		return result
	}

	return *NewPhotoAlbum(photoUID, albumUID)
}

func (m PhotoAlbumMap) Pointer(name, photoUID, albumUID string) *PhotoAlbum {
	if result, ok := m[name]; ok {
		return &result
	}

	return NewPhotoAlbum(photoUID, albumUID)
}

var PhotoAlbumFixtures = PhotoAlbumMap{
	"1": {
		PhotoUID:  "pt9jtdre2lvl0yh7",
		AlbumUID:  "at9lxuqxpogaaba8",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 3, 28, 14, 6, 0, 0, time.UTC),
		Photo:     PhotoFixtures.Pointer("19800101_000002_D640C559"),
		Album:     AlbumFixtures.Pointer("holiday-2030"),
	},
	"2": {
		PhotoUID:  "pt9jtdre2lvl0y11",
		AlbumUID:  "at9lxuqxpogaaba9",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
		Photo:     PhotoFixtures.Pointer("Photo04"),
		Album:     AlbumFixtures.Pointer("berlin-2019"),
	},
	"3": {
		PhotoUID:  "pt9jtdre2lvl0yh8",
		AlbumUID:  "at9lxuqxpogaaba9",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
		Photo:     PhotoFixtures.Pointer("Photo01"),
		Album:     AlbumFixtures.Pointer("berlin-2019"),
	},
	"4": {
		PhotoUID:  "pt9jtxrexxvl0yh0",
		AlbumUID:  "at9lxuqxpogaaba9",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
		Photo:     PhotoFixtures.Pointer("Photo19"),
		Album:     AlbumFixtures.Pointer("april-1990"),
	},
	"5": {
		PhotoUID:  "pt9jtdre2lvl0yh0",
		AlbumUID:  "at9lxuqxpogaaba9",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
		Photo:     PhotoFixtures.Pointer("Photo03"),
		Album:     AlbumFixtures.Pointer("berlin-2019"),
	},
	"6": {
		PhotoUID:  "pt9jtdre2lvl0yh0",
		AlbumUID:  "at9lxuqxpogaaba8",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
		Photo:     PhotoFixtures.Pointer("Photo03"),
		Album:     AlbumFixtures.Pointer("berlin-2019"),
	},
	"7": {
		PhotoUID:  "pt9jtdre2lvl0y21",
		AlbumUID:  "at9lxuqxpogaaba7",
		Hidden:    false,
		Missing:   false,
		Order:     1,
		CreatedAt: time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 5, 28, 14, 6, 0, 0, time.UTC),
		Photo:     PhotoFixtures.Pointer("Photo14"),
		Album:     AlbumFixtures.Pointer("berlin-2019"),
	},
	"8": {
		PhotoUID:  "pt9jtdre2lvl0y21",
		AlbumUID:  "at9lxuqxpogaaba8",
		Hidden:    false,
		Missing:   false,
		Order:     1,
		CreatedAt: time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 5, 28, 14, 6, 0, 0, time.UTC),
		Photo:     PhotoFixtures.Pointer("Photo14"),
		Album:     AlbumFixtures.Pointer("berlin-2019"),
	},
}

// CreatePhotoAlbumFixtures inserts known entities into the database for testing.
func CreatePhotoAlbumFixtures() {
	for _, entity := range PhotoAlbumFixtures {
		Db().Create(&entity)
	}
}
