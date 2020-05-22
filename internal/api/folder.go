package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service"
)

// GetFolders is a reusable request handler for directory listings (GET /api/v1/folders/*).
func GetFolders(router *gin.RouterGroup, conf *config.Config, root, pathName string)  {
	router.GET("/folders/" + root, func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()
		gc := service.Cache()
		recursive := c.Query("recursive") != ""

		cacheKey := fmt.Sprintf("folders:%s:%t", pathName, recursive)

		if cacheData, ok := gc.Get(cacheKey); ok {
			log.Debugf("folders: %s cache hit [%s]", cacheKey, time.Since(start))
			c.JSON(http.StatusOK, cacheData.([]entity.Folder))
			return
		}

		folders, err := entity.Folders(root, pathName, recursive)

		if err != nil {
			log.Errorf("folders: %s", err)
		} else {
			gc.Set(cacheKey, folders, time.Minute*5)

			log.Debugf("folders: %s cached [%s]", cacheKey, time.Since(start))
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
