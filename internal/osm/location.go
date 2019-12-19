package osm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/melihmucuk/geocache"
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
func FindLocation(lat, long float64) (result Location, err error) {
	if lat == 0.0 || long == 0.0 {
		return result, fmt.Errorf("osm: skipping lat %f / long %f", lat, long)
	}

	point := geocache.GeoPoint{Latitude: lat, Longitude: long}

	if hit, ok := geoCache.Get(point); ok {
		log.Debugf("osm: cache hit for lat %f / long %f", lat, long)
		result = hit.(Location)
		result.Cached = true
		return result, nil
	}

	url := fmt.Sprintf(ReverseLookupURL, lat, long)

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
