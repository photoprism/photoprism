//nolint:gocritic // explicit branching retained for clarity
package places

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/geo/s2"
)

// Cell returns location details based on the specified S2 cell ID.
func Cell(id string, locale string) (result Location, err error) {

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

	// Get request locale.
	locale = Locale(locale)

	// Create cache key based on query parameters.
	cacheKey := fmt.Sprintf("id:%s:%s", id, locale)

	// Location details cached?
	if hit, ok := clientCache.Get(cacheKey); ok {
		log.Tracef("places: cache hit for %s [%s]", cacheKey, time.Since(start))
		cached := hit.(Location)
		cached.Cached = true
		return cached, nil
	}

	var r *http.Response

	// Query the specified places service URLs.
	for _, serviceUrl := range LocationServiceUrls {
		reqUrl := fmt.Sprintf(serviceUrl, id)
		if r, err = GetRequest(reqUrl, locale); err == nil {
			break
		}
	}

	// Failed?
	if err != nil {
		log.Errorf("places: %s (location request failed)", err.Error())
		return result, err
	} else if r == nil {
		err = fmt.Errorf("location request could not be performed")
		return result, err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("location request failed with code %d", r.StatusCode)
		return result, err
	}

	// Decode JSON response body.
	err = json.NewDecoder(r.Body).Decode(&result)

	if err != nil {
		log.Errorf("places: %s (decode location failed)", err.Error())
		return result, err
	}

	if result.ID == "" {
		return result, fmt.Errorf("no location result for %s", id)
	}

	clientCache.SetDefault(cacheKey, result)
	log.Tracef("places: cached %s [%s]", clean.Log(cacheKey), time.Since(start))
	result.Cached = false

	return result, nil
}
