package entity

import (
	"github.com/photoprism/photoprism/pkg/s2"
)

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
	"mexico": {
		ID:            "mx:VvfNBpFegSCr",
		PlaceLabel:    "Teotihuacán, State of Mexico, Mexico",
		PlaceCity:     "Teotihuacán",
		PlaceState:    "State of Mexico",
		PlaceCountry:  "mx",
		PlaceKeywords: "ancient, pyramid",
		PlaceFavorite: false,
		PhotoCount:    1,
		CreatedAt:     TimeStamp(),
		UpdatedAt:     TimeStamp(),
	},
	"zinkwazi": {
		ID:            "za:Rc1K7dTWRzBD",
		PlaceLabel:    "KwaDukuza, KwaZulu-Natal, South Africa",
		PlaceCity:     "KwaDukuza",
		PlaceState:    "KwaZulu-Natal",
		PlaceCountry:  "za",
		PlaceKeywords: "",
		PlaceFavorite: true,
		PhotoCount:    2,
		CreatedAt:     TimeStamp(),
		UpdatedAt:     TimeStamp(),
	},
	"holidaypark": {
		ID:            "de:HFqPHxa2Hsol",
		PlaceLabel:    "Neustadt an der Weinstraße, Rheinland-Pfalz, Germany",
		PlaceDistrict: "Hambach an der Weinstraße",
		PlaceCity:     "Neustadt an der Weinstraße",
		PlaceState:    "Rheinland-Pfalz",
		PlaceCountry:  "de",
		PlaceKeywords: "",
		PlaceFavorite: true,
		PhotoCount:    2,
		CreatedAt:     TimeStamp(),
		UpdatedAt:     TimeStamp(),
	},
	"emptyNameLongCity": {
		ID:            s2.TokenPrefix + "1ef744d1e281",
		PlaceLabel:    "labelEmptyNameLongCity",
		PlaceCity:     "longlonglonglonglongcity",
		PlaceState:    "Rheinland-Pfalz",
		PlaceCountry:  "de",
		PlaceKeywords: "",
		PlaceFavorite: true,
		PhotoCount:    2,
		CreatedAt:     TimeStamp(),
		UpdatedAt:     TimeStamp(),
	},
	"emptyNameShortCity": {
		ID:            s2.TokenPrefix + "1ef744d1e282",
		PlaceLabel:    "labelEmptyNameShortCity",
		PlaceCity:     "shortcity",
		PlaceState:    "Rheinland-Pfalz",
		PlaceCountry:  "de",
		PlaceKeywords: "",
		PlaceFavorite: true,
		PhotoCount:    2,
		CreatedAt:     TimeStamp(),
		UpdatedAt:     TimeStamp(),
	},
	"veryLongLocName": {
		ID:            s2.TokenPrefix + "1ef744d1e283",
		PlaceLabel:    "labelVeryLongLocName",
		PlaceCity:     "Mainz",
		PlaceState:    "Rheinland-Pfalz",
		PlaceCountry:  "de",
		PlaceKeywords: "",
		PlaceFavorite: true,
		PhotoCount:    2,
		CreatedAt:     TimeStamp(),
		UpdatedAt:     TimeStamp(),
	},
	"mediumLongLocName": {
		ID:            s2.TokenPrefix + "1ef744d1e284",
		PlaceLabel:    "labelMediumLongLocName",
		PlaceCity:     "New york",
		PlaceState:    "New york",
		PlaceCountry:  "us",
		PlaceKeywords: "",
		PlaceFavorite: true,
		PhotoCount:    2,
		CreatedAt:     TimeStamp(),
		UpdatedAt:     TimeStamp(),
	},
	"Germany": {
		ID:            s2.TokenPrefix + "1ef744d1e285",
		PlaceLabel:    "Germany",
		PlaceCity:     "",
		PlaceState:    "",
		PlaceCountry:  "de",
		PlaceKeywords: "",
		PlaceFavorite: false,
		PhotoCount:    1,
		CreatedAt:     TimeStamp(),
		UpdatedAt:     TimeStamp(),
	},
}

// CreatePlaceFixtures inserts known entities into the database for testing.
func CreatePlaceFixtures() {
	for _, entity := range PlaceFixtures {
		Db().Create(&entity)
	}
}
