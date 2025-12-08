//nolint:gocritic // explicit branching retained for clarity
package places

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Search finds and returns matching locations based on the specified query string.
func Search(q, locale string, count int) (results SearchResults, err error) {
	q = clean.SearchString(q)
	locale = clean.WebLocale(locale, "en")

	if q == "" {
		return results, ErrMissingQuery
	}

	if count <= 0 {
		count = 10
	} else if count > 50 {
		count = 50
	}

	// Remember start time.
	start := time.Now()

	// Generate query parameter string.
	values := url.Values{"q": {q}, "count": {strconv.Itoa(count)}}
	params := values.Encode()

	// Get request locale.
	locale = Locale(locale)

	// Create cache key based on query parameters.
	cacheKey := fmt.Sprintf("search:%s:%s", params, locale)

	// Are location results cached?
	if hit, ok := clientCache.Get(cacheKey); ok {
		log.Tracef("places: cache hit for %s [%s]", cacheKey, time.Since(start))
		cached := hit.(SearchResults)
		return cached, nil
	}

	var r *http.Response

	// Query the specified places service URLs.
	for _, serviceUrl := range SearchServiceUrls {
		reqUrl := fmt.Sprintf("%s?%s", serviceUrl, params)
		if r, err = GetRequest(reqUrl, locale); err == nil {
			break
		}
	}

	// Failed?
	if err != nil {
		log.Errorf("places: %s (search request failed)", err.Error())
		return results, err
	} else if r == nil {
		err = fmt.Errorf("search request could not be performed")
		return results, err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("search request failed with code %d", r.StatusCode)
		return results, err
	}

	// Decode JSON response body.
	err = json.NewDecoder(r.Body).Decode(&results)

	if err != nil {
		log.Errorf("places: %s (decode search results)", err.Error())
		return results, err
	}

	clientCache.SetDefault(cacheKey, results)
	log.Tracef("places: cached %s [%s]", clean.Log(cacheKey), time.Since(start))

	return results, nil
}
