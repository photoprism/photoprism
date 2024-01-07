package api

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	folderCover = "folder-cover"
)

// FolderCover returns a folder cover image.
//
// Request Parameters:
// - uid: string folder uid
// - token: string url security token, see config
// - size: string thumb type, see thumb.Sizes
//
// GET /api/v1/folders/t/:hash/:token/:size
func FolderCover(router *gin.RouterGroup) {
	router.GET("/folders/t/:uid/:token/:size", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", folderIconSvg)
			return
		}

		start := time.Now()
		conf := get.Config()
		uid := c.Param("uid")
		thumbName := thumb.Name(clean.Token(c.Param("size")))
		download := c.Query("download") != ""

		size, ok := thumb.Sizes[thumbName]

		if !ok {
			log.Errorf("%s: invalid size %s", folderCover, thumbName)
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		}

		if size.Uncached() && !conf.ThumbUncached() {
			thumbName, size = thumb.Find(conf.ThumbSizePrecached())

			if thumbName == "" {
				log.Errorf("folder: invalid thumb size %d", conf.ThumbSizePrecached())
				c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
				return
			}
		}

		cache := get.CoverCache()
		cacheKey := CacheKey(folderCover, uid, string(thumbName))

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Tracef("api-v1: cache hit for %s [%s]", cacheKey, time.Since(start))

			cached := cacheData.(ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("%s: %s not found", folderCover, uid)
				c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
				return
			}

			AddCoverCacheHeader(c)

			if download {
				c.FileAttachment(cached.FileName, cached.ShareName)
			} else {
				c.File(cached.FileName)
			}

			return
		}

		f, err := query.FolderCoverByUID(uid)

		if err != nil {
			log.Debugf("%s: %s contains no pictures, using generic cover", folderCover, uid)
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("%s: could not find original for %s", folderCover, fileName)
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			log.Warnf("%s: %s is missing", folderCover, clean.Log(f.FileName))
			logError(folderCover, f.Update("FileMissing", true))
			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if size.ExceedsLimit() && !download {
			log.Debugf("%s: using original, size exceeds limit (width %d, height %d)", folderCover, size.Width, size.Height)
			AddCoverCacheHeader(c)
			c.File(fileName)
			return
		}

		var thumbnail string

		if conf.ThumbUncached() || size.Uncached() {
			thumbnail, err = thumb.FromFile(fileName, f.FileHash, conf.ThumbCachePath(), size.Width, size.Height, f.FileOrientation, size.Options...)
		} else {
			thumbnail, err = thumb.FromCache(fileName, f.FileHash, conf.ThumbCachePath(), size.Width, size.Height, size.Options...)
		}

		if err != nil {
			log.Errorf("%s: %s", folderCover, err)
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("%s: %s has empty thumb name - you may have found a bug", folderCover, filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		}

		cache.SetDefault(cacheKey, ThumbCache{thumbnail, f.ShareBase(0)})
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		AddCoverCacheHeader(c)

		if download {
			c.FileAttachment(thumbnail, f.DownloadName(DownloadName(c), 0))
		} else {
			c.File(thumbnail)
		}
	})
}
