package places

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/geo/s2"
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

// ApiName is the backend API name.
const ApiName = "places"

// ServiceUrls specifies the reverse geocoding service URLs.
var ServiceUrls = []string{
	"https://places.photoprism.app/v1/location/%s",
}

// Retries specifies the number of attempts to retry the service request.
var Retries = 2

// RetryDelay specifies the waiting time between retries.
var RetryDelay = 100 * time.Millisecond

var Key = "f60f5b25d59c397989e3cd374f81cdd7710a4fca"
var Secret = "photoprism"
var UserAgent = ""

// FindLocation retrieves location details from the backend API.
func FindLocation(id string) (result Location, err error) {

	// Normalize S2 Cell ID.
	id = s2.NormalizeToken(id)

	// Valid?
	if len(id) == 0 {
		return result, fmt.Errorf("empty cell id")
	} else if n := len(id); n < 4 || n > 16 {
		return result, fmt.Errorf("invalid cell id %s", clean.Log(id))
	}

	// Remember start time.
	start := time.Now()

	// Convert S2 Cell ID to latitude and longitude.
	lat, lng := s2.LatLng(id)

	// Return if latitude and longitude are null.
	if lat == 0.0 || lng == 0.0 {
		return result, fmt.Errorf("skipping lat %f, lng %f", lat, lng)
	}

	// Location details cached?
	if hit, ok := clientCache.Get(id); ok {
		log.Tracef("places: cache hit for lat %f, lng %f", lat, lng)
		cached := hit.(Location)
		cached.Cached = true
		return cached, nil
	}

	var r *http.Response

	// Try all the specified backend service URLs.
	for _, serviceUrl := range ServiceUrls {
		if r, err = PerformRequest(serviceUrl, id); err == nil {
			break
		}
	}

	// Failed?
	if err != nil {
		log.Errorf("places: %s (http request failed)", err.Error())
		return result, err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("request failed with code %d", r.StatusCode)
		return result, err
	}

	// Decode JSON response body.
	err = json.NewDecoder(r.Body).Decode(&result)

	if err != nil {
		log.Errorf("places: %s (decode json failed)", err.Error())
		return result, err
	}

	if result.ID == "" {
		return result, fmt.Errorf("no result for %s", id)
	}

	clientCache.SetDefault(id, result)
	log.Tracef("places: cached cell %s [%s]", clean.Log(id), time.Since(start))

	result.Cached = false

	return result, nil
}

// PerformRequest fetches the cell ID data from the service URL.
func PerformRequest(serviceUrl, id string) (r *http.Response, err error) {
	var req *http.Request

	// Compose request URL with S2 cell ID.
	reqUrl := fmt.Sprintf(serviceUrl, id)

	// Log request URL.
	log.Tracef("places: sending request to %s", reqUrl)

	// Create GET request instance.
	req, err = http.NewRequest(http.MethodGet, reqUrl, nil)

	// Ok?
	if err != nil {
		log.Errorf("places: %s", err.Error())
		return r, err
	}

	// Set user agent.
	if UserAgent != "" {
		req.Header.Set("User-Agent", UserAgent)
	} else {
		req.Header.Set("User-Agent", "PhotoPrism/Test")
	}

	// Add API key?
	if Key != "" {
		req.Header.Set("X-Key", Key)
		req.Header.Set("X-Signature", fmt.Sprintf("%x", sha1.Sum([]byte(Key+reqUrl+Secret))))
	}

	// Create new http.Client.
	//
	// NOTE: Timeout specifies a time limit for requests made by
	// this Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after Get, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
	client := &http.Client{Timeout: 60 * time.Second}

	// Perform request.
	for i := 0; i < Retries; i++ {
		r, err = client.Do(req)

		// Ok?
		if err == nil {
			return r, nil
		}

		// Wait before trying again?
		if RetryDelay.Nanoseconds() > 0 {
			time.Sleep(RetryDelay)
		}
	}

	return r, err
}

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
