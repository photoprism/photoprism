package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SearchGeo finds photos and returns results as JSON, so they can be displayed on a map or in a viewer.
// See form.SearchPhotosGeo for supported search params and data types.
//
// GET /api/v1/geo
func SearchGeo(router *gin.RouterGroup) {
	handler := func(c *gin.Context) {
		s := AuthAny(c, acl.ResourcePlaces, acl.Permissions{acl.ActionSearch, acl.ActionView, acl.AccessShared})

		// Abort if permission was not granted.
		if s.Abort(c) {
			return
		}

		var f form.SearchPhotosGeo
		var err error

		// Abort if request params are invalid.
		if err = c.MustBindWith(&f, binding.Form); err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourcePlaces), "form invalid", "%s"}, s.RefID, err)
			AbortBadRequest(c)
			return
		}

		conf := service.Config()

		// Ignore private flag if feature is disabled.
		if !conf.Settings().Features.Private {
			f.Public = false
		}

		// Find matching pictures.
		photos, err := search.UserPhotosGeo(f, s)

		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourcePlaces), "search", "%s"}, s.RefID, err)
			AbortBadRequest(c)
			return
		}

		// Add response headers.
		AddCountHeader(c, len(photos))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c)

		var resp []byte

		// Render JSON response.
		switch clean.Token(c.Param("format")) {
		case "view":
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
