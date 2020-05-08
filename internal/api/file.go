package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
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

// POST /api/v1/files/:uuid/link
//
// Parameters:
//   uuid: string SHA-1 hash of the file
func LinkFile(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/files/:uuid/link", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		m, err := query.FileByUUID(c.Param("uuid"))

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrFileNotFound)
			return
		}

		if link, err := newLink(c); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		} else {
			entity.Db().Model(&m).Association("Links").Append(link)
		}

		event.Success("created file share link")

		c.JSON(http.StatusOK, m)
	})
}
