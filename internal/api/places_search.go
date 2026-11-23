package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/hub/places"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GetPlacesSearch returns locations that match the specified search query.
//
// GET /api/v1/places/search?q=query&locale=en&count=10
//
//	@Summary	returns locations that match the specified search query
//	@Id			GetPlacesSearch
//	@Tags		Places
//	@Produce	json
//	@Param		q		query		string	true	"Search query"
//	@Param		locale	query		string	false	"Locale for results (default: en)"
//	@Param		count	query		int		false	"Maximum number of results (default: 10, max: 50)"
//	@Success	200		{object}	places.SearchResults
//	@Failure	400		{object}	i18n.Response
//	@Failure	401		{object}	i18n.Response
//	@Failure	500		{object}	i18n.Response
//	@Router		/api/v1/places/search [get]
func GetPlacesSearch(router *gin.RouterGroup) {
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

		// Get the search string, locale, and result count limit from the query parameters.
		query := clean.SearchString(c.Query("q"))
		locale := clean.WebLocale(c.Query("locale"), conf.PlacesLocale())
		count := txt.IntVal(c.Query("count"), 1, 50, 10)

		if query == "" {
			AbortBadRequest(c)
			return
		}

		results, err := places.Search(query, locale, count)

		if err != nil {
			event.SystemError([]string{"places", "find locations for query %s", "%s"}, clean.Log(query), clean.Error(err))
			AbortUnexpectedError(c)
			return
		}

		c.JSON(http.StatusOK, results)
	}

	router.GET("/places/search", handler)
}
