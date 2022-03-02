package api

import (
	"net/http"

	"github.com/photoprism/photoprism/internal/service"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"

	"github.com/gin-gonic/gin"
)

// SendFeedback sends a feedback message.
//
// POST /api/v1/feedback
func SendFeedback(router *gin.RouterGroup) {
	router.POST("/feedback", func(c *gin.Context) {
		conf := service.Config()

		if conf.Public() {
			Abort(c, http.StatusForbidden, i18n.ErrPublic)
			return
		}

		s := Auth(SessionID(c), acl.ResourceFeedback, acl.ActionCreate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf.UpdateHub()

		var f form.Feedback

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
