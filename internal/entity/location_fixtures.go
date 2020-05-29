package entity

import "time"

type LocationMap map[string]Location

func (m LocationMap) Get(name string) Location {
	if result, ok := m[name]; ok {
		return result
	}

	return UnknownLocation
}

func (m LocationMap) Pointer(name string) *Location {
	if result, ok := m[name]; ok {
		return &result
	}

	return &UnknownLocation
}

var LocationFixtures = LocationMap{
	"mexico": {
		LocUID:      "85d1ea7d382c",
		PlaceUID:    PlaceFixtures.Get("mexico").PlaceUID,
		LocName:     "Adosada Platform",
		LocCategory: "botanical garden",
		Place:       PlaceFixtures.Pointer("mexico"),
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"caravan park": {
		LocUID:   "1ef75a71a36c",
		PlaceUID: "1ef75a71a36c",
		Place: &Place{
			PlaceUID:   "1ef75a71a36",
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
		LocUID:      "1ef744d1e28c",
		PlaceUID:    PlaceFixtures.Get("zinkwazi").PlaceUID,
		Place:       PlaceFixtures.Pointer("zinkwazi"),
		LocName:     "Zinkwazi Beach",
		LocCategory: "beach",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"hassloch": {
		LocUID:      "1ef744d1e280",
		PlaceUID:    PlaceFixtures.Get("holidaypark").PlaceUID,
		Place:       PlaceFixtures.Pointer("holidaypark"),
		LocName:     "Holiday Park",
		LocCategory: "park",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"emptyNameLongCity": {
		LocUID:      "1ef744d1e281",
		PlaceUID:    PlaceFixtures.Get("emptyNameLongCity").PlaceUID,
		Place:       PlaceFixtures.Pointer("emptyNameLongCity"),
		LocName:     "",
		LocCategory: "botanical garden",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"emptyNameShortCity": {
		LocUID:      "1ef744d1e282",
		PlaceUID:    PlaceFixtures.Get("emptyNameShortCity").PlaceUID,
		Place:       PlaceFixtures.Pointer("emptyNameShortCity"),
		LocName:     "",
		LocCategory: "botanical garden",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"veryLongLocName": {
		LocUID:      "1ef744d1e283",
		PlaceUID:    PlaceFixtures.Get("veryLongLocName").PlaceUID,
		Place:       PlaceFixtures.Pointer("veryLongLocName"),
		LocName:     "longlonglonglonglonglonglonglonglonglonglonglonglongName",
		LocCategory: "cape",
		LocSource:   "places",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	"mediumLongLocName": {
		LocUID:      "1ef744d1e283",
		PlaceUID:    PlaceFixtures.Get("mediumLongLocName").PlaceUID,
		Place:       PlaceFixtures.Pointer("mediumLongLocName"),
		LocName:     "longlonglonglonglonglongName",
		LocCategory: "botanical garden",
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
