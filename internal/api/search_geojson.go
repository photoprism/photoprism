package api

import (
	"net/http"

	"github.com/photoprism/photoprism/pkg/sanitize"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SearchGeo finds photos and returns results as JSON, so they can be displayed on a map or in a viewer.
//
// GET /api/v1/geo
func SearchGeo(router *gin.RouterGroup) {
	handler := func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.SearchGeo

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
		photos, err := search.Geo(f)

		if err != nil {
			log.Warnf("search: %s", err)
			AbortBadRequest(c)
			return
		}

		var resp []byte

		// Render JSON response.
		switch sanitize.Token(c.Param("format")) {
		case "view":
			conf := service.Config()
			resp, err = photos.ViewerJSON(conf.ContentUri(), conf.ApiUri(), conf.PreviewToken(), conf.DownloadToken())
		default:
			resp, err = photos.GeoJSON()
		}

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		AddTokenHeaders(c)

		c.Data(http.StatusOK, "application/json", resp)
	}

	router.GET("/geo", handler)
	router.GET("/geo/:format", handler)
}
