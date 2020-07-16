package form

import (
	"time"

	"github.com/ulule/deepcopier"
)

type Details struct {
	PhotoID   uint   `json:"PhotoID" deepcopier:"skip"`
	Keywords  string `json:"Keywords"`
	Notes     string `json:"Notes"`
	Subject   string `json:"Subject"`
	Artist    string `json:"Artist"`
	Copyright string `json:"Copyright"`
	License   string `json:"License"`
}

// Photo represents a photo edit form.
type Photo struct {
	PhotoType        string    `json:"Type"`
	TypeSrc          string    `json:"TypeSrc"`
	TakenAt          time.Time `json:"TakenAt"`
	TakenAtLocal     time.Time `json:"TakenAtLocal"`
	TakenSrc         string    `json:"TakenSrc"`
	TimeZone         string    `json:"TimeZone"`
	PhotoYear        int       `json:"Year"`
	PhotoMonth       int       `json:"Month"`
	PhotoDay         int       `json:"Day"`
	PhotoTitle       string    `json:"Title"`
	TitleSrc         string    `json:"TitleSrc"`
	PhotoDescription string    `json:"Description"`
	DescriptionSrc   string    `json:"DescriptionSrc"`
	Details          Details   `json:"Details"`
	PhotoFavorite    bool      `json:"Favorite"`
	PhotoPrivate     bool      `json:"Private"`
	PhotoReview      bool      `json:"Review"`
	PhotoScan        bool      `json:"Scan"`
	PhotoPanorama    bool      `json:"Panorama"`
	PhotoAltitude    int       `json:"Altitude"`
	PhotoLat         float32   `json:"Lat"`
	PhotoLng         float32   `json:"Lng"`
	PhotoIso         int       `json:"Iso"`
	PhotoFocalLength int       `json:"FocalLength"`
	PhotoFNumber     float32   `json:"FNumber"`
	PhotoExposure    string    `json:"Exposure"`
	PhotoCountry     string    `json:"Country"`
	CellID           string    `json:"CellID"`
	CellAccuracy     int       `json:"CellAccuracy"`
	PlaceID          string    `json:"PlaceID"`
	PlaceSrc         string    `json:"PlaceSrc"`
	CameraID         uint      `json:"CameraID"`
	CameraSrc        string    `json:"CameraSrc"`
	LensID           uint      `json:"LensID"`
	OriginalName     string    `json:"OriginalName"`
}

func NewPhoto(m interface{}) (f Photo, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
