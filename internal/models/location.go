package models

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
