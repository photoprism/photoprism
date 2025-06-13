package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

type PlaceSearchResult struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Formatted string  `json:"formatted"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

type PlaceSearchResponse struct {
	Results []PlaceSearchResult `json:"results"`
	Count   int                 `json:"count"`
}

type PhotoPrismPlacesSearchResponse []struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	City    string  `json:"city"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

// GetPlaceSearch performs a place search using PhotoPrism Places API.
//
// GET /api/v1/maps/places/search?q=query&locale=en&count=10
//
//	@Summary		Search for places using text query
//	@Id				GetPlaceSearch
//	@Tags			Maps
//	@Produce		json
//	@Param			q		query		string	true	"Search query"
//	@Param			locale	query		string	false	"Locale for results (default: en)"
//	@Param			count	query		int		false	"Maximum number of results (default: 10, max: 50)"
//	@Success		200		{object}	PlaceSearchResponse
//	@Failure		400		{object}	gin.H	"Missing search query"
//	@Failure		401		{object}	ErrorResponse	"Not authorized"
//	@Failure		500		{object}	gin.H	"Search service error"
//	@Router			/maps/places/search [get]

func GetPlaceSearch(router *gin.RouterGroup) {
	handler := func(c *gin.Context) {
		s := AuthAny(c, acl.ResourcePlaces, acl.Permissions{acl.ActionSearch, acl.ActionView})

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		// Parse query parameters
		query := txt.Clip(c.Query("q"), 200)
		locale := txt.Clip(c.Query("locale"), 10)
		countStr := c.Query("count")

		if query == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing search query"})
			return
		}

		// Set default locale if not provided
		if locale == "" {
			locale = "en"
		}

		// Parse count parameter
		count := 10 // default
		if countStr != "" {
			if parsedCount, err := strconv.Atoi(countStr); err == nil && parsedCount > 0 && parsedCount <= 50 {
				count = parsedCount
			}
		}

		event.AuditInfo([]string{ClientIP(c), "session %s", "place search", "query %s, locale %s, count %d"}, s.RefID, query, locale, count)

		client := &http.Client{Timeout: 10 * time.Second}

		baseURL := "https://places.photoprism.app/v1/search"
		params := url.Values{}
		params.Add("q", query)
		params.Add("locale", locale)
		params.Add("count", strconv.Itoa(count))
		requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

		req, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "place search", "error creating request: %s"}, s.RefID, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create search request"})
			return
		}

		// Execute request
		resp, err := client.Do(req)
		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "place search", "request failed: %s"}, s.RefID, err)

			searchResponse := PlaceSearchResponse{
				Results: []PlaceSearchResult{},
				Count:   0,
			}
			c.JSON(http.StatusOK, searchResponse)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			event.AuditWarn([]string{ClientIP(c), "session %s", "place search", "status code %d"}, s.RefID, resp.StatusCode)
			// Return empty results instead of error for non-200 responses
			searchResponse := PlaceSearchResponse{
				Results: []PlaceSearchResult{},
				Count:   0,
			}
			c.JSON(http.StatusOK, searchResponse)
			return
		}

		// Parse response
		var placesResponse PhotoPrismPlacesSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&placesResponse); err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "place search", "decode failed: %s"}, s.RefID, err)
			searchResponse := PlaceSearchResponse{
				Results: []PlaceSearchResult{},
				Count:   0,
			}
			c.JSON(http.StatusOK, searchResponse)
			return
		}

		results := make([]PlaceSearchResult, 0, len(placesResponse))
		for _, place := range placesResponse {
			if place.ID == "" || place.Name == "" {
				continue
			}

			result := PlaceSearchResult{
				ID:        place.ID,
				Name:      place.Name,
				Formatted: place.Name,
				City:      place.City,
				Country:   place.Country,
				Latitude:  place.Lat,
				Longitude: place.Lng,
			}
			results = append(results, result)
		}

		searchResponse := PlaceSearchResponse{
			Results: results,
			Count:   len(results),
		}

		c.JSON(http.StatusOK, searchResponse)
	}

	router.GET("/maps/places/search", handler)
}
