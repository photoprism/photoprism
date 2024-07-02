package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GetErrors searches the error logs and returns the results as JSON.
//
// GET /api/v1/errors
func GetErrors(router *gin.RouterGroup) {
	router.GET("/errors", func(c *gin.Context) {
		// Check authentication and authorization.
		s := Auth(c, acl.ResourceLogs, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		limit := txt.Int(c.Query("count"))
		offset := txt.Int(c.Query("offset"))

		// Find and return matching logs.
		if resp, err := query.Errors(limit, offset, c.Query("q")); err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		} else {
			AddCountHeader(c, len(resp))
			AddLimitHeader(c, limit)
			AddOffsetHeader(c, offset)

			c.JSON(http.StatusOK, resp)
		}
	})
}

// DeleteErrors removes all entries from the error logs.
//
// DELETE /api/v1/errors
func DeleteErrors(router *gin.RouterGroup) {
	router.DELETE("/errors", func(c *gin.Context) {
		conf := get.Config()

		// Disabled in public mode so that attackers cannot cover their tracks.
		if conf.Public() {
			Abort(c, http.StatusForbidden, i18n.ErrPublic)
			return
		}

		// Check authentication and authorization.
		s := Auth(c, acl.ResourceLogs, acl.ActionDelete)

		if s.Abort(c) {
			return
		}

		// Delete error logs.
		if err := query.DeleteErrors(); err != nil {
			log.Errorf("errors: %s (delete)", err)
			AbortDeleteFailed(c)
			return
		} else {
			c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgPermanentlyDeleted))
			return
		}
	})
}
