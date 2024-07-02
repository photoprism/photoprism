package maps

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/internal/remote/hub/places"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Location represents a geolocation.
type Location struct {
	ID          string
	placeID     string
	LocName     string
	LocStreet   string
	LocPostcode string
	LocCategory string
	LocLabel    string
	LocDistrict string
	LocCity     string
	LocState    string
	LocCountry  string
	LocKeywords []string
	LocSource   string
}

type LocationSource interface {
	CellID() string
	PlaceID() string
	Name() string
	Street() string
	Category() string
	Postcode() string
	District() string
	City() string
	State() string
	CountryCode() string
	Keywords() []string
	Source() string
}

func (l *Location) QueryApi(api string) error {
	switch api {
	case places.ApiName:
		return l.QueryPlaces()
	}

	return errors.New("maps: location lookup disabled")
}

func (l *Location) QueryPlaces() error {
	s, err := places.FindLocation(l.ID)

	if err != nil {
		return err
	}

	l.placeID = s.PlaceID()
	l.LocSource = s.Source()
	l.LocName = s.Name()
	l.LocStreet = s.Street()
	l.LocPostcode = s.Postcode()
	l.LocCategory = s.Category()
	l.LocLabel = s.Label()
	l.LocDistrict = s.District()
	l.LocCity = s.City()
	l.LocState = s.State()
	l.LocCountry = s.CountryCode()
	l.LocKeywords = s.Keywords()

	return nil
}

func (l *Location) Unknown() bool {
	return l.ID == ""
}

func (l Location) PlaceID() string {
	return l.placeID
}

func (l Location) S2Token() string {
	return l.ID
}

func (l Location) PrefixedToken() string {
	return s2.Prefix(l.ID)
}

func (l Location) Name() string {
	return txt.Clip(l.LocName, 200)
}

func (l Location) Street() string {
	return txt.Clip(l.LocStreet, 100)
}

func (l Location) Postcode() string {
	return txt.Clip(l.LocPostcode, 50)
}

func (l Location) Category() string {
	return txt.Clip(l.LocCategory, 50)
}

func (l Location) Label() string {
	return txt.Clip(l.LocLabel, 400)
}

func (l Location) City() string {
	return txt.Clip(l.LocCity, 100)
}

func (l Location) District() string {
	return txt.Clip(l.LocDistrict, 100)
}

func (l Location) CountryCode() string {
	return txt.Clip(l.LocCountry, 2)
}

func (l Location) State() string {
	return txt.Clip(clean.State(l.LocState, l.CountryCode()), 100)
}

func (l Location) CountryName() string {
	return CountryNames[l.LocCountry]
}

func (l Location) Keywords() []string {
	return l.LocKeywords
}

func (l Location) KeywordString() string {
	return txt.Clip(strings.Join(l.LocKeywords, ", "), 300)
}

func (l Location) Source() string {
	return l.LocSource
}
