package entity

import (
	"github.com/photoprism/photoprism/pkg/s2"
)

type GeoMap map[string]Geo

func (m GeoMap) Get(name string) Geo {
	if result, ok := m[name]; ok {
		return result
	}

	return UnknownLocation
}

func (m GeoMap) Pointer(name string) *Geo {
	if result, ok := m[name]; ok {
		return &result
	}

	return &UnknownLocation
}

var GeoFixtures = GeoMap{
	"mexico": {
		ID:          s2.TokenPrefix + "85d1ea7d382c",
		PlaceID:     PlaceFixtures.Get("mexico").ID,
		GeoName:     "Adosada Platform",
		GeoCategory: "botanical garden",
		Place:       PlaceFixtures.Pointer("mexico"),
		GeoSource:   "places",
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"caravan park": {
		ID:      s2.TokenPrefix + "1ef75a71a36c",
		PlaceID: s2.TokenPrefix + "1ef75a71a36c",
		Place: &Place{
			ID:         s2.TokenPrefix + "1ef75a71a36",
			GeoLabel:   "Mandeni, KwaZulu-Natal, South Africa",
			GeoCity:    "Mandeni",
			GeoState:   "KwaZulu-Natal",
			GeoCountry: "za",
			CreatedAt:  Timestamp(),
			UpdatedAt:  Timestamp(),
		},
		GeoName:     "Lobotes Caravan Park",
		GeoCategory: "camping",
		GeoSource:   "manual",
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"zinkwazi": {
		ID:          s2.TokenPrefix + "1ef744d1e28c",
		PlaceID:     PlaceFixtures.Get("zinkwazi").ID,
		Place:       PlaceFixtures.Pointer("zinkwazi"),
		GeoName:     "Zinkwazi Beach",
		GeoCategory: "beach",
		GeoSource:   "places",
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"hassloch": {
		ID:          s2.TokenPrefix + "1ef744d1e280",
		PlaceID:     PlaceFixtures.Get("holidaypark").ID,
		Place:       PlaceFixtures.Pointer("holidaypark"),
		GeoName:     "Holiday Park",
		GeoCategory: "park",
		GeoSource:   "places",
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"emptyNameLongCity": {
		ID:          s2.TokenPrefix + "1ef744d1e281",
		PlaceID:     PlaceFixtures.Get("emptyNameLongCity").ID,
		Place:       PlaceFixtures.Pointer("emptyNameLongCity"),
		GeoName:     "",
		GeoCategory: "botanical garden",
		GeoSource:   "places",
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"emptyNameShortCity": {
		ID:          s2.TokenPrefix + "1ef744d1e282",
		PlaceID:     PlaceFixtures.Get("emptyNameShortCity").ID,
		Place:       PlaceFixtures.Pointer("emptyNameShortCity"),
		GeoName:     "",
		GeoCategory: "botanical garden",
		GeoSource:   "places",
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"veryLongLocName": {
		ID:          s2.TokenPrefix + "1ef744d1e283",
		PlaceID:     PlaceFixtures.Get("veryLongLocName").ID,
		Place:       PlaceFixtures.Pointer("veryLongLocName"),
		GeoName:     "longlonglonglonglonglonglonglonglonglonglonglonglongName",
		GeoCategory: "cape",
		GeoSource:   "places",
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
	"mediumLongLocName": {
		ID:          s2.TokenPrefix + "1ef744d1e283",
		PlaceID:     PlaceFixtures.Get("mediumLongLocName").ID,
		Place:       PlaceFixtures.Pointer("mediumLongLocName"),
		GeoName:     "longlonglonglonglonglongName",
		GeoCategory: "botanical garden",
		GeoSource:   "places",
		CreatedAt:   Timestamp(),
		UpdatedAt:   Timestamp(),
	},
}

// CreateGeoFixtures inserts known entities into the database for testing.
func CreateGeoFixtures() {
	for _, entity := range GeoFixtures {
		Db().Create(&entity)
	}
}
