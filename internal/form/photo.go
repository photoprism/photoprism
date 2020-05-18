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
	TakenAt          time.Time `json:"TakenAt"`
	TakenAtLocal     time.Time `json:"TakenAtLocal"`
	TakenSrc         string    `json:"TakenSrc"`
	TimeZone         string    `json:"TimeZone"`
	PhotoTitle       string    `json:"PhotoTitle"`
	TitleSrc         string    `json:"TitleSrc"`
	PhotoDescription string    `json:"PhotoDescription"`
	DescriptionSrc   string    `json:"DescriptionSrc"`
	Details          Details   `json:"Details"`
	PhotoFavorite    bool      `json:"PhotoFavorite"`
	PhotoPrivate     bool      `json:"PhotoPrivate"`
	PhotoVideo       bool      `json:"PhotoVideo"`
	PhotoReview      bool      `json:"PhotoReview"`
	PhotoLat         float32   `json:"PhotoLat"`
	PhotoLng         float32   `json:"PhotoLng"`
	PhotoAltitude    int       `json:"PhotoAltitude"`
	PhotoIso         int       `json:"PhotoIso"`
	PhotoFocalLength int       `json:"PhotoFocalLength"`
	PhotoFNumber     float32   `json:"PhotoFNumber"`
	PhotoExposure    string    `json:"PhotoExposure"`
	CameraID         uint      `json:"CameraID"`
	CameraSrc        string    `json:"CameraSrc"`
	LensID           uint      `json:"LensID"`
	LocationID       string    `json:"LocationID"`
	LocationSrc      string    `json:"LocationSrc"`
	PlaceID          string    `json:"PlaceID"`
	PhotoCountry     string    `json:"PhotoCountry"`
}

func NewPhoto(m interface{}) (f Photo, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
