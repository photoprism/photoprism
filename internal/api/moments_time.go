package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GetMomentsTime returns monthly albums as JSON.
//
// GET /api/v1/moments/time
func GetMomentsTime(router *gin.RouterGroup) {
	router.GET("/moments/time", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCalendar, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		conf := service.Config()

		result, err := query.MomentsTime(1, conf.Settings().Features.Private)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		c.JSON(http.StatusOK, result)
	})
}
