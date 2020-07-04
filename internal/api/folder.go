package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
)

type FoldersResponse struct {
	Root      string          `json:"root,omitempty"`
	Folders   []entity.Folder `json:"folders"`
	Files     []entity.File   `json:"files,omitempty"`
	Recursive bool            `json:"recursive,omitempty"`
	Cached    bool            `json:"cached,omitempty"`
}

// GetFolders is a reusable request handler for directory listings (GET /api/v1/folders/*).
func GetFolders(router *gin.RouterGroup, urlPath, rootName, rootPath string) {
	handler := func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceFolders, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.FolderSearch

		start := time.Now()
		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		cache := service.Cache()
		recursive := f.Recursive
		listFiles := f.Files
		uncached := listFiles || f.Uncached
		resp := FoldersResponse{Root: rootName, Recursive: recursive, Cached: !uncached}
		path := c.Param("path")

		cacheKey := fmt.Sprintf("folders:%s:%t:%t", filepath.Join(rootPath, path), recursive, listFiles)

		if !uncached {
			if cacheData, err := cache.Get(cacheKey); err == nil {
				var cached FoldersResponse

				if err := json.Unmarshal(cacheData, &cached); err != nil {
					log.Errorf("folders: %s", err)
				} else {
					log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))
					c.JSON(http.StatusOK, cached)
					return
				}
			}
		}

		if folders, err := query.FoldersByPath(rootName, rootPath, path, recursive); err != nil {
			log.Errorf("folders: %s", err)
			c.JSON(http.StatusOK, resp)
			return
		} else {
			resp.Folders = folders
		}

		if listFiles {
			if files, err := query.FilesByPath(f.Count, f.Offset, rootName, path); err != nil {
				log.Errorf("folders: %s", err)
			} else {
				resp.Files = files
			}
		}

		if !uncached {
			if c, err := json.Marshal(resp); err == nil {
				logError("folders", cache.Set(cacheKey, c))
				log.Debugf("cached %s [%s]", cacheKey, time.Since(start))
			}
		}

		c.Header("X-Files", strconv.Itoa(len(resp.Files)))
		c.Header("X-Folders", strconv.Itoa(len(resp.Folders)))
		c.Header("X-Count", strconv.Itoa(len(resp.Files)+len(resp.Folders)))
		c.Header("X-Limit", strconv.Itoa(f.Count))
		c.Header("X-Offset", strconv.Itoa(f.Offset))

		c.JSON(http.StatusOK, resp)
	}

	router.GET("/folders/"+urlPath, handler)
	router.GET("/folders/"+urlPath+"/*path", handler)
}

// GET /api/v1/folders/originals
func GetFoldersOriginals(router *gin.RouterGroup) {
	conf := service.Config()
	GetFolders(router, "originals", entity.RootOriginals, conf.OriginalsPath())
}

// GET /api/v1/folders/import
func GetFoldersImport(router *gin.RouterGroup) {
	conf := service.Config()
	GetFolders(router, "import", entity.RootImport, conf.ImportPath())
}
