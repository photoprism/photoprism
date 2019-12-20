package osm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/melihmucuk/geocache"
	"github.com/photoprism/photoprism/internal/util"
)

type Location struct {
	PlaceID     int    `json:"place_id"`
	Lat         string  `json:"lat"`
	Lon         string  `json:"lon"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Type        string  `json:"type"`
	DisplayName string  `json:"display_name"`
	Address     Address `json:"address"`
	Cached      bool
}

var ReverseLookupURL = "https://nominatim.openstreetmap.org/reverse?lat=%f&lon=%f&format=jsonv2&accept-language=en&zoom=18"

// API docs see https://wiki.openstreetmap.org/wiki/Nominatim#Reverse_Geocoding
func FindLocation(lat, lng float64) (result Location, err error) {
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

	geoCache.Set(point, result, time.Hour)

	result.Cached = false

	return result, nil
}

func (o Location) State() (result string) {
	result = o.Address.State

	return strings.TrimSpace(result)
}

func (o Location) City() (result string) {
	if o.Address.City != "" {
		result = o.Address.City
	} else if o.Address.Town != "" {
		result = o.Address.Town
	} else if o.Address.County != "" {
		result = o.Address.County
	} else if o.Address.State != "" {
		result = o.Address.State
	}

	return strings.TrimSpace(result)
}

func (o Location) Suburb() (result string) {
	result = o.Address.Suburb

	return strings.TrimSpace(result)
}

func (o Location) Country() (result string) {
	result = o.Address.CountryCode

	return strings.ToLower(strings.TrimSpace(result))
}

func (o Location) Latitude() (result float64) {
	result, err := strconv.ParseFloat(o.Lat, 64)

	if err != nil {
		log.Errorf("osm: %s", err.Error())
	}

	return result
}

func (o Location) Longitude() (result float64) {
	result, err := strconv.ParseFloat(o.Lon, 64)

	if err != nil {
		log.Errorf("osm: %s", err.Error())
	}

	return result
}

func (o Location) Keywords() (result []string) {
	return util.Keywords(o.DisplayName)
}

func (o Location) Source() string {
	return "osm"
}
