package osm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/melihmucuk/geocache"

	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

type Location struct {
	ID             string  `json:"-"`
	PlaceID        int     `json:"place_id"`
	LocName        string  `json:"name"`
	LocCategory    string  `json:"category"`
	LocType        string  `json:"type"`
	LocDisplayName string  `json:"display_name"`
	Address        Address `json:"address"`
	Cached         bool
}

var ReverseLookupURL = "https://nominatim.openstreetmap.org/reverse?lat=%f&lon=%f&format=jsonv2&accept-language=en&zoom=18"

// FindLocation retrieves geolocation data, docs see https://wiki.openstreetmap.org/wiki/Nominatim#Reverse_Geocoding
func FindLocation(id string) (result Location, err error) {
	if len(id) > 16 || len(id) == 0 {
		return result, errors.New("osm: invalid location id")
	}

	lat, lng := s2.LatLng(id)

	if lat == 0.0 || lng == 0.0 {
		return result, fmt.Errorf("osm: skipping lat %f, lng %f", lat, lng)
	}

	point := geocache.GeoPoint{Latitude: lat, Longitude: lng}

	if hit, ok := geoCache.Get(point); ok {
		log.Debugf("osm: cache hit for lat %f, lng %f", lat, lng)
		result = hit.(Location)
		result.Cached = true
		return result, nil
	}

	url := fmt.Sprintf(ReverseLookupURL, lat, lng)

	log.Debugf("osm: query %s", url)

	r, err := http.Get(url)

	if err != nil {
		log.Errorf("osm: %s", err.Error())
		return result, err
	}

	err = json.NewDecoder(r.Body).Decode(&result)

	if err != nil {
		log.Errorf("osm: %s", err.Error())
		return result, err
	}

	if result.PlaceID == 0 {
		result.ID = ""

		return result, fmt.Errorf("osm: no result for %s", id)
	}

	result.ID = id

	geoCache.Set(point, result, time.Hour)

	result.Cached = false

	return result, nil
}

func (l Location) CellID() (result string) {
	return l.ID
}

func (l Location) State() (result string) {
	return txt.NormalizeState(l.Address.State)
}

func (l Location) City() (result string) {
	if l.Address.City != "" {
		result = l.Address.City
	} else if l.Address.Town != "" {
		result = l.Address.Town
	} else if l.Address.Village != "" {
		result = l.Address.Village
	} else if l.Address.County != "" {
		result = l.Address.County
	} else if l.Address.State != "" {
		result = l.Address.State
	}

	if len([]rune(result)) > 19 {
		result = ""
	}

	return strings.TrimSpace(result)
}

func (l Location) Suburb() (result string) {
	result = l.Address.Suburb

	return strings.TrimSpace(result)
}

func (l Location) CountryCode() (result string) {
	result = l.Address.CountryCode

	return strings.ToLower(strings.TrimSpace(result))
}

func (l Location) Keywords() (result []string) {
	return result
}

func (l Location) Source() string {
	return "osm"
}
