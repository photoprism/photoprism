package places

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	gc "github.com/patrickmn/go-cache"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Location
type Location struct {
	ID          string  `json:"id"`
	LocLat      float64 `json:"lat"`
	LocLng      float64 `json:"lng"`
	LocName     string  `json:"name"`
	LocCategory string  `json:"category"`
	Place       Place   `json:"place"`
	Cached      bool    `json:"-"`
}

var ReverseLookupURL = "https://places.photoprism.org/v1/location/%s"
var client = &http.Client{Timeout: 30 * time.Second} // TODO: Change timeout if needed

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
		return result, fmt.Errorf("places: invalid location id %s", id)
	}

	lat, lng := s2.LatLng(id)

	if lat == 0.0 || lng == 0.0 {
		return result, fmt.Errorf("places: skipping lat %f, lng %f", lat, lng)
	}

	if hit, ok := cache.Get(id); ok {
		log.Debugf("places: cache hit for lat %f, lng %f", lat, lng)
		result = hit.(Location)
		result.Cached = true
		return result, nil
	}

	url := fmt.Sprintf(ReverseLookupURL, id)

	log.Debugf("places: query %s", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Errorf("places: %s", err.Error())
		return result, err
	}

	var r *http.Response

	for i := 0; i < 3; i++ {
		r, err = client.Do(req)

		if err == nil {
			break
		}
	}

	if err != nil {
		log.Errorf("places: %s", err.Error())
		return result, err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("places: request failed with status code %d", r.StatusCode)
		log.Error(err)
		return result, err
	}

	err = json.NewDecoder(r.Body).Decode(&result)

	if err != nil {
		log.Errorf("places: %s", err.Error())
		return result, err
	}

	if result.ID == "" {
		log.Debugf("result: %+v", result)
		return result, fmt.Errorf("places: no result for %s", id)
	}

	cache.Set(id, result, gc.DefaultExpiration)

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

func (l Location) State() (result string) {
	return l.Place.LocState
}

func (l Location) City() (result string) {
	return l.Place.LocCity
}

func (l Location) CountryCode() (result string) {
	return l.Place.LocCountry
}

func (l Location) Latitude() (result float64) {
	return l.LocLat
}

func (l Location) Longitude() (result float64) {
	return l.LocLng
}

func (l Location) Keywords() (result []string) {
	return txt.UniqueKeywords(l.Place.LocKeywords)
}

func (l Location) Source() string {
	return "places"
}
