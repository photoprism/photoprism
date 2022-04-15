package api

import (
	"net/http"

	"github.com/photoprism/photoprism/pkg/clean"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/service"
)

// PUT /api/v1/users/:uid/password
func ChangePassword(router *gin.RouterGroup) {
	router.PUT("/users/:uid/password", func(c *gin.Context) {
		conf := service.Config()

		if conf.Public() || conf.DisableSettings() {
			Abort(c, http.StatusForbidden, i18n.ErrPublic)
			return
		}

		s := Auth(SessionID(c), acl.ResourceUsers, acl.ActionUpdateSelf)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		uid := clean.IdString(c.Param("uid"))
		m := entity.FindUserByUID(uid)

		if s.User.UserUID != m.UserUID {
			AbortUnauthorized(c)
			return
		}

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrUserNotFound)
			return
		}

		f := form.ChangePassword{}

		if err := c.BindJSON(&f); err != nil {
			Error(c, http.StatusBadRequest, err, i18n.ErrInvalidPassword)
			return
		}

		if m.InvalidPassword(f.OldPassword) {
			Abort(c, http.StatusBadRequest, i18n.ErrInvalidPassword)
			return
		}

		if err := m.SetPassword(f.NewPassword); err != nil {
			Error(c, http.StatusBadRequest, err, i18n.ErrInvalidPassword)
			return
		}

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgPasswordChanged))
	})
}
