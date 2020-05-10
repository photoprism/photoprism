package entity

import "time"

type PlacesMap map[string]Place

func (m PlacesMap) Get(name string) Place {
	if result, ok := m[name]; ok {
		return result
	}

	return UnknownPlace
}

func (m PlacesMap) Pointer(name string) *Place {
	if result, ok := m[name]; ok {
		return &result
	}

	return &UnknownPlace
}

var PlaceFixtures = PlacesMap{
	"teotihuacan": {
		ID:          "85d1ea7d382c",
		LocLabel:    "Teotihuacán, Mexico, Mexico",
		LocCity:     "Teotihuacán",
		LocState:    "Mexico",
		LocCountry:  "mx",
		LocKeywords: "ancient, pyramid",
		LocNotes:    "",
		LocFavorite: false,
		PhotoCount:  1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"zinkwazi": {
		ID:          "1ef744d1e28c",
		LocLabel:    "KwaDukuza, KwaZulu-Natal, South Africa",
		LocCity:     "KwaDukuza",
		LocState:    "KwaZulu-Natal",
		LocCountry:  "za",
		LocKeywords: "",
		LocNotes:    "africa",
		LocFavorite: true,
		PhotoCount:  2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
}

// CreatePlaceFixtures inserts known entities into the database for testing.
func CreatePlaceFixtures() {
	for _, entity := range PlaceFixtures {
		Db().Create(&entity)
	}
}
