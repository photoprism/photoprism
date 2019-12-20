package maps

import (
	"errors"
	"strings"

	olc "github.com/google/open-location-code/go"
	"github.com/photoprism/photoprism/internal/maps/osm"
)

const SourceOSM = "osm"

// Photo location
type Location struct {
	ID         string `gorm:"primary_key"`
	LocLat     float64
	LocLng     float64
	LocTitle   string
	LocCity    string
	LocSuburb  string
	LocState   string
	LocCountry string
	LocRegion  string
	LocLabel   string
	LocSource  string
}

func (l *Location) Query(lat, lng float64) error {
	return l.QueryOpenStreetMap(lat, lng)
}

func (l *Location) QueryOpenStreetMap(lat, lng float64) error {
	o, err := osm.FindLocation(lat, lng)

	if err != nil {
		return err
	}

	return l.OpenStreetMap(o)
}


func (l *Location) OpenStreetMap(o osm.Location) error {
	l.LocSource = SourceOSM

	l.LocLat = o.Latitude()
	l.LocLng = o.Longitude()

	if l.Unknown() {
		l.LocLabel = "unknown"
		return errors.New("maps: unknown location")
	}

	l.ID = olc.Encode(l.LocLat, l.LocLng, 11)
	l.LocTitle = o.Title()
	l.LocCity = o.City()
	l.LocSuburb = o.Suburb()
	l.LocState = o.State()
	l.LocCountry = o.Country()
	l.LocRegion = l.region()
	l.LocLabel = o.Label()

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

	var countryName = Countries[l.LocCountry]
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
