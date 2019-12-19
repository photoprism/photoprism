package entity

var locTypeLabels = map[string]string{
	"bay": "bay",
	"art": "art exhibition",
	"fire station": "fire station",
	"hairdresser": "hairdresser",
	"cape": "cape",
	"coastline": "coastline",
	"cliff": "cliff",
	"wetland": "wetland",
	"nature reserve": "nature reserve",
	"beach": "beach",
	"cafe": "cafe",
	"internet cafe": "cafe",
	"ice cream": "ice cream parlor",
	"bistro": "restaurant",
	"restaurant": "restaurant",
	"ship": "ship",
	"wholesale": "shop",
	"food": "shop",
	"supermarket": "supermarket",
	"florist": "florist",
	"pharmacy": "pharmacy",
	"seafood": "seafood",
	"clothes": "clothing store",
	"residential": "residential area",
	"museum": "museum",
	"castle": "castle",
	"terminal": "airport",
	"ferry terminal": "harbor",
	"bridge": "bridge",
	"university": "university",
	"mall": "mall",
	"marina": "marina",
	"garden": "garden",
	"pedestrian": "shopping area",
	"bunker": "bunker",
	"viewpoint": "viewpoint",
	"train station": "train station",
	"farm": "farm",
}

// Photo location
type Location struct {
	Model
	LocDisplayName string
	LocLat         float64
	LocLong        float64
	LocCategory    string
	LocType        string
	LocName        string
	LocHouseNr     string
	LocStreet      string
	LocSuburb      string
	LocCity        string
	LocPostcode    string
	LocCounty      string
	LocState       string
	LocCountry     string
	LocCountryCode string
	LocDescription string `gorm:"type:text;"`
	LocNotes       string `gorm:"type:text;"`
	LocPhoto       *Photo
	LocPhotoID     uint
	LocFavorite    bool
}

func (m *Location) Label() string {
	return locTypeLabels[m.LocType]
}
