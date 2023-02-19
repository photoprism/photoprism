package entity

import "time"

var date = time.Date(2050, 3, 6, 2, 6, 51, 0, time.UTC)

type LinkMap map[string]Link

var LinkFixtures = LinkMap{
	"1jxf3jfn2k": {
		LinkUID:     "sqn2xpryd1ob7gtf",
		ShareUID:    "at9lxuqxpogaaba8",
		ShareSlug:   "holiday-2030",
		LinkToken:   "1jxf3jfn2k",
		LinkExpires: 0,
		LinkViews:   12,
		MaxViews:    0,
		HasPassword: false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		ModifiedAt:  time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"4jxf3jfn2k": {
		LinkUID:     "sqn2xpryd1ob8gtf",
		ShareUID:    "at9lxuqxpogaaba7",
		ShareSlug:   "christmas-2030",
		LinkToken:   "4jxf3jfn2k",
		LinkExpires: 0,
		LinkViews:   0,
		MaxViews:    0,
		HasPassword: false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		ModifiedAt:  time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"5jxf3jfn2k": {
		LinkUID:     "sqn9xpryd1ob9gtf",
		ShareUID:    "ft2es39w45bnlqdw",
		ShareSlug:   "ft2es39w45bnlqdw",
		LinkToken:   "5jxf3jfn2k",
		LinkExpires: 0,
		LinkViews:   0,
		MaxViews:    0,
		HasPassword: false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		ModifiedAt:  time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"6jxf3jfn2k": {
		LinkUID:     "sqn1xpryd1ob1gtf",
		ShareUID:    "lt9k3pw1wowuy3c3",
		ShareSlug:   "lt9k3pw1wowuy3c3",
		LinkToken:   "6jxf3jfn2k",
		LinkExpires: 0,
		LinkViews:   0,
		MaxViews:    0,
		HasPassword: false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		ModifiedAt:  time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"7jxf3jfn2k": {
		LinkUID:     "sqn2xpryd1ob2gtf",
		ShareUID:    "pt9k3pw1wowuy3c3",
		ShareSlug:   "pt9k3pw1wowuy3c3",
		LinkToken:   "7jxf3jfn2k",
		LinkExpires: 0,
		LinkViews:   0,
		MaxViews:    0,
		HasPassword: false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		ModifiedAt:  time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
}

// CreateLinkFixtures inserts known entities into the database for testing.
func CreateLinkFixtures() {
	for _, entity := range LinkFixtures {
		Db().Create(&entity)
	}
}
