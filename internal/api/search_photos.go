package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/internal/service"
)

// SearchPhotos searches the pictures index and returns the result as JSON.
//
// GET /api/v1/photos
//
// See form.SearchPhotos for supported search params and data types.
func SearchPhotos(router *gin.RouterGroup) {
	// searchPhotos checking authorization and parses the search request.
	searchForm := func(c *gin.Context) (f form.SearchPhotos, err error) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return f, i18n.Error(i18n.ErrUnauthorized)
		}

		err = c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return f, err
		}

		// Guests may only see public content in shared albums.
		if s.Guest() {
			if f.Album == "" || !s.HasShare(f.Album) {
				AbortUnauthorized(c)
				return f, i18n.Error(i18n.ErrUnauthorized)
			}

			f.UID = ""
			f.Albums = ""
			f.Public = true
			f.Private = false
			f.Hidden = false
			f.Archived = false
			f.Review = false
		}

		return f, nil
	}

	// defaultHandler a standard JSON result with all fields.
	defaultHandler := func(c *gin.Context) {
		f, err := searchForm(c)

		// Abort if authorization or form are invalid.
		if err != nil {
			return
		}

		result, count, err := search.Photos(f)

		if err != nil {
			log.Warnf("search: %s", err)
			AbortBadRequest(c)
			return
		}

		// Add response headers.
		AddCountHeader(c, count)
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c)

		// Render as JSON.
		c.JSON(http.StatusOK, result)
	}

	// viewHandler returns a photo viewer formatted result.
	viewHandler := func(c *gin.Context) {
		f, err := searchForm(c)

		// Abort if authorization or form are invalid.
		if err != nil {
			return
		}

		conf := service.Config()
		result, count, err := search.PhotosViewerResults(f, conf.ContentUri(), conf.ApiUri(), conf.PreviewToken(), conf.DownloadToken())

		if err != nil {
			log.Warnf("search: %s", err)
			AbortBadRequest(c)
			return
		}

		// Add response headers.
		AddCountHeader(c, count)
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c)

		// Render as JSON.
		c.JSON(http.StatusOK, result)
	}

	// Register route handlers.
	router.GET("/photos", defaultHandler)
	router.GET("/photos/view", viewHandler)
}
