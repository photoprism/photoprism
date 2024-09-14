package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GetMomentsTime returns monthly albums as JSON.
//
//	@Summary	returns monthly albums as JSON
//	@Id			GetMomentsTime
//	@Tags		Albums
//	@Produce	json
//	@Success	200				{object}	entity.Album
//	@Failure	401,403,429,500	{object}	i18n.Response
//	@Router		/api/v1/moments/time [get]
func GetMomentsTime(router *gin.RouterGroup) {
	router.GET("/moments/time", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCalendar, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		result, err := query.MomentsTime(1, conf.Settings().Features.Private)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		c.JSON(http.StatusOK, result)
	})
}
