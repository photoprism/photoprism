package maps

import (
	"strings"

	olc "github.com/google/open-location-code/go"
	"github.com/photoprism/photoprism/internal/maps/osm"
)

// Photo location
type Location struct {
	ID      string `gorm:"primary_key"`
	Lat     float64
	Lng     float64
	Title   string
	City    string
	Suburb  string
	State   string
	Country string
	Region  string
	Label   string
}

func (l *Location) FromOSM(o osm.Location) {
	l.Lat = o.Latitude()
	l.Lng = o.Longitude()

	if l.Unknown() {
		log.Warnf("maps: unknown location")
		l.Label = "unknown"
		return
	}

	l.ID = olc.Encode(l.Lat, l.Lng, 11)
	l.Title = o.Title()
	l.City = o.City()
	l.Suburb = o.Suburb()
	l.State = o.State()
	l.Country = o.Country()
	l.Region = l.region()
	l.Label = o.Label()
}

func (l *Location) Unknown() bool {
	if l.Lng == 0.0 && l.Lat == 0.0 {
		return true
	}

	return false
}

func (l *Location) region() string {
	if l.Unknown() {
		return "Unknown"
	}

	var countryName = Countries[l.Country]
	var loc []string
	shortCountry := len([]rune(countryName)) <= 20
	shortCity := len([]rune(l.City)) <= 20

	if shortCity && l.City != "" {
		loc = append(loc, l.City)
	}

	if shortCountry && l.State != "" && l.City != l.State {
		loc = append(loc, l.State)
	}

	if countryName != "" {
		loc = append(loc, countryName)
	}

	return strings.Join(loc[:], ", ")
}
