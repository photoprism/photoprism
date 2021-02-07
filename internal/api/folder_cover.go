package api

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	folderCover = "folder-cover"
)

// GET /api/v1/folders/t/:hash/:token/:type
//
// Parameters:
//   uid: string folder uid
//   token: string url security token, see config
//   type: string thumb type, see thumb.Types
func GetFolderCover(router *gin.RouterGroup) {
	router.GET("/folders/t/:uid/:token/:type", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", folderIconSvg)
			return
		}

		start := time.Now()
		conf := service.Config()
		uid := c.Param("uid")
		typeName := c.Param("type")
		download := c.Query("download") != ""

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("folder: invalid thumb type %s", txt.Quote(typeName))
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		}

		if thumbType.ExceedsSize() && !conf.ThumbUncached() {
			typeName, thumbType = thumb.Find(conf.ThumbSize())

			if typeName == "" {
				log.Errorf("folder: invalid thumb size %d", conf.ThumbSize())
				c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
				return
			}
		}

		cache := service.CoverCache()
		cacheKey := CacheKey(folderCover, uid, typeName)

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Debugf("api: cache hit for %s [%s]", cacheKey, time.Since(start))

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
			log.Debugf("%s: no photos yet, using generic image for %s", folderCover, uid)
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("%s: could not find original for %s", folderCover, fileName)
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			log.Warnf("%s: %s is missing", folderCover, txt.Quote(f.FileName))
			logError(folderCover, f.Update("FileMissing", true))
			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsSizeUncached() && !download {
			log.Debugf("%s: using original, size exceeds limit (width %d, height %d)", folderCover, thumbType.Width, thumbType.Height)
			AddCoverCacheHeader(c)
			c.File(fileName)
			return
		}

		var thumbnail string

		if conf.ThumbUncached() || thumbType.OnDemand() {
			thumbnail, err = thumb.FromFile(fileName, f.FileHash, conf.ThumbPath(), thumbType.Width, thumbType.Height, thumbType.Options...)
		} else {
			thumbnail, err = thumb.FromCache(fileName, f.FileHash, conf.ThumbPath(), thumbType.Width, thumbType.Height, thumbType.Options...)
		}

		if err != nil {
			log.Errorf("%s: %s", folderCover, err)
			c.Data(http.StatusOK, "image/svg+xml", folderIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("%s: %s has empty thumb name - bug?", folderCover, filepath.Base(fileName))
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
