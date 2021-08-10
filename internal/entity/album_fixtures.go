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
		FolderUID:        "",
		AlbumSlug:        "christmas-2030",
		AlbumPath:        "",
		AlbumType:        AlbumDefault,
		AlbumTitle:       "Christmas 2030",
		AlbumLocation:    "",
		AlbumCategory:    "",
		AlbumCaption:     "",
		AlbumDescription: "Wonderful Christmas",
		AlbumNotes:       "",
		AlbumFilter:      "",
		AlbumOrder:       "oldest",
		AlbumTemplate:    "",
		AlbumCountry:     "zz",
		AlbumYear:        0,
		AlbumMonth:       0,
		AlbumDay:         0,
		AlbumFavorite:    false,
		AlbumPrivate:     false,
		CreatedAt:        time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"holiday-2030": {
		ID:               1000001,
		CoverUID:         "",
		AlbumUID:         "at9lxuqxpogaaba8",
		FolderUID:        "",
		AlbumSlug:        "holiday-2030",
		AlbumPath:        "",
		AlbumType:        AlbumDefault,
		AlbumTitle:       "Holiday 2030",
		AlbumLocation:    "",
		AlbumCategory:    "",
		AlbumCaption:     "",
		AlbumDescription: "Wonderful Christmas Holiday",
		AlbumNotes:       "",
		AlbumFilter:      "",
		AlbumOrder:       "newest",
		AlbumTemplate:    "",
		AlbumCountry:     "zz",
		AlbumYear:        0,
		AlbumMonth:       0,
		AlbumDay:         0,
		AlbumFavorite:    true,
		AlbumPrivate:     false,
		CreatedAt:        time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"berlin-2019": {
		ID:               1000002,
		CoverUID:         "",
		AlbumUID:         "at9lxuqxpogaaba9",
		FolderUID:        "",
		AlbumSlug:        "berlin-2019",
		AlbumPath:        "",
		AlbumType:        AlbumDefault,
		AlbumTitle:       "Berlin 2019",
		AlbumLocation:    "Berlin",
		AlbumCategory:    "City",
		AlbumCaption:     "",
		AlbumDescription: "We love Berlin 🌿!",
		AlbumNotes:       "",
		AlbumFilter:      "",
		AlbumOrder:       "oldest",
		AlbumTemplate:    "",
		AlbumCountry:     "",
		AlbumYear:        0,
		AlbumMonth:       0,
		AlbumDay:         0,
		AlbumFavorite:    false,
		AlbumPrivate:     false,
		CreatedAt:        time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"april-1990": {
		ID:               1000003,
		CoverUID:         "",
		AlbumUID:         "at1lxuqipogaaba1",
		FolderUID:        "",
		AlbumSlug:        "april-1990",
		AlbumPath:        "1990/04",
		AlbumType:        AlbumFolder,
		AlbumTitle:       "April 1990",
		AlbumLocation:    "",
		AlbumCategory:    "Friends",
		AlbumCaption:     "",
		AlbumDescription: "Spring is the time of year when many things change.",
		AlbumNotes:       "",
		AlbumFilter:      "path:\"1990/04\" public:true",
		AlbumOrder:       "added",
		AlbumTemplate:    "",
		AlbumCountry:     "ca",
		AlbumYear:        1990,
		AlbumMonth:       4,
		AlbumDay:         11,
		AlbumFavorite:    false,
		AlbumPrivate:     false,
		CreatedAt:        time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"import": {
		ID:               1000004,
		CoverUID:         "",
		AlbumUID:         "at6axuzitogaaiax",
		FolderUID:        "",
		AlbumSlug:        "import",
		AlbumPath:        "",
		AlbumType:        AlbumDefault,
		AlbumTitle:       "Import Album",
		AlbumLocation:    "",
		AlbumCategory:    "",
		AlbumCaption:     "",
		AlbumDescription: "",
		AlbumNotes:       "",
		AlbumFilter:      "",
		AlbumOrder:       "name",
		AlbumTemplate:    "",
		AlbumCountry:     "ca",
		AlbumYear:        0,
		AlbumMonth:       0,
		AlbumDay:         0,
		AlbumFavorite:    false,
		AlbumPrivate:     false,
		CreatedAt:        time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"emptyMoment": {
		ID:               1000005,
		CoverUID:         "",
		AlbumUID:         "at7axuzitogaaiax",
		FolderUID:        "",
		AlbumSlug:        "empty-moment",
		AlbumPath:        "",
		AlbumType:        AlbumMoment,
		AlbumTitle:       "Empty Moment",
		AlbumLocation:    "Favorite Park",
		AlbumCategory:    "Fun",
		AlbumCaption:     "",
		AlbumDescription: "",
		AlbumNotes:       "",
		AlbumFilter:      "public:true country:at year:2016",
		AlbumOrder:       "oldest",
		AlbumTemplate:    "",
		AlbumCountry:     "at",
		AlbumYear:        2016,
		AlbumMonth:       0,
		AlbumDay:         0,
		AlbumFavorite:    false,
		AlbumPrivate:     false,
		CreatedAt:        time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		DeletedAt:        nil,
	},
	"2016-04": {
		ID:               1000006,
		CoverUID:         "",
		AlbumUID:         "at1lxuqipogaabj8",
		FolderUID:        "",
		AlbumSlug:        "2016-04",
		AlbumPath:        "2016/04",
		AlbumType:        AlbumFolder,
		AlbumTitle:       "April 2016",
		AlbumLocation:    "",
		AlbumCategory:    "Fun",
		AlbumCaption:     "",
		AlbumDescription: "",
		AlbumNotes:       "",
		AlbumFilter:      "path:\"2016/04\" public:true",
		AlbumOrder:       "added",
		AlbumTemplate:    "",
		AlbumCountry:     "zz",
		AlbumYear:        0,
		AlbumMonth:       0,
		AlbumDay:         0,
		AlbumFavorite:    false,
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
