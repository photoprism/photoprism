package places

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

const ApiName = "photoprism places"

var ReverseLookupURL = "https://places.photoprism.org/v1/location/%s"
var client = &http.Client{Timeout: 60 * time.Second} // TODO: Change timeout if needed

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
		return result, fmt.Errorf("api: invalid location id %s (%s)", id, ApiName)
	}

	start := time.Now()
	lat, lng := s2.LatLng(id)

	if lat == 0.0 || lng == 0.0 {
		return result, fmt.Errorf("api: skipping lat %f, lng %f (%s)", lat, lng, ApiName)
	}

	if hit, err := cache.Get(id); err == nil {
		log.Debugf("api: cache hit for lat %f, lng %f (%s)", lat, lng, ApiName)
		var cached Location
		if err := json.Unmarshal(hit, &cached); err != nil {
			log.Errorf("api: %s (%s)", err.Error(), ApiName)
		} else {
			cached.Cached = true
			return cached, nil
		}
	}

	url := fmt.Sprintf(ReverseLookupURL, id)

	log.Debugf("api: sending request to %s (%s)", url, ApiName)

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Errorf("api: %s (%s)", err.Error(), ApiName)
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
		log.Errorf("api: %s", err.Error())
		return result, err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("api: request failed with status code %d (%s)", r.StatusCode, ApiName)
		log.Error(err)
		return result, err
	}

	err = json.NewDecoder(r.Body).Decode(&result)

	if err != nil {
		log.Errorf("api: %s (%s)", err.Error(), ApiName)
		return result, err
	}

	if result.ID == "" {
		log.Debugf("api: %+v", result)
		return result, fmt.Errorf("api: no result for %s (%s)", id, ApiName)
	}

	if cached, err := json.Marshal(result); err == nil {
		if err := cache.Set(id, cached); err != nil {
			log.Errorf("api: %s (%s)", id, ApiName)
		} else {
			log.Debugf("cached %s [%s]", id, time.Since(start))
		}
	}

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
