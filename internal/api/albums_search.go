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

		var f form.SearchAlbums

		err := c.MustBindWith(&f, binding.Form)

		// Abort if request params are invalid.
		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "albums", "search", "form invalid", "%s"}, s.RefID, err)
			AbortBadRequest(c)
			return
		}

		conf := service.Config()

		// Sharing link visitors permissions are limited to shared albums.
		if s.IsVisitor() {
			f.UID = s.SharedUIDs().Join(txt.Or)
			f.Public = true
			event.AuditDebug([]string{ClientIP(c), "session %s", "albums", "search", "shared", "%s"}, s.RefID, f.UID)
		} else if conf.Settings().Features.Private {
			f.Public = true
			event.AuditDebug([]string{ClientIP(c), "session %s", "albums", "search", "all public"}, s.RefID)
		} else {
			f.Public = false
			event.AuditDebug([]string{ClientIP(c), "session %s", "albums", "search", "all public and private"}, s.RefID)
		}

		result, err := search.Albums(f)

		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "albums", "search", "%s"}, s.RefID, err)
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		AddCountHeader(c, len(result))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c)

		c.JSON(http.StatusOK, result)
	})
}
