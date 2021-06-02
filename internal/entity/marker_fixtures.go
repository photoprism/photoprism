package entity

type MarkerMap map[string]Marker

func (m MarkerMap) Get(name string) Marker {
	if result, ok := m[name]; ok {
		return result
	}

	return *UnknownMarker
}

func (m MarkerMap) Pointer(name string) *Marker {
	if result, ok := m[name]; ok {
		return &result
	}

	return UnknownMarker
}

var MarkerFixtures = MarkerMap{
	"1000003-1": Marker{
		FileID:     1000003,
		RefUID:     "lt9k3pw1wowuy3c3",
		MarkerSrc:  SrcImage,
		MarkerType: MarkerLabel,
		X:          0.308333,
		Y:          0.206944,
		W:          0.355556,
		H:          .355556,
	},
	"1000003-2": Marker{
		FileID:      1000003,
		RefUID:      "",
		MarkerLabel: "Unknown",
		MarkerSrc:   SrcImage,
		MarkerType:  MarkerLabel,
		X:           0.208333,
		Y:           0.106944,
		W:           0.05,
		H:           0.05,
	},
	"1000003-3": Marker{
		FileID:      1000003,
		RefUID:      "",
		MarkerSrc:   SrcImage,
		MarkerType:  MarkerLabel,
		MarkerLabel: "Center",
		X:           0.5,
		Y:           0.5,
		W:           0,
		H:           0,
	},
}

// CreateMarkerFixtures inserts known entities into the database for testing.
func CreateMarkerFixtures() {
	for _, entity := range MarkerFixtures {
		Db().Create(&entity)
	}
}
