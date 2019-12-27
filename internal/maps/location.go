package maps

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/internal/maps/osm"
)

// Photo location
type Location struct {
	ID          uint64 `gorm:"type:BIGINT;primary_key;auto_increment:false;"`
	LocLat      float64
	LocLng      float64
	LocName     string `gorm:"type:varchar(100);"`
	LocCategory string `gorm:"type:varchar(50);"`
	LocSuburb   string `gorm:"type:varchar(100);"`
	LocPlace    string `gorm:"type:varbinary(500);index;"`
	LocCity     string `gorm:"type:varchar(100);"`
	LocState    string `gorm:"type:varchar(100);"`
	LocCountry  string `gorm:"type:binary(2);"`
	LocSource   string `gorm:"type:varbinary(16);"`
}

type LocationSource interface {
	CountryCode() string
	Latitude() float64
	Longitude() float64
	Category() string
	Name() string
	City() string
	Suburb() string
	State() string
	Source() string
}

func NewLocation(lat, lng float64) *Location {
	id := S2Encode(lat, lng)

	result := &Location{
		ID:     id,
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

	if l.LocLat == 0 {
		l.LocLat = s.Latitude()
	}
	if l.LocLng == 0 {
		l.LocLng = s.Longitude()
	}

	if l.Unknown() {
		l.LocCategory = "unknown"
		return errors.New("maps: unknown location")
	}

	if l.ID == 0 {
		l.ID = S2Encode(l.LocLat, l.LocLng)
	}

	l.LocName = s.Name()
	l.LocCity = s.City()
	l.LocSuburb = s.Suburb()
	l.LocState = s.State()
	l.LocCountry = s.CountryCode()
	l.LocCategory = s.Category()
	l.LocPlace = l.place()

	return nil
}

func (l *Location) Unknown() bool {
	if l.LocLng == 0.0 && l.LocLat == 0.0 {
		return true
	}

	return false
}

func (l *Location) place() string {
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

func (l Location) Name() string {
	return l.LocName
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

func (l Location) Place() string {
	return l.LocPlace
}

func (l Location) CountryCode() string {
	return l.LocCountry
}

func (l Location) CountryName() string {
	return CountryNames[l.LocCountry]
}
