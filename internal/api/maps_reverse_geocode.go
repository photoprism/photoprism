package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

// ReverseGeocodeResponse represents the response structure for the reverse geocode API.
type ReverseGeocodeResponse struct {
	Formatted string `json:"formatted"`
	Street    string `json:"street"`
}

// PhotoPrismPlacesResponse represents the response from the PhotoPrism Places API
type PhotoPrismPlacesResponse struct {
	Street string `json:"street"`
	Place  struct {
		Label    string `json:"label"`
		District string `json:"district"`
		City     string `json:"city"`
		State    string `json:"state"`
		Country  string `json:"country"`
	} `json:"place"`
}

// GetReverseGeocode performs a reverse geocoding lookup using PhotoPrism Places API.
//
// GET /api/v1/maps/geocode/reverse?lat=xx.xxx&lng=xx.xxx
//
//	@Summary		Reverse geocodes coordinates to a place name
//	@Id				GetReverseGeocode
//	@Tags			Maps
//	@Produce		json
//	@Param			lat	query		string	true	"Latitude"
//	@Param			lng	query		string	true	"Longitude"
//	@Success		200	{object}	ReverseGeocodeResponse
//	@Failure		400	{object}	gin.H	"Missing latitude or longitude"
//	@Failure		401	{object}	ErrorResponse	"Not authorized"
//	@Failure		500	{object}	gin.H	"Geocoding service error"
//	@Router			/maps/geocode/reverse [get]

func GetReverseGeocode(router *gin.RouterGroup) {
	handler := func(c *gin.Context) {
		s := AuthAny(c, acl.ResourcePlaces, acl.Permissions{acl.ActionSearch, acl.ActionView})

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		// Parse latitude and longitude
		lat := txt.Numeric(c.Query("lat"))
		lng := txt.Numeric(c.Query("lng"))

		if lat == "" || lng == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing latitude or longitude"})
			return
		}

		// Log request
		event.AuditInfo([]string{ClientIP(c), "session %s", "reverse geocoding", "lat %s, lng %s"}, s.RefID, lat, lng)

		// Create HTTP client with timeout
		client := &http.Client{Timeout: 10 * time.Second}

		// Create PhotoPrism Places API request
		url := fmt.Sprintf("https://places.photoprism.app/v1/latlng/%s/%s", lat, lng)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "reverse geocoding", "error creating request: %s"}, s.RefID, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create geocoding request"})
			return
		}

		// Execute request
		resp, err := client.Do(req)
		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "reverse geocoding", "request failed: %s"}, s.RefID, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Geocoding service unavailable"})
			return
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			event.AuditWarn([]string{ClientIP(c), "session %s", "reverse geocoding", "status code %d"}, s.RefID, resp.StatusCode)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Geocoding service returned status code %d", resp.StatusCode)})
			return
		}

		// Parse response
		var placesResponse PhotoPrismPlacesResponse
		if err := json.NewDecoder(resp.Body).Decode(&placesResponse); err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "reverse geocoding", "decode failed: %s"}, s.RefID, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode geocoding response"})
			return
		}

		// Prepare response
		geocodeResponse := ReverseGeocodeResponse{
			Formatted: placesResponse.Place.Label,
			Street:    placesResponse.Street,
		}

		c.JSON(http.StatusOK, geocodeResponse)
	}

	router.GET("/maps/geocode/reverse", handler)
}
