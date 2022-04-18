package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/internal/service"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SearchGeo finds photos and returns results as JSON, so they can be displayed on a map or in a viewer.
//
// GET /api/v1/geo
//
// See form.SearchPhotosGeo for supported search params and data types.
func SearchGeo(router *gin.RouterGroup) {
	handler := func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.SearchPhotosGeo

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		// Guests may only see public content.
		if s.Guest() {
			if f.Album == "" || !s.HasShare(f.Album) {
				AbortUnauthorized(c)
				return
			}

			f.Public = true
			f.Private = false
			f.Archived = false
			f.Review = false
		}

		// Find matching pictures.
		photos, err := search.PhotosGeo(f)

		if err != nil {
			log.Warnf("search: %s", err)
			AbortBadRequest(c)
			return
		}

		// Add response headers.
		AddTokenHeaders(c)

		var resp []byte

		// Render JSON response.
		switch clean.Token(c.Param("format")) {
		case "view":
			conf := service.Config()
			resp, err = photos.ViewerJSON(conf.ContentUri(), conf.ApiUri(), conf.PreviewToken(), conf.DownloadToken())
		default:
			resp, err = photos.GeoJSON()
		}

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		c.Data(http.StatusOK, "application/json", resp)
	}

	// Register route handlers.
	router.GET("/geo", handler)
	router.GET("/geo/:format", handler)
}
