package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// FindUserSessions finds user sessions and returns them as JSON.
//
//	@Summary	list sessions for a user
//	@Id			FindUserSessions
//	@Tags		Users, Authentication
//	@Produce	json
//	@Param		uid			path		string	true	"user uid"
//	@Param		count		query		int		true	"maximum number of results"	minimum(1)	maximum(100000)
//	@Param		offset		query		int		false	"result offset"				minimum(0)
//	@Param		q			query		string	false	"filter by username or client name"
//	@Success	200			{object}	entity.Sessions
//	@Failure	401,403,429	{object}	i18n.Response
//	@Router		/api/v1/users/{uid}/sessions [get]
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

		// Init search request form.
		var frm form.SearchSessions
		err := c.MustBindWith(&frm, binding.Form)

		// Abort if invalid.
		if err != nil {
			AbortBadRequest(c, err)
			return
		}

		// Find applications that belong to the current user and sort them by name.
		frm.UID = s.UserUID
		frm.Order = sortby.ClientName
		frm.Provider = authn.ProviderApplication.String()
		frm.Method = authn.MethodDefault.String()

		// Perform search.
		result, err := search.Sessions(frm)

		if err != nil {
			AbortBadRequest(c, err)
			return
		}

		AddLimitHeader(c, frm.Count)
		AddOffsetHeader(c, frm.Offset)

		c.JSON(http.StatusOK, result)
	})
}
