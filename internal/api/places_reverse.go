package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/hub/places"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GetPlacesReverse returns location details for the specified coordinates.
//
// GET /api/v1/places/reverse?lat=12.444526469291622&lng=-69.94435584903263
//
//	@Summary	returns location details for the specified coordinates
//	@Id			GetPlacesReverse
//	@Tags		Places
//	@Produce	json
//	@Param		lat		query		string	true	"Latitude"
//	@Param		lng		query		string	true	"Longitude"
//	@Param		locale	query		string	false	"Locale"
//	@Success	200		{object}	places.Location
//	@Failure	400		{object}	gin.H	"Missing latitude or longitude"
//	@Failure	401		{object}	i18n.Response
//	@Failure	500		{object}	gin.H	"Geocoding service error"
//	@Router		/api/v1/places/reverse [get]
func GetPlacesReverse(router *gin.RouterGroup) {
	handler := func(c *gin.Context) {
		// Allow request if user is allowed to search places.
		s := AuthAny(c, acl.ResourcePlaces, acl.Permissions{acl.ActionSearch, acl.ActionView, acl.ActionUse})

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		// Abort if geocoding is disabled.
		conf := get.Config()

		if conf.DisablePlaces() {
			AbortFeatureDisabled(c)
			return
		}

		// Get latitude, longitude, and locale from query parameters.
		var lat, lng string

		if lat = txt.Numeric(c.Query("lat")); lat == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing latitude"})
			return
		}

		if lng = txt.Numeric(c.Query("lng")); lng == "" {
			lng = txt.Numeric(c.Query("lon"))
		}

		if lng == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing longitude"})
			return
		}

		locale := clean.WebLocale(c.Query("locale"), conf.PlacesLocale())

		// Perform service request.
		result, err := places.LatLng(txt.Float64(lat), txt.Float64(lng), locale)

		// Return error if request was not successful.
		if err != nil {
			log.Errorf("places: failed to resolve location at lat %s, lng %s", lat, lng)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		// Return location details.
		c.JSON(http.StatusOK, result)
	}

	router.GET("/places/reverse", handler)
}
