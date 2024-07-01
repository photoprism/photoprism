package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SearchAlbums finds albums and returns them as JSON.
//
// GET /api/v1/albums
func SearchAlbums(router *gin.RouterGroup) {
	router.GET("/albums", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourceAlbums, acl.Permissions{acl.ActionSearch, acl.ActionView, acl.AccessShared})

		// Abort if permission was not granted.
		if s.Abort(c) {
			return
		}

		var err error
		var f form.SearchAlbums

		// Abort if request params are invalid.
		if err = c.MustBindWith(&f, binding.Form); err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "albums", "search", "form invalid", "%s"}, s.RefID, err)
			AbortBadRequest(c)
			return
		}

		settings := get.Config().Settings()

		// Ignore private flag if feature is disabled.
		if !settings.Features.Private {
			f.Public = false
		}

		// Find matching albums.
		result, err := search.UserAlbums(f, s)

		// Ok?
		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "albums", "search", "%s"}, s.RefID, err)
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		AddCountHeader(c, len(result))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c, s)

		// Return as JSON.
		c.JSON(http.StatusOK, result)
	})
}
