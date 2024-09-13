package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// SearchPhotos finds pictures and returns them as JSON.
//
//	@Summary		finds pictures and returns them as JSON
//	@Description	Fore more information see:
//	@Description	- https://docs.photoprism.app/developer-guide/api/search/#get-apiv1photos
//	@Id				SearchPhotos
//	@Tags			Photos
//	@Produce		json
//	@Success		200				{object}	search.PhotoResults
//	@Failure		400,401,403,404	{object}	i18n.Response
//	@Param			count			query		int		true	"maximum number of files"	minimum(1)	maximum(100000)
//	@Param			offset			query		int		false	"file offset"				minimum(0)	maximum(100000)
//	@Param			order			query		string	false	"sort order"				Enums(favorites, name, title, added, edited)
//	@Param			merged			query		bool	false	"groups consecutive files that belong to the same photo"
//	@Param			public			query		bool	false	"excludes private pictures"
//	@Param			quality			query		int		true	"minimum quality score (1-7)"	Enums(0, 1, 2, 3, 4, 5, 6, 7)
//	@Param			q				query		string	false	"search query"
//	@Param			s				query		string	false	"album uid"
//	@Param			path			query		string	false	"photo path"
//	@Param			video			query		bool	false	"is type video"
//	@Router			/api/v1/photos [get]
func SearchPhotos(router *gin.RouterGroup) {
	// searchPhotos checking authorization and parses the search request.
	searchForm := func(c *gin.Context) (f form.SearchPhotos, s *entity.Session, err error) {
		s = AuthAny(c, acl.ResourcePhotos, acl.Permissions{acl.ActionSearch, acl.ActionView, acl.AccessShared})

		// Abort if permission was not granted.
		if s.Abort(c) {
			return f, s, i18n.Error(i18n.ErrForbidden)
		}

		// Abort if request params are invalid.
		if err = c.MustBindWith(&f, binding.Form); err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourcePhotos), "form invalid", "%s"}, s.RefID, err)
			AbortBadRequest(c)
			return f, s, err
		}

		settings := get.Config().Settings()

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

		return f, s, nil
	}

	// defaultHandler a standard JSON result with all fields.
	defaultHandler := func(c *gin.Context) {
		f, s, err := searchForm(c)

		// Abort if authorization or form are invalid.
		if err != nil {
			return
		}

		// Find matching pictures.
		result, count, err := search.UserPhotos(f, s)

		// Ok?
		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourcePhotos), "search", "%s"}, s.RefID, err)
			AbortBadRequest(c)
			return
		}

		// Add response headers.
		AddCountHeader(c, count)
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c, s)

		// Return as JSON.
		c.JSON(http.StatusOK, result)
	}

	// viewHandler returns a photo viewer formatted result.
	viewHandler := func(c *gin.Context) {
		f, s, err := searchForm(c)

		// Abort if authorization or form are invalid.
		if err != nil {
			return
		}

		conf := get.Config()

		result, count, err := search.UserPhotosViewerResults(f, s, conf.ContentUri(), conf.ApiUri(), s.PreviewToken, s.DownloadToken)

		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourcePhotos), "view", "%s"}, s.RefID, err)
			AbortBadRequest(c)
			return
		}

		// Add response headers.
		AddCountHeader(c, count)
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c, s)

		// Return as JSON.
		c.JSON(http.StatusOK, result)
	}

	// Register route handlers.
	router.GET("/photos", defaultHandler)
	router.GET("/photos/view", viewHandler)
}
