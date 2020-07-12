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
		ID:          s2.TokenPrefix + "85d1ea7d3278",
		GeoLabel:    "Teotihuacán, Mexico, Mexico",
		GeoCity:     "Teotihuacán",
		GeoState:    "State of Mexico",
		GeoCountry:  "mx",
		GeoKeywords: "ancient, pyramid",
		GeoFavorite: false,
		PhotoCount:  1,
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"zinkwazi": {
		ID:          s2.TokenPrefix + "1ef744d1e279",
		GeoLabel:    "KwaDukuza, KwaZulu-Natal, South Africa",
		GeoCity:     "KwaDukuza",
		GeoState:    "KwaZulu-Natal",
		GeoCountry:  "za",
		GeoKeywords: "",
		GeoFavorite: true,
		PhotoCount:  2,
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"holidaypark": {
		ID:          s2.TokenPrefix + "1ef744d1e280",
		GeoLabel:    "Holiday Park, Amusement",
		GeoCity:     "",
		GeoState:    "Rheinland-Pfalz",
		GeoCountry:  "de",
		GeoKeywords: "",
		GeoFavorite: true,
		PhotoCount:  2,
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"emptyNameLongCity": {
		ID:          s2.TokenPrefix + "1ef744d1e281",
		GeoLabel:    "labelEmptyNameLongCity",
		GeoCity:     "longlonglonglonglongcity",
		GeoState:    "Rheinland-Pfalz",
		GeoCountry:  "de",
		GeoKeywords: "",
		GeoFavorite: true,
		PhotoCount:  2,
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"emptyNameShortCity": {
		ID:          s2.TokenPrefix + "1ef744d1e282",
		GeoLabel:    "labelEmptyNameShortCity",
		GeoCity:     "shortcity",
		GeoState:    "Rheinland-Pfalz",
		GeoCountry:  "de",
		GeoKeywords: "",
		GeoFavorite: true,
		PhotoCount:  2,
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"veryLongLocName": {
		ID:          s2.TokenPrefix + "1ef744d1e283",
		GeoLabel:    "labelVeryLongLocName",
		GeoCity:     "Mainz",
		GeoState:    "Rheinland-Pfalz",
		GeoCountry:  "de",
		GeoKeywords: "",
		GeoFavorite: true,
		PhotoCount:  2,
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"mediumLongLocName": {
		ID:          s2.TokenPrefix + "1ef744d1e284",
		GeoLabel:    "labelMediumLongLocName",
		GeoCity:     "New york",
		GeoState:    "New york",
		GeoCountry:  "us",
		GeoKeywords: "",
		GeoFavorite: true,
		PhotoCount:  2,
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
}

// CreatePlaceFixtures inserts known entities into the database for testing.
func CreatePlaceFixtures() {
	for _, entity := range PlaceFixtures {
		Db().Create(&entity)
	}
}
