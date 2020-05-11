package entity

import "time"

type LocationMap map[string]Location

var LocationFixtures = LocationMap{
	"mexico": {
		ID:          "85d1ea7d382c",
		PlaceID:     PlaceFixtures.Get("teotihuacan").ID,
		LocName:     "Adosada Platform",
		LocCategory: "tourism",
		Place:       PlaceFixtures.Pointer("teotihuacan"),
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"caravan park": {
		ID:      "1ef75a71a36c",
		PlaceID: "1ef75a71a36c",
		Place: &Place{
			ID:         "1ef75a71a36",
			LocLabel:   "Mandeni, KwaZulu-Natal, South Africa",
			LocCity:    "Mandeni",
			LocState:   "KwaZulu-Natal",
			LocCountry: "za",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		LocName:     "Lobotes Caravan Park",
		LocCategory: "camping",
		LocSource:   "manual",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"zinkwazi": {
		ID:          "1ef744d1e28c",
		PlaceID:     PlaceFixtures.Get("zinkwazi").ID,
		Place:       PlaceFixtures.Pointer("zinkwazi"),
		LocName:     "Zinkwazi Beach",
		LocCategory: "",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"hassloch": {
		ID:          "1ef744d1e280",
		PlaceID:     PlaceFixtures.Get("holidaypark").ID,
		Place:       PlaceFixtures.Pointer("holidaypark"),
		LocName:     "Holiday Park",
		LocCategory: "",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"emptyNameLongCity": {
		ID:          "1ef744d1e281",
		PlaceID:     PlaceFixtures.Get("emptyNameLongCity").ID,
		Place:       PlaceFixtures.Pointer("emptyNameLongCity"),
		LocName:     "",
		LocCategory: "",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"emptyNameShortCity": {
		ID:          "1ef744d1e282",
		PlaceID:     PlaceFixtures.Get("emptyNameShortCity").ID,
		Place:       PlaceFixtures.Pointer("emptyNameShortCity"),
		LocName:     "",
		LocCategory: "",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"veryLongLocName": {
		ID:          "1ef744d1e283",
		PlaceID:     PlaceFixtures.Get("veryLongLocName").ID,
		Place:       PlaceFixtures.Pointer("veryLongLocName"),
		LocName:     "longlonglonglonglonglonglonglonglonglonglonglonglongName",
		LocCategory: "",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"mediumLongLocName": {
		ID:          "1ef744d1e283",
		PlaceID:     PlaceFixtures.Get("mediumLongLocName").ID,
		Place:       PlaceFixtures.Pointer("mediumLongLocName"),
		LocName:     "longlonglonglonglonglongName",
		LocCategory: "",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
}

var LocationFixturesMexico = LocationFixtures["mexico"]
var LocationFixturesHassloch = LocationFixtures["hassloch"]
var LocationFixturesEmptyNameLongCity = LocationFixtures["emptyNameLongCity"]
var LocationFixturesEmptyNameShortCity = LocationFixtures["emptyNameShortCity"]
var LocationFixturesVeryLongLocName = LocationFixtures["veryLongLocName"]
var LocationFixturesMediumLongLocName = LocationFixtures["mediumLongLocName"]

// CreateLocationFixtures inserts known entities into the database for testing.
func CreateLocationFixtures() {
	for _, entity := range LocationFixtures {
		Db().Create(&entity)
	}
}
