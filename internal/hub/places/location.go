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
	LocCategory string  `json:"category"`
	Place       Place   `json:"place"`
	Cached      bool    `json:"-"`
}

const ApiName = "places"

var Key = "f60f5b25d59c397989e3cd374f81cdd7710a4fca"
var Secret = "photoprism"
var UserAgent = "PhotoPrism/0.0.0"
var ReverseLookupURL = "https://places.photoprism.app/v1/location/%s"
var client = &http.Client{Timeout: 60 * time.Second}

func NewLocation(id string, lat, lng float64, name, category string, place Place, cached bool) *Location {
	result := &Location{
		ID:          id,
		LocLat:      lat,
		LocLng:      lng,
		LocName:     name,
		LocCategory: category,
		Place:       place,
		Cached:      cached,
	}

	return result
}

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
		log.Debugf("places: cache hit for lat %f, lng %f", lat, lng)
		cached := hit.(Location)
		cached.Cached = true
		return cached, nil
	}

	url := fmt.Sprintf(ReverseLookupURL, id)

	log.Debugf("places: sending request to %s", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Errorf("places: %s", err.Error())
		return result, err
	}

	req.Header.Set("User-Agent", UserAgent)

	if Key != "" {
		req.Header.Set("X-Key", Key)
		req.Header.Set("X-Signature", fmt.Sprintf("%x", sha1.Sum([]byte(Key+url+Secret))))
	}

	var r *http.Response

	for i := 0; i < 3; i++ {
		r, err = client.Do(req)

		if err == nil {
			break
		}
	}

	if err != nil {
		log.Errorf("places: %s (http request)", err.Error())
		return result, err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("request failed with code %d (%s)", r.StatusCode, ApiName)
		return result, err
	}

	err = json.NewDecoder(r.Body).Decode(&result)

	if err != nil {
		log.Errorf("places: %s (decode json)", err.Error())
		return result, err
	}

	if result.ID == "" {
		return result, fmt.Errorf("no result for %s (%s)", id, ApiName)
	}

	cache.SetDefault(id, result)
	log.Debugf("places: cached cell %s [%s]", id, time.Since(start))

	result.Cached = false

	return result, nil
}

func (l Location) CellID() (result string) {
	return l.ID
}

func (l Location) Name() (result string) {
	return strings.SplitN(l.LocName, "/", 2)[0]
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
