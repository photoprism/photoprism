package entity

import "time"

var date = time.Date(2050, 3, 6, 2, 6, 51, 0, time.UTC)

type LinkMap map[string]Link

var LinkFixtures = LinkMap{
	"1jxf3jfn2k": {
		LinkUID:     "abcfgty",
		LinkToken:   "1jxf3jfn2k",
		LinkExpires: 0,
		ShareUID:    "st9lxuqxpogaaba7",
		CanComment:  true,
		CanEdit:     false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"4jxf3jfn2k": {
		LinkToken:   "4jxf3jfn2k",
		LinkExpires: 0,
		ShareUID:    "at9lxuqxpogaaba7",
		CanComment:  true,
		CanEdit:     false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"5jxf3jfn2k": {
		LinkToken:   "5jxf3jfn2k",
		LinkExpires: 0,
		ShareUID:    "ft2es39w45bnlqdw",
		ShareSlug:   "ft2es39w45bnlqdw",
		CanComment:  true,
		CanEdit:     false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"6jxf3jfn2k": {
		LinkToken:   "6jxf3jfn2k",
		LinkExpires: 0,
		ShareUID:    "st9lxuqxpogaaba7",
		ShareSlug:   "lt9k3pw1wowuy3c3",
		CanComment:  true,
		CanEdit:     false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"7jxf3jfn2k": {
		LinkToken:   "7jxf3jfn2k",
		LinkExpires: 0,
		ShareUID:    "st9lxuqxpogaaba7",
		ShareSlug:   "pt9k3pw1wowuy3c3",
		CanComment:  true,
		CanEdit:     false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
}

// CreateLinkFixtures inserts known entities into the database for testing.
func CreateLinkFixtures() {
	for _, entity := range LinkFixtures {
		Db().Create(&entity)
	}
}
