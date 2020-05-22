package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// GetFolders is a reusable request handler for directory listings (GET /api/v1/folders/*).
func GetFolders(router *gin.RouterGroup, conf *config.Config, root, pathName string)  {
	router.GET("/folders/" + root, func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		recursive := c.Query("recursive") != ""

		folders, err := entity.Folders(root, pathName, recursive)

		if err != nil {
			log.Errorf("folder: %s", err)
		}

		c.JSON(http.StatusOK, folders)
	})
}

// GET /api/v1/folders/originals
func GetFoldersOriginals(router *gin.RouterGroup, conf *config.Config) {
	GetFolders(router, conf, entity.FolderRootOriginals, conf.OriginalsPath())
}

// GET /api/v1/folders/import
func GetFoldersImport(router *gin.RouterGroup, conf *config.Config) {
	GetFolders(router, conf, entity.FolderRootImport, conf.ImportPath())
}
