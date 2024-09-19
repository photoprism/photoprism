package entity

import "time"

type PhotoAlbumMap map[string]PhotoAlbum

func (m PhotoAlbumMap) Get(name, photoUid, albumUid string) PhotoAlbum {
	if result, ok := m[name]; ok {
		return result
	}

	return *NewPhotoAlbum(photoUid, albumUid)
}

func (m PhotoAlbumMap) Pointer(name, photoUid, albumUid string) *PhotoAlbum {
	if result, ok := m[name]; ok {
		return &result
	}

	return NewPhotoAlbum(photoUid, albumUid)
}

var PhotoAlbumFixtures = PhotoAlbumMap{
	"1": {
		PhotoUID:  "ps6sg6be2lvl0yh7",
		AlbumUID:  "as6sg6bxpogaaba8",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 3, 28, 14, 6, 0, 0, time.UTC),
	},
	"2": {
		PhotoUID:  "ps6sg6be2lvl0y11",
		AlbumUID:  "as6sg6bxpogaaba9",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"3": {
		PhotoUID:  "ps6sg6be2lvl0yh8",
		AlbumUID:  "as6sg6bxpogaaba9",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"4": {
		PhotoUID:  "ps6sg6bexxvl0yh0",
		AlbumUID:  "as6sg6bxpogaaba9",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"5": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bxpogaaba9",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"6": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bxpogaaba8",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"7": {
		PhotoUID:  "ps6sg6be2lvl0y21",
		AlbumUID:  "as6sg6bxpogaaba7",
		Hidden:    false,
		Missing:   false,
		Order:     1,
		CreatedAt: time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 5, 28, 14, 6, 0, 0, time.UTC),
	},
	"8": {
		PhotoUID:  "ps6sg6be2lvl0y21",
		AlbumUID:  "as6sg6bxpogaaba8",
		Hidden:    false,
		Missing:   false,
		Order:     1,
		CreatedAt: time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 5, 28, 14, 6, 0, 0, time.UTC),
	},
	"9": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab24",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"10": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab23",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"11": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab19",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"12": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab22",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"13": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab21",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"14": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab20",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"15": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab25",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"16": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab26",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"17": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab27",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"18": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab28",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"19": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab29",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"20": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab30",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"21": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab31",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"22": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab32",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"23": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab33",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"24": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab34",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"25": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab35",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"26": {
		PhotoUID:  "ps6sg6be2lvl0yh0",
		AlbumUID:  "as6sg6bipotaab36",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"27": {
		PhotoUID:  "ps6sg6be2lvl0yh9",
		AlbumUID:  "as6sg6bipotaab26",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
	"28": {
		PhotoUID:  "ps6sg6be2lvl0yh9",
		AlbumUID:  "as6sg6bipotaab24",
		Hidden:    false,
		Missing:   false,
		Order:     0,
		CreatedAt: time.Date(2020, 2, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt: time.Date(2020, 4, 28, 14, 6, 0, 0, time.UTC),
	},
}

// CreatePhotoAlbumFixtures inserts known entities into the database for testing.
func CreatePhotoAlbumFixtures() {
	for _, entity := range PhotoAlbumFixtures {
		Db().Save(&entity)
	}
}
