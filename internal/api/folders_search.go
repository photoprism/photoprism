package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

// FoldersResponse represents the folders API response.
type FoldersResponse struct {
	Root      string          `json:"root,omitempty"`
	Folders   []entity.Folder `json:"folders"`
	Files     []entity.File   `json:"files,omitempty"`
	Recursive bool            `json:"recursive,omitempty"`
	Cached    bool            `json:"cached,omitempty"`
}

// SearchFoldersOriginals returns folders in originals as JSON.
//
// GET /api/v1/folders/originals
func SearchFoldersOriginals(router *gin.RouterGroup) {
	conf := service.Config()
	SearchFolders(router, "originals", entity.RootOriginals, conf.OriginalsPath())
}

// SearchFoldersImport returns import folders as JSON.
//
// GET /api/v1/folders/import
func SearchFoldersImport(router *gin.RouterGroup) {
	conf := service.Config()
	SearchFolders(router, "import", entity.RootImport, conf.ImportPath())
}

// SearchFolders is a reusable request handler for directory listings (GET /api/v1/folders/*).
func SearchFolders(router *gin.RouterGroup, urlPath, rootName, rootPath string) {
	handler := func(c *gin.Context) {
		s := Auth(c, acl.ResourceFolders, acl.ActionSearch)

		// Abort if permission was not granted.
		if s.Abort(c) {
			return
		}

		var f form.SearchFolders

		start := time.Now()
		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
			return
		}

		cache := service.FolderCache()
		recursive := f.Recursive
		listFiles := f.Files
		uncached := listFiles || f.Uncached
		resp := FoldersResponse{Root: rootName, Recursive: recursive, Cached: !uncached}
		path := clean.Path(c.Param("path"))

		cacheKey := fmt.Sprintf("folder:%s:%t:%t", filepath.Join(rootName, path), recursive, listFiles)

		if !uncached {
			if cacheData, ok := cache.Get(cacheKey); ok {
				cached := cacheData.(FoldersResponse)

				log.Tracef("api-v1: cache hit for %s [%s]", cacheKey, time.Since(start))

				c.JSON(http.StatusOK, cached)
				return
			}
		}

		if folders, err := query.FoldersByPath(rootName, rootPath, path, recursive); err != nil {
			log.Errorf("folder: %s", err)
			c.JSON(http.StatusOK, resp)
			return
		} else {
			resp.Folders = folders
		}

		if listFiles {
			if files, err := query.FilesByPath(f.Count, f.Offset, rootName, path); err != nil {
				log.Errorf("folder: %s", err)
			} else {
				resp.Files = files
			}
		}

		if !uncached {
			cache.SetDefault(cacheKey, resp)
			log.Debugf("cached %s [%s]", cacheKey, time.Since(start))
		}

		AddFileCountHeaders(c, len(resp.Files), len(resp.Folders))
		AddCountHeader(c, len(resp.Files)+len(resp.Folders))
		AddLimitHeader(c, f.Count)
		AddOffsetHeader(c, f.Offset)
		AddTokenHeaders(c, s)

		c.JSON(http.StatusOK, resp)
	}

	router.GET("/folders/"+urlPath, handler)
	router.GET("/folders/"+urlPath+"/*path", handler)
}
