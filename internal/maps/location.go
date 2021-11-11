package maps

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/internal/hub/places"
	"github.com/photoprism/photoprism/internal/maps/osm"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Location represents a geolocation.
type Location struct {
	ID          string
	LocName     string
	LocCategory string
	LocLabel    string
	LocCity     string
	LocState    string
	LocCountry  string
	LocSource   string
	LocKeywords []string
}

type LocationSource interface {
	CellID() string
	CountryCode() string
	Category() string
	Name() string
	City() string
	State() string
	Source() string
	Keywords() []string
}

func NewLocation(id, name, category, label, city, state, country, source string, keywords []string) *Location {
	result := &Location{
		ID:          id,
		LocName:     name,
		LocCategory: category,
		LocLabel:    label,
		LocCity:     city,
		LocCountry:  country,
		LocState:    txt.NormalizeState(state, country),
		LocSource:   source,
		LocKeywords: keywords,
	}

	return result
}

func (l *Location) QueryApi(api string) error {
	switch api {
	case "osm":
		return l.QueryOSM()
	case "places":
		return l.QueryPlaces()
	}

	return errors.New("maps: reverse lookup disabled")
}

func (l *Location) QueryPlaces() error {
	s, err := places.FindLocation(l.ID)

	if err != nil {
		return err
	}

	l.LocSource = s.Source()
	l.LocName = s.Name()
	l.LocCity = s.City()
	l.LocState = s.State()
	l.LocCountry = s.CountryCode()
	l.LocCategory = s.Category()
	l.LocLabel = s.Label()
	l.LocKeywords = s.Keywords()

	return nil
}

func (l *Location) QueryOSM() error {
	s, err := osm.FindLocation(l.ID)

	if err != nil {
		return err
	}

	return l.Assign(s)
}

func (l *Location) Assign(s LocationSource) error {
	l.LocSource = s.Source()

	l.ID = s.CellID()

	if l.Unknown() {
		l.LocCategory = "unknown"
		return errors.New("maps: unknown location")
	}

	l.LocName = s.Name()
	l.LocCity = s.City()
	l.LocState = s.State()
	l.LocCountry = s.CountryCode()
	l.LocCategory = s.Category()
	l.LocLabel = l.label()
	l.LocKeywords = s.Keywords()

	return nil
}

func (l *Location) Unknown() bool {
	return l.ID == ""
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

	if shortCountry && l.LocState != "" && !strings.EqualFold(l.LocState, l.LocCity) && !strings.EqualFold(l.LocState, l.LocCountry) {
		loc = append(loc, l.LocState)
	}

	if countryName != "" {
		loc = append(loc, countryName)
	}

	return strings.Join(loc[:], ", ")
}

func (l Location) S2Token() string {
	return l.ID
}

func (l Location) PrefixedToken() string {
	return s2.Prefix(l.ID)
}

func (l Location) Name() string {
	return l.LocName
}

func (l Location) Category() string {
	return l.LocCategory
}

func (l Location) Label() string {
	return l.LocLabel
}

func (l Location) City() string {
	return l.LocCity
}

func (l Location) CountryCode() string {
	return l.LocCountry
}

func (l Location) State() string {
	return txt.NormalizeState(l.LocState, l.CountryCode())
}

func (l Location) CountryName() string {
	return CountryNames[l.LocCountry]
}

func (l Location) Source() string {
	return l.LocSource
}

func (l Location) Keywords() []string {
	return l.LocKeywords
}

func (l Location) KeywordString() string {
	return strings.Join(l.LocKeywords, ", ")
}
