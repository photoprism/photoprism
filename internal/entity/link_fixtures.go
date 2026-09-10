package entity

import "time"

// LinkMap represents a map of share link fixtures keyed by token.
type LinkMap map[string]Link

// Get retrieves the named link
func (m LinkMap) Get(name string) Link {
	if result, ok := m[name]; ok {
		return result
	}

	return Link{}
}

// Get retrieves the pointer to the named link
func (m LinkMap) Pointer(name string) *Link {
	if result, ok := m[name]; ok {
		return &result
	}

	return &Link{}
}

// LinkFixtures provides share link fixtures for use in tests.
//
//nolint:gosec // G101: Deterministic fixture tokens for tests only.
var LinkFixtures = LinkMap{
	"1jxf3jfn2k": {
		LinkUID:     "ss62xpryd1ob7gtf",
		ShareUID:    "as6sg6bxpogaaba8",
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
		LinkUID:     "ss62xpryd1ob8gtf",
		ShareUID:    "as6sg6bxpogaaba7",
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
		LinkUID:     "ss69xpryd1ob9gtf",
		ShareUID:    "fs6sg6bw45bn0004",
		ShareSlug:   "fs6sg6bw45bn0004",
		LinkToken:   "5jxf3jfn2k",
		LinkExpires: 0,
		LinkViews:   0,
		MaxViews:    0,
		HasPassword: false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		ModifiedAt:  time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"6jxf3jfn2k": {
		LinkUID:     "ss61xpryd1ob1gtf",
		ShareUID:    "ls6sg6b1wowuy3c3",
		ShareSlug:   "ls6sg6b1wowuy3c3",
		LinkToken:   "6jxf3jfn2k",
		LinkExpires: 0,
		LinkViews:   0,
		MaxViews:    0,
		HasPassword: false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		ModifiedAt:  time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"7jxf3jfn2k": {
		LinkUID:     "ss62xpryd1ob2gtf",
		ShareUID:    "ps6sg6b1wowuy3c3",
		ShareSlug:   "ps6sg6b1wowuy3c3",
		LinkToken:   "7jxf3jfn2k",
		LinkExpires: 0,
		LinkViews:   0,
		MaxViews:    0,
		HasPassword: false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		ModifiedAt:  time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"8jxf3jfn2k": {
		LinkUID:     "ss62xpryd1ob3gtf",
		ShareUID:    "as6sg6bipogaaba1", // "april-1990" folder album (smart, path filter).
		ShareSlug:   "april-1990",
		LinkToken:   "8jxf3jfn2k",
		LinkExpires: 0,
		LinkViews:   0,
		MaxViews:    0,
		HasPassword: false,
		CreatedAt:   time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		ModifiedAt:  time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
	},
	"9jxf3jfn2k": {
		LinkUID:     "ss62xpryd1ob4gtf",
		ShareUID:    "as6sg6bipogaab11", // "california-usa" state album (smart, place filter).
		ShareSlug:   "california-usa",
		LinkToken:   "9jxf3jfn2k",
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
		fixtureDb().Create(&entity)
	}
}
