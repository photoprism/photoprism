package entity

import "time"

var date = time.Date(2050, 3, 6, 2, 6, 51, 0, time.UTC)

type LinkMap map[string]Link

var LinkFixtures = LinkMap{
	"1jxf3jfn2k": {
		ShareToken:   "1jxf3jfn2k",
		ShareExpires: 0,
		ShareUID:     "st9lxuqxpogaaba7",
		CanComment:   true,
		CanEdit:      false,
		CreatedAt:    time.Date(2020, 3, 6, 2, 6, 51, 0, time.UTC),
		UpdatedAt:    time.Date(2020, 3, 28, 14, 6, 0, 0, time.UTC),
	},
}

// CreateLinkFixtures inserts known entities into the database for testing.
func CreateLinkFixtures() {
	for _, entity := range LinkFixtures {
		Db().Create(&entity)
	}
}
