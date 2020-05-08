package entity

import "time"

var LocationFixtures = map[string]Location{
	"mexico": {
		ID:          "1000000",
		PlaceID:     "1000000",
		LocName:     "Adosada Platform",
		LocCategory: "tourism",
		Place:       &PlaceFixtureTeotihuacan,
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"caravan park": {
		ID:          "1000001",
		PlaceID:     "",
		Place:       nil,
		LocName:     "Lobotes Caravan Park",
		LocCategory: "camping",
		LocSource:   "manual",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"zinkwazi": {
		ID:          "1000002",
		PlaceID:     "",
		Place:       &PlaceFixtureZinkwazi,
		LocName:     "Zinkwazi Beach",
		LocCategory: "",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
}

var LocationFixtureMexico = LocationFixtures["mexico"]
var LocationFixtureCaravanPark = LocationFixtures["caravan park"]
var LocationFixtureZinkawzi = LocationFixtures["zinkwazi"]

// CreateLocationFixtures inserts known entities into the database for testing.
func CreateLocationFixtures() {
	for _, entity := range LocationFixtures {
		Db().Create(&entity)
	}
}
