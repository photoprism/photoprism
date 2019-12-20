package maps

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/internal/maps/osm"
)

// Photo location
type Location struct {
	ID             string `gorm:"primary_key"`
	CountryID      string
	LocLat         float64
	LocLng         float64
	LocCategory    string
	LocTitle       string
	LocDescription string
	LocCity        string
	LocSuburb      string
	LocState       string
	LocSource      string
}

type LocationSource interface {
	CountryCode() string
	Latitude() float64
	Longitude() float64
	Category() string
	Title() string
	City() string
	Suburb() string
	State() string
	Source() string
}

func NewLocation (lat, lng float64) *Location {
	id := OlcEncode(lat, lng)

	result := &Location{
		ID: id,
		LocLat: lat,
		LocLng: lng,
	}

	return result
}

func (l *Location) Query() error {
	o, err := osm.FindLocation(l.LocLat, l.LocLng)

	if err != nil {
		return err
	}

	return l.Assign(o)
}

func (l *Location) Assign(s LocationSource) error {
	l.LocSource = s.Source()

	if l.LocLat == 0 { l.LocLat = s.Latitude()	}
	if l.LocLng == 0 { l.LocLng = s.Longitude() }

	if l.Unknown() {
		l.LocCategory = "unknown"
		return errors.New("maps: unknown location")
	}

	if l.ID == "" { l.ID = OlcEncode(l.LocLat, l.LocLng) }

	l.LocTitle = s.Title()
	l.LocCity = s.City()
	l.LocSuburb = s.Suburb()
	l.LocState = s.State()
	l.CountryID = s.CountryCode()
	l.LocCategory = s.Category()
	l.LocDescription = l.description()

	return nil
}

func (l *Location) Unknown() bool {
	if l.LocLng == 0.0 && l.LocLat == 0.0 {
		return true
	}

	return false
}

func (l *Location) description() string {
	if l.Unknown() {
		return "Unknown"
	}

	var countryName = l.CountryName()
	var loc []string

	shortCountry := len([]rune(countryName)) <= 20
	shortCity := len([]rune(l.LocCity)) <= 20

	if shortCity && l.LocCity != "" {
		loc = append(loc, l.LocCity)
	}

	if shortCountry && l.LocState != "" && l.LocCity != l.LocState {
		loc = append(loc, l.LocState)
	}

	if countryName != "" {
		loc = append(loc, countryName)
	}

	return strings.Join(loc[:], ", ")
}

func (l Location) Latitude() float64 {
	return l.LocLat
}

func (l Location) Longitude() float64 {
	return l.LocLng
}

func (l Location) Title() string {
	return l.LocTitle
}

func (l Location) City() string {
	return l.LocCity
}

func (l Location) Suburb() string {
	return l.LocSuburb
}

func (l Location) State() string {
	return l.LocState
}

func (l Location) Category() string {
	return l.LocCategory
}

func (l Location) Source() string {
	return l.LocSource
}

func (l Location) Region() string {
	return l.LocDescription
}

func (l Location) CountryCode() string {
	return l.CountryID
}

func (l Location) CountryName() string {
	return CountryNames[l.CountryID]
}
