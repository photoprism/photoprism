package entity

import (
	"time"
)

var LabelFixtures = map[string]Label{
	"landscape": {
		ID:               1000000,
		LabelUUID:        "lt9k3pw1wowuy3c2",
		LabelSlug:        "landscape",
		CustomSlug:       "landscape",
		LabelName:        "Landscape",
		LabelPriority:    0,
		LabelFavorite:    true,
		LabelDescription: "",
		LabelNotes:       "",
		LabelCategories:  []*Label{},
		Links:            []Link{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        nil,
		New:              false,
	},
	"flower": {
		ID:               1000001,
		LabelUUID:        "lt9k3pw1wowuy3c3",
		LabelSlug:        "flower",
		CustomSlug:       "flower",
		LabelName:        "Flower",
		LabelPriority:    1,
		LabelFavorite:    true,
		LabelDescription: "",
		LabelNotes:       "",
		LabelCategories:  []*Label{},
		Links:            []Link{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        nil,
		New:              false,
	},
	"cake": {
		ID:               1000002,
		LabelUUID:        "lt9k3pw1wowuy3c4",
		LabelSlug:        "cake",
		CustomSlug:       "kuchen",
		LabelName:        "Cake",
		LabelPriority:    5,
		LabelFavorite:    false,
		LabelDescription: "",
		LabelNotes:       "",
		LabelCategories:  []*Label{},
		Links:            []Link{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        nil,
		New:              false,
	},
	"cow": {
		ID:               1000003,
		LabelUUID:        "lt9k3pw1wowuy3c5",
		LabelSlug:        "cow",
		CustomSlug:       "kuh",
		LabelName:        "COW",
		LabelPriority:    -1,
		LabelFavorite:    true,
		LabelDescription: "",
		LabelNotes:       "",
		LabelCategories:  []*Label{},
		Links:            []Link{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        nil,
		New:              false,
	},
}

var LabelFixtureLandscape = LabelFixtures["landscape"]
var LabelFixtureFlower = LabelFixtures["flower"]
var LabelFixtureCake = LabelFixtures["cake"]

// CreateLabelFixtures inserts known entities into the database for testing.
func CreateLabelFixtures() {
	for _, entity := range LabelFixtures {
		Db().Create(&entity)
	}
}
