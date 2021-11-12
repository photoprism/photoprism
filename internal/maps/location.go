package maps

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/internal/hub/places"

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
	LocDistrict string
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
	District() string
	State() string
	Source() string
	Keywords() []string
}

func NewLocation(id, name, category, label, city, district, state, country, source string, keywords []string) *Location {
	result := &Location{
		ID:          id,
		LocName:     name,
		LocCategory: category,
		LocLabel:    label,
		LocCity:     city,
		LocDistrict: district,
		LocCountry:  country,
		LocState:    txt.NormalizeState(state, country),
		LocSource:   source,
		LocKeywords: keywords,
	}

	return result
}

func (l *Location) QueryApi(api string) error {
	switch api {
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
	l.LocDistrict = s.District()
	l.LocState = s.State()
	l.LocCountry = s.CountryCode()
	l.LocCategory = s.Category()
	l.LocLabel = s.Label()
	l.LocKeywords = s.Keywords()

	return nil
}

func (l *Location) Unknown() bool {
	return l.ID == ""
}

func (l Location) S2Token() string {
	return l.ID
}

func (l Location) PrefixedToken() string {
	return s2.Prefix(l.ID)
}

func (l Location) Name() string {
	return txt.Shorten(l.LocName, txt.ClipTitle, txt.Ellipsis)
}

func (l Location) Category() string {
	return txt.Shorten(l.LocCategory, txt.ClipKeyword, txt.Ellipsis)
}

func (l Location) Label() string {
	return txt.Shorten(l.LocLabel, txt.ClipLabel, txt.Ellipsis)
}

func (l Location) City() string {
	return txt.Shorten(l.LocCity, txt.ClipPlace, txt.Ellipsis)
}

func (l Location) District() string {
	return txt.Shorten(l.LocDistrict, txt.ClipPlace, txt.Ellipsis)
}

func (l Location) CountryCode() string {
	return txt.Clip(l.LocCountry, txt.ClipCountryCode)
}

func (l Location) State() string {
	return txt.Shorten(txt.NormalizeState(l.LocState, l.CountryCode()), txt.ClipPlace, txt.Ellipsis)
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
	return txt.Clip(strings.Join(l.LocKeywords, ", "), txt.ClipVarchar)
}
