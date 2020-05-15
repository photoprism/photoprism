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
		ID:          "85d1ea7d3278",
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
		ID:          "1ef744d1e279",
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
	"holidaypark": {
		ID:          "1ef744d1e280",
		LocLabel:    "Holiday Park, Amusement",
		LocCity:     "",
		LocState:    "Rheinland-Pfalz",
		LocCountry:  "de",
		LocKeywords: "",
		LocNotes:    "germany",
		LocFavorite: true,
		PhotoCount:  2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"emptyNameLongCity": {
		ID:          "1ef744d1e281",
		LocLabel:    "label",
		LocCity:     "longlonglonglonglongcity",
		LocState:    "Rheinland-Pfalz",
		LocCountry:  "de",
		LocKeywords: "",
		LocNotes:    "germany",
		LocFavorite: true,
		PhotoCount:  2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"emptyNameShortCity": {
		ID:          "1ef744d1e282",
		LocLabel:    "label",
		LocCity:     "shortcity",
		LocState:    "Rheinland-Pfalz",
		LocCountry:  "de",
		LocKeywords: "",
		LocNotes:    "germany",
		LocFavorite: true,
		PhotoCount:  2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"veryLongLocName": {
		ID:          "1ef744d1e283",
		LocLabel:    "label",
		LocCity:     "Mainz",
		LocState:    "Rheinland-Pfalz",
		LocCountry:  "de",
		LocKeywords: "",
		LocNotes:    "germany",
		LocFavorite: true,
		PhotoCount:  2,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"mediumLongLocName": {
		ID:          "1ef744d1e284",
		LocLabel:    "label",
		LocCity:     "New york",
		LocState:    "New york",
		LocCountry:  "us",
		LocKeywords: "",
		LocNotes:    "",
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
