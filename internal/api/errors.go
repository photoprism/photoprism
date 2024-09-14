package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GetErrors searches the error logs and returns the results as JSON.
//
//	@Summary	searches the error logs and returns the results as JSON
//	@Id			GetErrors
//	@Tags		Errors
//	@Produce	json
//	@Success	200				{object}	entity.Error
//	@Failure	401,403,429,400	{object}	i18n.Response
//	@Param		count			query		int		true	"maximum number of results"	minimum(1)	maximum(100000)
//	@Param		offset			query		int		false	"search result offset"		minimum(0)	maximum(100000)
//	@Param		q				query		string	false	"search query"
//	@Router		/api/v1/errors [get]
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
//	@Summary	removes all entries from the error logs
//	@Id			DeleteErrors
//	@Tags		Errors
//	@Produce	json
//	@Failure	401,403,429,500	{object}	i18n.Response
//	@Router		/api/v1/errors [delete]
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
