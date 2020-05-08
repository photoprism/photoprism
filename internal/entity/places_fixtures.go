package entity

import "time"

var PlaceFixtures = map[string]Place{
	"teotihuacan": {
		ID:          "1000000",
		LocLabel:    "Teotihuacán, Mexico, Mexico",
		LocCity:     "Teotihuacán",
		LocState:    "Mexico",
		LocCountry:  "mx",
		LocKeywords: "ancient, pyramid",
		LocNotes:    "",
		LocFavorite: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"zinkwazi": {
		ID:          "1000001",
		LocLabel:    "KwaDukuza, KwaZulu-Natal, South Africa",
		LocCity:     "KwaDukuza",
		LocState:    "KwaZulu-Natal",
		LocCountry:  "za",
		LocKeywords: "",
		LocNotes:    "africa",
		LocFavorite: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
}

var PlaceFixtureTeotihuacan = PlaceFixtures["teotihuacan"]
var PlaceFixtureZinkwazi = PlaceFixtures["zinkwazi"]

// CreatePlaceFixtures inserts known entities into the database for testing.
func CreatePlaceFixtures() {
	for _, entity := range PlaceFixtures {
		Db().Create(&entity)
	}
}
