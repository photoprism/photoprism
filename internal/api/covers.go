package api

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Namespaces for caching and logs.
const (
	albumCover = "album-cover"
	labelCover = "label-cover"
)

// AlbumCover returns an album cover image.
//
// The request parameters are:
//
//   - uid: string album uid
//   - token: string security token (see config)
//   - size: string thumb type, see photoprism.ThumbnailTypes
//
// GET /api/v1/albums/:uid/t/:token/:size
func AlbumCover(router *gin.RouterGroup) {
	router.GET("/albums/:uid/t/:token/:size", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", albumIconSvg)
			return
		}

		start := time.Now()
		conf := get.Config()
		thumbName := thumb.Name(clean.Token(c.Param("size")))
		uid := clean.UID(c.Param("uid"))

		size, ok := thumb.Sizes[thumbName]

		if !ok {
			log.Errorf("%s: invalid size %s", albumCover, clean.Log(thumbName.String()))
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		cache := get.CoverCache()
		cacheKey := CacheKey(albumCover, uid, string(thumbName))

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Tracef("api-v1: cache hit for %s [%s]", cacheKey, time.Since(start))

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

		f, err := query.AlbumCoverByUID(uid, conf.Settings().Features.Private)

		if err != nil {
			log.Debugf("%s: %s contains no pictures, using generic cover", albumCover, uid)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("%s: found no original for %s", albumCover, clean.Log(fileName))
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			log.Warnf("%s: %s is missing", albumCover, clean.Log(f.FileName))
			logErr(albumCover, f.Update("FileMissing", true))
			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if size.ExceedsLimit() && c.Query("download") == "" {
			log.Debugf("%s: using original, size exceeds limit (width %d, height %d)", albumCover, size.Width, size.Height)
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
			log.Errorf("%s: %s", albumCover, err)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("%s: %s has empty thumb name - you may have found a bug", albumCover, filepath.Base(fileName))
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

// LabelCover returns a label cover image.
//
// The request parameters are:
//
//   - uid: string label uid
//   - token: string security token (see config)
//   - size: string thumb type, see photoprism.ThumbnailTypes
//
// GET /api/v1/labels/:uid/t/:token/:size
func LabelCover(router *gin.RouterGroup) {
	router.GET("/labels/:uid/t/:token/:size", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", labelIconSvg)
			return
		}

		start := time.Now()
		conf := get.Config()
		thumbName := thumb.Name(clean.Token(c.Param("size")))
		uid := clean.UID(c.Param("uid"))

		size, ok := thumb.Sizes[thumbName]

		if !ok {
			log.Errorf("%s: invalid size %s", labelCover, clean.Log(thumbName.String()))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		cache := get.CoverCache()
		cacheKey := CacheKey(labelCover, uid, string(thumbName))

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Tracef("api-v1: cache hit for %s [%s]", cacheKey, time.Since(start))

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
			log.Errorf("%s: file %s is missing", labelCover, clean.Log(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logErr(labelCover, f.Update("FileMissing", true))

			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if size.ExceedsLimit() {
			log.Debugf("%s: using original, size exceeds limit (width %d, height %d)", labelCover, size.Width, size.Height)

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
			log.Errorf("%s: %s", labelCover, err)
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("%s: %s has empty thumb name - you may have found a bug", labelCover, filepath.Base(fileName))
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
