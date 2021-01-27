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

// GET /api/v1/t/:hash/:token/:type
//
// Parameters:
//   hash: string sha1 file hash
//   token: string url security token, see config
//   type: string thumb type, see thumb.Types
func GetThumb(router *gin.RouterGroup) {
	router.GET("/t/:hash/:token/:type", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		start := time.Now()
		conf := service.Config()
		fileHash := c.Param("hash")
		typeName := c.Param("type")
		download := c.Query("download") != ""

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("thumbs: invalid type %s", txt.Quote(typeName))
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		if thumbType.ExceedsSize() && !conf.ThumbUncached() {
			typeName, thumbType = thumb.Find(conf.ThumbSize())

			if typeName == "" {
				log.Errorf("thumbs: invalid size %d", conf.ThumbSize())
				c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
				return
			}
		}

		cache := service.ThumbCache()
		cacheKey := CacheKey("thumbs", fileHash, typeName)

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))

			cached := cacheData.(ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("thumbs: %s not found", fileHash)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			}

			AddThumbCacheHeader(c)

			if c.Query("download") != "" {
				c.FileAttachment(cached.FileName, cached.ShareName)
			} else {
				c.File(cached.FileName)
			}

			return
		}

		// Return existing thumbs straight away.
		if !download {
			if fileName, err := thumb.Filename(fileHash, conf.ThumbPath(), thumbType.Width, thumbType.Height, thumbType.Options...); err == nil && fs.FileExists(fileName) {
				c.File(fileName)
				return
			}
		}

		// Query index for file infos.
		f, err := query.FileByHash(fileHash)

		if err != nil {
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		// Find fallback if file is not a JPEG image.
		if f.NoJPEG() {
			f, err = query.FileByPhotoUID(f.PhotoUID)

			if err != nil {
				c.Data(http.StatusOK, "image/svg+xml", fileIconSvg)
				return
			}
		}

		// Return SVG icon as placeholder if file has errors.
		if f.FileError != "" {
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("thumbs: file %s is missing", txt.Quote(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logError("thumbs", f.Update("FileMissing", true))

			if f.AllFilesMissing() {
				log.Infof("thumbs: deleting photo, all files missing for %s", txt.Quote(f.FileName))

				logError("thumbs", f.RelatedPhoto().Delete(false))
			}

			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsSizeUncached() && c.Query("download") == "" {
			log.Debugf("thumbs: using original, size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)

			AddThumbCacheHeader(c)
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
			log.Errorf("thumbs: %s", err)
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("thumbs: %s has empty thumb name - bug?", filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		}

		cache.SetDefault(cacheKey, ThumbCache{thumbnail, f.ShareBase(0)})
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		AddThumbCacheHeader(c)

		if download {
			c.FileAttachment(thumbnail, f.DownloadName(DownloadName(c), 0))
		} else {
			c.File(thumbnail)
		}
	})
}
