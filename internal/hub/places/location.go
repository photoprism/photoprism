package places

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Location represents a specific geolocation identified by its S2 ID.
type Location struct {
	ID          string  `json:"id"`
	LocLat      float64 `json:"lat"`
	LocLng      float64 `json:"lng"`
	LocName     string  `json:"name"`
	LocStreet   string  `json:"street"`
	LocPostcode string  `json:"postcode"`
	LocCategory string  `json:"category"`
	Place       Place   `json:"place"`
	Cached      bool    `json:"-"`
}

const ApiName = "places"

var Key = "f60f5b25d59c397989e3cd374f81cdd7710a4fca"
var Secret = "photoprism"
var UserAgent = "PhotoPrism/dev"
var ReverseLookupURL = "https://places.photoprism.app/v1/location/%s"

var Retries = 3
var RetryDelay = 33 * time.Millisecond
var client = &http.Client{Timeout: 60 * time.Second}

func FindLocation(id string) (result Location, err error) {
	if len(id) > 16 || len(id) == 0 {
		return result, fmt.Errorf("invalid cell %s (%s)", id, ApiName)
	}

	start := time.Now()
	lat, lng := s2.LatLng(id)

	if lat == 0.0 || lng == 0.0 {
		return result, fmt.Errorf("skipping lat %f, lng %f (%s)", lat, lng, ApiName)
	}

	if hit, ok := cache.Get(id); ok {
		log.Tracef("places: cache hit for lat %f, lng %f", lat, lng)
		cached := hit.(Location)
		cached.Cached = true
		return cached, nil
	}

	// Compose request URL.
	url := fmt.Sprintf(ReverseLookupURL, id)

	log.Tracef("places: sending request to %s", url)

	// Create GET request instance.
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Errorf("places: %s", err.Error())
		return result, err
	}

	// Add User-Agent header?
	if UserAgent != "" {
		req.Header.Set("User-Agent", UserAgent)
	}

	// Add API key?
	if Key != "" {
		req.Header.Set("X-Key", Key)
		req.Header.Set("X-Signature", fmt.Sprintf("%x", sha1.Sum([]byte(Key+url+Secret))))
	}

	var r *http.Response

	// Perform request.
	for i := 0; i < Retries; i++ {
		r, err = client.Do(req)

		// Successful?
		if err == nil {
			break
		}

		// Wait before trying again?
		if RetryDelay.Nanoseconds() > 0 {
			time.Sleep(RetryDelay)
		}
	}

	// Failed?
	if err != nil {
		log.Errorf("places: %s (http request)", err.Error())
		return result, err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("request failed with code %d (%s)", r.StatusCode, ApiName)
		return result, err
	}

	// Decode JSON response body.
	err = json.NewDecoder(r.Body).Decode(&result)

	if err != nil {
		log.Errorf("places: %s (decode json)", err.Error())
		return result, err
	}

	if result.ID == "" {
		return result, fmt.Errorf("no result for %s (%s)", id, ApiName)
	}

	cache.SetDefault(id, result)
	log.Tracef("places: cached cell %s [%s]", id, time.Since(start))

	result.Cached = false

	return result, nil
}

func (l Location) CellID() string {
	return l.ID
}

func (l Location) PlaceID() string {
	return l.Place.PlaceID
}

func (l Location) Name() (result string) {
	return strings.SplitN(l.LocName, "/", 2)[0]
}

func (l Location) Street() (result string) {
	return strings.SplitN(l.LocStreet, "/", 2)[0]
}

func (l Location) Postcode() (result string) {
	return strings.SplitN(l.LocPostcode, "/", 2)[0]
}

func (l Location) Category() (result string) {
	return l.LocCategory
}

func (l Location) Label() (result string) {
	return l.Place.LocLabel
}

func (l Location) City() (result string) {
	return l.Place.LocCity
}

func (l Location) District() (result string) {
	return l.Place.LocDistrict
}

func (l Location) CountryCode() (result string) {
	return l.Place.LocCountry
}

func (l Location) State() (result string) {
	return txt.NormalizeState(l.Place.LocState, l.CountryCode())
}

func (l Location) Latitude() (result float64) {
	return l.LocLat
}

func (l Location) Longitude() (result float64) {
	return l.LocLng
}

func (l Location) Keywords() (result []string) {
	return txt.UniqueWords(txt.Words(l.Place.LocKeywords))
}

func (l Location) Source() string {
	return "places"
}
