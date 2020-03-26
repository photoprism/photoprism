package form

import (
	"time"

	"github.com/ulule/deepcopier"
)

// Photo represents a photo edit form.
type Photo struct {
	TakenAt     time.Time `json:"TakenAt"`
	PhotoTitle  string    `json:"PhotoTitle"`
	Description struct {
		PhotoID          uint   `json:"PhotoID" deepcopier:"skip"`
		PhotoDescription string `json:"PhotoDescription"`
		PhotoKeywords    string `json:"PhotoKeywords"`
		PhotoNotes       string `json:"PhotoNotes"`
		PhotoSubject     string `json:"PhotoSubject"`
		PhotoArtist      string `json:"PhotoArtist"`
		PhotoCopyright   string `json:"PhotoCopyright"`
		PhotoLicense     string `json:"PhotoLicense"`
	} `json:"Description"`
	PhotoFavorite    bool      `json:"PhotoFavorite"`
	PhotoPrivate     bool      `json:"PhotoPrivate"`
	PhotoNSFW        bool      `json:"PhotoNSFW"`
	PhotoStory       bool      `json:"PhotoStory"`
	PhotoLat         float64   `json:"PhotoLat"`
	PhotoLng         float64   `json:"PhotoLng"`
	PhotoAltitude    int       `json:"PhotoAltitude"`
	PhotoFocalLength int       `json:"PhotoFocalLength"`
	PhotoIso         int       `json:"PhotoIso"`
	PhotoFNumber        float64   `json:"PhotoFNumber"`
	PhotoExposure       string    `json:"PhotoExposure"`
	CameraID            uint      `json:"CameraID"`
	LensID              uint      `json:"LensID"`
	LocationID          string    `json:"LocationID"`
	PlaceID             string    `json:"PlaceID"`
	PhotoCountry        string    `json:"PhotoCountry"`
	TimeZone            string    `json:"TimeZone"`
	TakenAtLocal        time.Time `json:"TakenAtLocal"`
	ModifiedTitle       bool      `json:"ModifiedTitle"`
	ModifiedDescription bool      `json:"ModifiedDescription"`
	ModifiedDate        bool      `json:"ModifiedDate"`
	ModifiedLocation    bool      `json:"ModifiedLocation"`
	ModifiedCamera      bool      `json:"ModifiedCamera"`
}

func NewPhoto(m interface{}) (f Photo, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
