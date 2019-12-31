package maps

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/internal/maps/osm"
)

/* TODO

(SELECT pl.loc_label as album_name, pl.loc_country, YEAR(ph.taken_at) as taken_year, round(count(ph.id)) as photo_count FROM photos ph
        JOIN places pl ON ph.place_id = pl.id AND pl.id <> 1
        GROUP BY album_name, taken_year HAVING photo_count > 5) UNION (
            SELECT c.country_name AS album_name, pl.loc_country, YEAR(ph.taken_at) as taken_year, round(count(ph.id)) as photo_count FROM photos ph
        JOIN places pl ON ph.place_id = pl.id AND pl.id <> 1
            JOIN countries c ON c.id = pl.loc_country
        GROUP BY album_name, taken_year
        HAVING photo_count > 10)
ORDER BY loc_country, album_name, taken_year;

*/

// Photo location
type Location struct {
	ID          string
	LocLat      float64
	LocLng      float64
	LocName     string
	LocCategory string
	LocSuburb   string
	LocLabel    string
	LocCity     string
	LocState    string
	LocCountry  string
	LocSource   string
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
	id := S2Token(lat, lng)

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

	if l.ID == "" {
		l.ID = S2Token(l.LocLat, l.LocLng)
	}

	l.LocName = s.Name()
	l.LocCity = s.City()
	l.LocSuburb = s.Suburb()
	l.LocState = s.State()
	l.LocCountry = s.CountryCode()
	l.LocCategory = s.Category()
	l.LocLabel = l.label()

	return nil
}

func (l *Location) Unknown() bool {
	if l.LocLng == 0.0 && l.LocLat == 0.0 {
		return true
	}

	return false
}

func (l *Location) label() string {
	if l.Unknown() {
		return "Unknown"
	}

	var countryName = l.CountryName()
	var loc []string

	shortCountry := len([]rune(countryName)) <= 20

	if l.LocCity != "" {
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

func (l Location) Category() string {
	return l.LocCategory
}

func (l Location) Suburb() string {
	return l.LocSuburb
}

func (l Location) Label() string {
	return l.LocLabel
}

func (l Location) City() string {
	return l.LocCity
}

func (l Location) State() string {
	return l.LocState
}

func (l Location) CountryCode() string {
	return l.LocCountry
}

func (l Location) CountryName() string {
	return CountryNames[l.LocCountry]
}

func (l Location) Source() string {
	return l.LocSource
}
