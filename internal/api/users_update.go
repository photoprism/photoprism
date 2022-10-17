package api

import (
	"net/http"

	"github.com/photoprism/photoprism/pkg/clean"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
)

// UpdateUser updates the profile information of the currently authenticated user.
//
// PUT /api/v1/users/:uid
func UpdateUser(router *gin.RouterGroup) {
	router.PUT("/users/:uid", func(c *gin.Context) {
		conf := get.Config()

		if conf.Demo() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

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

		// 1) Init form with model values
		f, err := form.NewUser(m)

		if err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// 2) Update form with values from request
		if err = c.BindJSON(&f); err != nil {
			log.Error(err)
			AbortBadRequest(c)
			return
		}

		// 3) Save model with values from form
		if err = m.SaveForm(f); err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// Clear the session cache, as it contains user information.
		s.ClearCache()

		event.SuccessMsg(i18n.MsgChangesSaved)

		m = entity.FindUserByUID(uid)

		if m == nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, m)
	})
}
