package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SearchGeo finds photos and returns results as JSON, so they can be displayed on a map or in a viewer.
// See form.SearchPhotosGeo for supported search params and data types.
//
//	@Summary	finds photos and returns results as JSON, so they can be displayed on a map or in a viewer
//	@Id			SearchGeo
//	@Tags		Photos
//	@Produce	json
//	@Success	200				{object}	search.GeoResults
//	@Failure	400,401,403,404	{object}	i18n.Response
//	@Param		count			query		int		true	"maximum number of files"	minimum(1)	maximum(100000)
//	@Param		offset			query		int		false	"file offset"				minimum(0)	maximum(100000)
//	@Param		public			query		bool	false	"excludes private pictures"
//	@Param		quality			query		int		true	"minimum quality score (1-7)"	Enums(0, 1, 2, 3, 4, 5, 6, 7)
//	@Param		q				query		string	false	"search query"
//	@Param		s				query		string	false	"album uid"
//	@Param		path			query		string	false	"photo path"
//	@Param		video			query		bool	false	"is type video"
//	@Router		/api/v1/geo [get]
func SearchGeo(router *gin.RouterGroup) {
	handler := func(c *gin.Context) {
		s := AuthAny(c, acl.ResourcePlaces, acl.Permissions{acl.ActionSearch, acl.ActionView, acl.AccessShared})

		// Abort if permission was not granted.
		if s.Abort(c) {
			return
		}

		var err error
		var f form.SearchPhotosGeo

		// Abort if request params are invalid.
		if err = c.MustBindWith(&f, binding.Form); err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourcePlaces), "form invalid", "%s"}, s.RefID, err)
			AbortBadRequest(c)
			return
		}

		conf := get.Config()
		settings := conf.Settings()

		// Ignore private flag if feature is disabled.
		if !settings.Features.Private {
			f.Public = false
		}

		// Ignore private flag if feature is disabled.
		if f.Scope == "" &&
			settings.Features.Review &&
			acl.Rules.Deny(acl.ResourcePhotos, s.UserRole(), acl.ActionManage) {
			f.Quality = 3
		}

		// Find matching pictures.
		photos, err := search.UserPhotosGeo(f, s)

		// Ok?
		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourcePlaces), "search", "%s"}, s.RefID, err)
			AbortBadRequest(c)
			return
		}

		// Add response headers.
		AddCountHeader(c, len(photos))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c, s)

		var resp []byte

		// Render JSON response.
		switch clean.Token(c.Param("format")) {
		case "view":
			resp, err = photos.ViewerJSON(conf.ContentUri(), conf.ApiUri(), s.PreviewToken, s.DownloadToken)
		default:
			resp, err = photos.GeoJSON()
		}

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		c.Data(http.StatusOK, header.ContentTypeJsonUtf8, resp)
	}

	// Register route handlers.
	router.GET("/geo", handler)
	router.GET("/geo/:format", handler)
}
