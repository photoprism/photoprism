package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

func GetErrors(router *gin.RouterGroup) {
	router.GET("/errors", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceLogs, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		limit := txt.Int(c.Query("count"))
		offset := txt.Int(c.Query("offset"))

		if resp, err := query.Errors(limit, offset, c.Query("q")); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		} else {
			AddCountHeader(c, len(resp))
			AddLimitHeader(c, limit)
			AddOffsetHeader(c, offset)

			c.JSON(http.StatusOK, resp)
		}
	})
}
