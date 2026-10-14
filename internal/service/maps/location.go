//revive:disable:exported // getters intentionally exported for use across services
package maps

import (
	"errors"
	"strings"

	"github.com/photoprism/photoprism/internal/service/hub/places"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/geo/s2"
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

func (l *Location) QueryApi(api string) error { //nolint:gocritic // single-case switch retained for future APIs
	if api == places.ApiName {
		return l.QueryPlaces()
	}

	return errors.New("maps: location lookup disabled")
}

func (l *Location) QueryPlaces() error {
	s, err := places.Cell(l.ID, places.DefaultLocale)

	if err != nil {
		return err
	}

	l.setPlace(s)

	return nil
}

// setPlace copies the values of a place from the remote service, which are single-line text.
func (l *Location) setPlace(s places.Location) {
	l.placeID = clean.StripControl(s.PlaceID())
	l.LocSource = clean.StripControl(s.Source())
	l.LocName = clean.StripControl(s.Name())
	l.LocStreet = clean.StripControl(s.Street())
	l.LocPostcode = clean.StripControl(s.Postcode())
	l.LocCategory = clean.StripControl(s.Category())
	l.LocLabel = clean.StripControl(s.Label())
	l.LocDistrict = clean.StripControl(s.District())
	l.LocCity = clean.StripControl(s.City())
	l.LocState = clean.StripControl(s.State())
	l.LocCountry = clean.StripControl(s.CountryCode())
	s.Place.LocKeywords = clean.StripControl(s.Place.LocKeywords)
	l.LocKeywords = s.Keywords()
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
