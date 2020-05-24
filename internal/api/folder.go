package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
)

type FoldersResponse struct {
	Root      string          `json:"root,omitempty"`
	Folders   []entity.Folder `json:"folders"`
	Files     []entity.File   `json:"files,omitempty"`
	Recursive bool            `json:"recursive,omitempty"`
}

// GetFolders is a reusable request handler for directory listings (GET /api/v1/folders/*).
func GetFolders(router *gin.RouterGroup, conf *config.Config, root, rootPath string) {
	handler := func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		start := time.Now()
		gc := service.Cache()
		recursive := c.Query("recursive") != ""
		listFiles := c.Query("files") != ""
		resp := FoldersResponse{Root: root, Recursive: recursive}
		path := c.Param("path")

		cacheKey := fmt.Sprintf("folders:%s:%t:%t", filepath.Join(rootPath, path), recursive, listFiles)

		if cacheData, ok := gc.Get(cacheKey); ok {
			log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))
			c.JSON(http.StatusOK, cacheData.(FoldersResponse))
			return
		}

		if folders, err := query.FoldersByPath(root, rootPath, path, recursive); err != nil {
			log.Errorf("folders: %s", err)
			c.JSON(http.StatusOK, resp)
			return
		} else {
			resp.Folders = folders
		}

		if listFiles {
			if files, err := query.FilesByPath(root, path); err != nil {
				log.Errorf("folders: %s", err)
			} else {
				resp.Files = files
			}
		}

		gc.Set(cacheKey, resp, time.Minute*5)
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		c.Header("X-Count", strconv.Itoa(len(resp.Files) + len(resp.Folders)))
		c.Header("X-Offset", "0")

		c.JSON(http.StatusOK, resp)
	}

	router.GET("/folders/"+root, handler)
	router.GET("/folders/"+root+"/*path", handler)
}

// GET /api/v1/folders/originals
func GetFoldersOriginals(router *gin.RouterGroup, conf *config.Config) {
	GetFolders(router, conf, entity.RootOriginals, conf.OriginalsPath())
}

// GET /api/v1/folders/import
func GetFoldersImport(router *gin.RouterGroup, conf *config.Config) {
	GetFolders(router, conf, entity.RootImport, conf.ImportPath())
}
