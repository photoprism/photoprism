package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/query"
)

// GET /api/v1/files/:hash
//
// Parameters:
//   hash: string SHA-1 hash of the file
func GetFile(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/files/:hash", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		p, err := query.FileByHash(c.Param("hash"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		c.JSON(http.StatusOK, p)
	})
}
