package api

import (
	"net/http"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gin-gonic/gin"
)

// GET /api/v1/moments/time
func GetMomentsTime(router *gin.RouterGroup) {
	router.GET("/moments/time", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceAlbums, acl.ActionExport)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		result, err := query.MomentsTime(1)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UpperFirst(err.Error())})
			return
		}

		c.JSON(http.StatusOK, result)
	})
}
