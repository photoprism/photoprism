package places

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Location represents a specific geolocation identified by its S2 ID.
type Location struct {
	ID          string  `json:"id"`
	Locale      string  `json:"locale,omitempty"`
	LocLat      float64 `json:"lat"`
	LocLng      float64 `json:"lng"`
	LocName     string  `json:"name"`
	LocStreet   string  `json:"street"`
	LocPostcode string  `json:"postcode"`
	LocCountry  string  `json:"country,omitempty"`
	LocCategory string  `json:"category"`
	TimeZone    string  `json:"timezone,omitempty"`
	Licence     string  `json:"licence,omitempty"`
	Place       Place   `json:"place,omitempty"`
	Cached      bool    `json:"-"`
}

// Locations represents a set of locations.
type Locations = []Location

// CellID returns the S2 cell identifier string.
func (l Location) CellID() string {
	return l.ID
}

// PlaceID returns the place identifier string.
func (l Location) PlaceID() string {
	return l.Place.PlaceID
}

// Name returns the location name if any.
func (l Location) Name() (result string) {
	return strings.SplitN(l.LocName, "/", 2)[0]
}

// Street returns the location street if any.
func (l Location) Street() (result string) {
	return strings.SplitN(l.LocStreet, "/", 2)[0]
}

// Postcode returns the location postcode if any.
func (l Location) Postcode() (result string) {
	return strings.SplitN(l.LocPostcode, "/", 2)[0]
}

// Category returns the location category if any.
func (l Location) Category() (result string) {
	return l.LocCategory
}

// Label returns the location label.
func (l Location) Label() (result string) {
	return l.Place.LocLabel
}

// City returns the location address city name.
func (l Location) City() (result string) {
	return l.Place.LocCity
}

// District returns the location address district name.
func (l Location) District() (result string) {
	return l.Place.LocDistrict
}

// CountryCode returns the location address country code.
func (l Location) CountryCode() (result string) {
	if l.LocCountry != "" && l.LocCountry != "zz" {
		return l.LocCountry
	}

	return l.Place.LocCountry
}

// State returns the location address state name.
func (l Location) State() (result string) {
	return clean.State(l.Place.LocState, l.CountryCode())
}

// Latitude returns the location position latitude.
func (l Location) Latitude() (result float64) {
	return l.LocLat
}

// Longitude returns the location position longitude.
func (l Location) Longitude() (result float64) {
	return l.LocLng
}

// Keywords returns location keywords if any.
func (l Location) Keywords() (result []string) {
	return txt.UniqueWords(txt.Words(l.Place.LocKeywords))
}

// Source returns the backend API name.
func (l Location) Source() string {
	return "places"
}
