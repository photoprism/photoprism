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

// Namespaces for caching and logs.
const (
	albumCover = "album-cover"
	labelCover = "label-cover"
)

// GET /api/v1/albums/:uid/t/:token/:type
//
// Parameters:
//   uid: string album uid
//   token: string security token (see config)
//   type: string thumb type, see photoprism.ThumbnailTypes
func AlbumCover(router *gin.RouterGroup) {
	router.GET("/albums/:uid/t/:token/:type", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", albumIconSvg)
			return
		}

		start := time.Now()
		conf := service.Config()
		typeName := c.Param("type")
		uid := c.Param("uid")

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("%s: invalid type %s", albumCover, typeName)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		cache := service.CoverCache()
		cacheKey := CacheKey(albumCover, uid, typeName)

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))

			cached := cacheData.(ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("%s: %s not found", albumCover, uid)
				c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
				return
			}

			AddCoverCacheHeader(c)

			if c.Query("download") != "" {
				c.FileAttachment(cached.FileName, cached.ShareName)
			} else {
				c.File(cached.FileName)
			}

			return
		}

		f, err := query.AlbumCoverByUID(uid)

		if err != nil {
			log.Debugf("%s: no photos yet, using generic image for %s", albumCover, uid)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("%s: could not find original for %s", albumCover, fileName)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			log.Warnf("%s: %s is missing", albumCover, txt.Quote(f.FileName))
			logError(albumCover, f.Update("FileMissing", true))
			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsSizeUncached() && c.Query("download") == "" {
			log.Debugf("%s: using original, size exceeds limit (width %d, height %d)", albumCover, thumbType.Width, thumbType.Height)
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
			log.Errorf("album: %s", err)
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("%s: %s has empty thumb name - bug?", albumCover, filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		cache.SetDefault(cacheKey, ThumbCache{thumbnail, f.ShareBase(0)})
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		AddCoverCacheHeader(c)

		if c.Query("download") != "" {
			c.FileAttachment(thumbnail, f.DownloadName(DownloadName(c), 0))
		} else {
			c.File(thumbnail)
		}
	})
}

// GET /api/v1/labels/:uid/t/:token/:type
//
// Parameters:
//   uid: string label uid
//   token: string security token (see config)
//   type: string thumb type, see photoprism.ThumbnailTypes
func LabelCover(router *gin.RouterGroup) {
	router.GET("/labels/:uid/t/:token/:type", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", labelIconSvg)
			return
		}

		start := time.Now()
		conf := service.Config()
		typeName := c.Param("type")
		uid := c.Param("uid")

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("%s: invalid type %s", labelCover, txt.Quote(typeName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		cache := service.CoverCache()
		cacheKey := CacheKey(labelCover, uid, typeName)

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))

			cached := cacheData.(ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("%s: %s not found", labelCover, uid)
				c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
				return
			}

			AddCoverCacheHeader(c)

			if c.Query("download") != "" {
				c.FileAttachment(cached.FileName, cached.ShareName)
			} else {
				c.File(cached.FileName)
			}

			return
		}

		f, err := query.LabelThumbByUID(uid)

		if err != nil {
			log.Errorf(err.Error())
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("%s: file %s is missing", labelCover, txt.Quote(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logError(labelCover, f.Update("FileMissing", true))

			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsSizeUncached() {
			log.Debugf("%s: using original, size exceeds limit (width %d, height %d)", labelCover, thumbType.Width, thumbType.Height)

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
			log.Errorf("%s: %s", labelCover, err)
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("%s: %s has empty thumb name - bug?", labelCover, filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		cache.SetDefault(cacheKey, ThumbCache{thumbnail, f.ShareBase(0)})
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		AddCoverCacheHeader(c)

		if c.Query("download") != "" {
			c.FileAttachment(thumbnail, f.DownloadName(DownloadName(c), 0))
		} else {
			c.File(thumbnail)
		}
	})
}
