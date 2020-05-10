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
}

// CreateLocationFixtures inserts known entities into the database for testing.
func CreateLocationFixtures() {
	for _, entity := range LocationFixtures {
		Db().Create(&entity)
	}
}
