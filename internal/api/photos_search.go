package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
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
	searchForm := func(c *gin.Context) (f form.SearchPhotos, s *entity.Session, err error) {
		s = AuthAny(c, acl.ResourcePhotos, acl.Permissions{acl.ActionSearch, acl.ActionView, acl.AccessShared})

		// Abort if permission was not granted.
		if s.Abort(c) {
			return f, s, i18n.Error(i18n.ErrForbidden)
		}

		// Abort if request params are invalid.
		if err = c.MustBindWith(&f, binding.Form); err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "photos", "form invalid", "%s"}, s.RefID, err)
			AbortBadRequest(c)
			return f, s, err
		}

		// Limit results to a specific album?
		if f.Album == "" {
			if acl.Resources.Deny(acl.ResourcePhotos, s.User().AclRole(), acl.ActionSearch) {
				event.AuditErr([]string{ClientIP(c), "session %s", "%s %s as %s", "denied"}, s.RefID, acl.ActionSearch.String(), string(acl.ResourcePhotos), s.User().AclRole().String())
				c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
				return f, s, i18n.Error(i18n.ErrForbidden)
			}
		} else if a, err := entity.CachedAlbumByUID(f.Album); err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "photos", "album", f.Album, "not found"}, s.RefID)
			AbortAlbumNotFound(c)
			return f, s, i18n.Error(i18n.ErrAlbumNotFound)
		} else {
			f.Filter = a.AlbumFilter
		}

		// Parse query string and filter.
		if err = f.ParseQueryString(); err != nil {
			log.Debugf("search: %s", err)
			AbortBadRequest(c)
			return f, s, err
		}

		conf := service.Config()

		// Enforce ACL.
		if acl.Resources.Deny(acl.ResourcePhotos, s.User().AclRole(), acl.AccessPrivate) {
			f.Public = true
			f.Private = false
		}
		if acl.Resources.Deny(acl.ResourcePhotos, s.User().AclRole(), acl.ActionDelete) {
			f.Archived = false
			f.Review = false
		}
		if acl.Resources.Deny(acl.ResourceFiles, s.User().AclRole(), acl.ActionManage) {
			f.Hidden = false
		}

		// Sharing link visitors may only see public content in shared albums.
		if s.IsVisitor() {
			if f.Album == "" || !s.HasShare(f.Album) {
				event.AuditErr([]string{ClientIP(c), "session %s", "photos", "shared album", f.Album, "not shared"}, s.RefID)
				AbortForbidden(c)
				return f, s, i18n.Error(i18n.ErrUnauthorized)
			}

			f.UID = ""
			f.Albums = ""
		} else if !conf.Settings().Features.Private {
			f.Public = false
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

		result, count, err := search.Photos(f)

		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "photos", "search", "%s"}, s.RefID, err)
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
		f, s, err := searchForm(c)

		// Abort if authorization or form are invalid.
		if err != nil {
			return
		}

		conf := service.Config()

		result, count, err := search.PhotosViewerResults(f, conf.ContentUri(), conf.ApiUri(), conf.PreviewToken(), conf.DownloadToken())

		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "photos", "view", "%s"}, s.RefID, err)
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
