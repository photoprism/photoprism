package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// SendFeedback sends a feedback message.
//
// POST /api/v1/feedback
func SendFeedback(router *gin.RouterGroup) {
	router.POST("/feedback", func(c *gin.Context) {
		conf := get.Config()

		if conf.Public() {
			Abort(c, http.StatusForbidden, i18n.ErrPublic)
			return
		}

		s := Auth(c, acl.ResourceFeedback, acl.ActionCreate)

		// Abort if permission was not granted.
		if s.Abort(c) {
			return
		}

		conf.RenewApiKeys()

		var f form.Feedback

		// Assign and validate request form values.
		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		if f.Empty() {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		if err := conf.Hub().SendFeedback(f); err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK})
	})
}
