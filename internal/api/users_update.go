package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/pkg/clean"
)

// UpdateUser updates the profile information of the currently authenticated user.
//
// PUT /api/v1/users/:uid
func UpdateUser(router *gin.RouterGroup) {
	router.PUT("/users/:uid", func(c *gin.Context) {
		conf := get.Config()

		if conf.Public() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		// Check if the session user is allowed to manage all accounts or update his/her own account.
		s := AuthAny(c, acl.ResourceUsers, acl.Permissions{acl.ActionManage, acl.AccessOwn, acl.ActionUpdate})

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		m := entity.FindUserByUID(uid)

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrUserNotFound)
			return
		}

		// Init form with model values.
		f, err := m.Form()

		if err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// Update form with values from request.
		if err = c.BindJSON(&f); err != nil {
			log.Error(err)
			AbortBadRequest(c)
			return
		}

		// Check if the session user is has user management privileges.
		isPrivileged := acl.Resources.AllowAll(acl.ResourceUsers, s.User().AclRole(), acl.Permissions{acl.AccessAll, acl.ActionManage})

		// Prevent super admins from locking themselves out.
		if u := s.User(); u.IsSuperAdmin() && u.Equal(m) && !f.CanLogin {
			f.CanLogin = true
		}

		// Save model with values from form.
		if err = m.SaveForm(f, isPrivileged); err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// Clear the session cache, as it contains user information.
		s.ClearCache()

		m = entity.FindUserByUID(uid)

		if m == nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, m)
	})
}
