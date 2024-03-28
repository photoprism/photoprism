package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// FindUserSessions finds user sessions and returns them as JSON.
//
// GET /api/v1/users/:uid/sessions
func FindUserSessions(router *gin.RouterGroup) {
	router.GET("/users/:uid/sessions", func(c *gin.Context) {
		// Check if the session user is has user management privileges.
		s := Auth(c, acl.ResourceSessions, acl.ActionManageOwn)

		if s.Abort(c) {
			return
		}

		// Get global config.
		conf := get.Config()

		// Check feature flags and authorization.
		if conf.Public() || conf.Demo() || !s.HasRegisteredUser() || conf.DisableSettings() {
			c.JSON(http.StatusNotFound, entity.Users{})
			return
		} else if !rnd.IsUID(s.UserUID, entity.UserUID) || s.UserUID != c.Param("uid") {
			c.JSON(http.StatusForbidden, entity.Users{})
			return
		}

		var f form.SearchSessions

		// Init search form.
		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		// Filter by user.
		f.UID = s.UserUID

		// Perform search.
		result, err := search.Sessions(f)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)

		c.JSON(http.StatusOK, result)
	})
}
