package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GetMemoriesOnThisDay returns years with photos taken on today's month and day as JSON.
//
//	@Summary	returns years with photos taken on today's month and day as JSON
//	@Id			GetMemoriesOnThisDay
//	@Tags		Memories
//	@Produce	json
//	@Success	200				{object}	[]query.MemoryYear
//	@Failure	401,403,429,500	{object}	i18n.Response
//	@Router		/api/v1/memories/on-this-day [get]
func GetMemoriesOnThisDay(router *gin.RouterGroup) {
	router.GET("/memories/on-this-day", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionSearch)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		result, err := query.MemoriesOnThisDay(conf.Settings().Features.Private)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		c.JSON(http.StatusOK, result)
	})
}
