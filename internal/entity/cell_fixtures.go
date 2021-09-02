package entity

import (
	"github.com/photoprism/photoprism/pkg/s2"
)

type CellMap map[string]Cell

func (m CellMap) Get(name string) Cell {
	if result, ok := m[name]; ok {
		return result
	}

	return UnknownLocation
}

func (m CellMap) Pointer(name string) *Cell {
	if result, ok := m[name]; ok {
		return &result
	}

	return &UnknownLocation
}

var CellFixtures = CellMap{
	"mexico": {
		ID:           s2.TokenPrefix + "85d1ea7d382c",
		PlaceID:      PlaceFixtures.Get("mexico").ID,
		CellName:     "Adosada Platform",
		CellCategory: "botanical garden",
		Place:        PlaceFixtures.Pointer("mexico"),
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
	},
	"caravan park": {
		ID:      s2.TokenPrefix + "1ef75a71a36c",
		PlaceID: s2.TokenPrefix + "1ef75a71a36c",
		Place: &Place{
			ID:           s2.TokenPrefix + "1ef75a71a36",
			PlaceLabel:   "Mandeni, KwaZulu-Natal, South Africa",
			PlaceCity:    "Mandeni",
			PlaceState:   "KwaZulu-Natal",
			PlaceCountry: "za",
			CreatedAt:    TimeStamp(),
			UpdatedAt:    TimeStamp(),
		},
		CellName:     "Lobotes Caravan Park",
		CellCategory: "camping",
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
	},
	"zinkwazi": {
		ID:           s2.TokenPrefix + "1ef744d1e28c",
		PlaceID:      PlaceFixtures.Get("zinkwazi").ID,
		Place:        PlaceFixtures.Pointer("zinkwazi"),
		CellName:     "Zinkwazi Beach",
		CellCategory: "beach",
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
	},
	"hassloch": {
		ID:           s2.TokenPrefix + "1ef744d1e280",
		PlaceID:      PlaceFixtures.Get("holidaypark").ID,
		Place:        PlaceFixtures.Pointer("holidaypark"),
		CellName:     "Holiday Park",
		CellCategory: "park",
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
	},
	"emptyNameLongCity": {
		ID:           s2.TokenPrefix + "1ef744d1e281",
		PlaceID:      PlaceFixtures.Get("emptyNameLongCity").ID,
		Place:        PlaceFixtures.Pointer("emptyNameLongCity"),
		CellName:     "",
		CellCategory: "botanical garden",
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
	},
	"emptyNameShortCity": {
		ID:           s2.TokenPrefix + "1ef744d1e282",
		PlaceID:      PlaceFixtures.Get("emptyNameShortCity").ID,
		Place:        PlaceFixtures.Pointer("emptyNameShortCity"),
		CellName:     "",
		CellCategory: "botanical garden",
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
	},
	"veryLongLocName": {
		ID:           s2.TokenPrefix + "1ef744d1e283",
		PlaceID:      PlaceFixtures.Get("veryLongLocName").ID,
		Place:        PlaceFixtures.Pointer("veryLongLocName"),
		CellName:     "longlonglonglonglonglonglonglonglonglonglonglonglongName",
		CellCategory: "cape",
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
	},
	"mediumLongLocName": {
		ID:           s2.TokenPrefix + "1ef744d1e283",
		PlaceID:      PlaceFixtures.Get("mediumLongLocName").ID,
		Place:        PlaceFixtures.Pointer("mediumLongLocName"),
		CellName:     "longlonglonglonglonglongName",
		CellCategory: "botanical garden",
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
	},
	"Neckarbrücke": {
		ID:           s2.TokenPrefix + "1ef744d1e284",
		PlaceID:      PlaceFixtures.Get("Germany").ID,
		Place:        PlaceFixtures.Pointer("Germany"),
		CellName:     "Neckarbrücke",
		CellCategory: "",
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
	},
}

// CreateCellFixtures inserts known entities into the database for testing.
func CreateCellFixtures() {
	for _, entity := range CellFixtures {
		Db().Create(&entity)
	}
}
