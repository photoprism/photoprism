package maps

import (
	"errors"
	"strings"

	olc "github.com/google/open-location-code/go"
	"github.com/photoprism/photoprism/internal/maps/osm"
)

// Photo location
type Location struct {
	ID             string `gorm:"primary_key"`
	LocLat         float64
	LocLng         float64
	LocTitle       string
	LocRegion      string
	LocCity        string
	LocSuburb      string
	LocState       string
	LocCountryCode string
	LocLabel       string
	LocSource      string
}

type LocationSource interface {
	Latitude() float64
	Longitude() float64
	Title() string
	City() string
	Suburb() string
	State() string
	CountryCode() string
	Label() string
	Source() string
}

func NewLocation (lat, lng float64) *Location {
	id := olc.Encode(lat, lng, 11)

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
		l.LocLabel = "unknown"
		return errors.New("maps: unknown location")
	}

	if l.ID == "" { l.ID = olc.Encode(l.LocLat, l.LocLng, 11) }

	l.LocTitle = s.Title()
	l.LocCity = s.City()
	l.LocSuburb = s.Suburb()
	l.LocState = s.State()
	l.LocCountryCode = s.CountryCode()
	l.LocLabel = s.Label()
	l.LocRegion = l.region()

	return nil
}

func (l *Location) Unknown() bool {
	if l.LocLng == 0.0 && l.LocLat == 0.0 {
		return true
	}

	return false
}

func (l *Location) region() string {
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

func (l Location) Label() string {
	return l.LocLabel
}

func (l Location) Source() string {
	return l.LocSource
}

func (l Location) Region() string {
	return l.LocRegion
}

func (l Location) CountryCode() string {
	return l.LocCountryCode
}

func (l Location) CountryName() string {
	return CountryNames[l.LocCountryCode]
}
