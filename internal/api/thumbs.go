package api

import (
	"encoding/json"
	"fmt"
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

type ThumbCache struct {
	FileName  string
	ShareName string
}

type ByteCache struct {
	Data []byte
}

// GET /api/v1/t/:hash/:token/:type
//
// Parameters:
//   hash: string file hash as returned by the search API
//   token: string security token (see config)
//   type: string thumb type, see photoprism.ThumbnailTypes
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

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("thumbs: invalid type %s", txt.Quote(typeName))
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		cache := service.Cache()
		cacheKey := fmt.Sprintf("thumbs:%s:%s", fileHash, typeName)

		if cacheData, err := cache.Get(cacheKey); err == nil {
			log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))

			var cached ThumbCache

			if err := json.Unmarshal(cacheData, &cached); err != nil {
				log.Errorf("thumbs: %s not found", fileHash)
				c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
				return
			}

			if !fs.FileExists(cached.FileName) {
				log.Errorf("thumbs: %s not found", fileHash)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			}

			if c.Query("download") != "" {
				c.FileAttachment(cached.FileName, cached.ShareName)
			} else {
				c.File(cached.FileName)
			}

			return
		}

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
			logError("thumbnail", f.Update("FileMissing", true))

			if f.AllFilesMissing() {
				log.Infof("thumbs: deleting photo, all files missing for %s", txt.Quote(f.FileName))

				logError("thumbnail", f.RelatedPhoto().Delete(false))
			}

			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsLimit() && c.Query("download") == "" {
			log.Debugf("thumbs: using original, size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)

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

		// Cache thumbnail filename.
		if cached, err := json.Marshal(ThumbCache{thumbnail, f.ShareFileName()}); err == nil {
			logError("thumbnail", cache.Set(cacheKey, cached))
			log.Debugf("cached %s [%s]", cacheKey, time.Since(start))
		}

		if c.Query("download") != "" {
			c.FileAttachment(thumbnail, f.ShareFileName())
		} else {
			c.File(thumbnail)
		}
	})
}

// GET /api/v1/albums/:uid/t/:token/:type
//
// Parameters:
//   uid: string album uid
//   token: string security token (see config)
//   type: string thumb type, see photoprism.ThumbnailTypes
func AlbumThumb(router *gin.RouterGroup) {
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
			log.Errorf("album-thumbs: invalid type %s", typeName)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		cache := service.Cache()
		cacheKey := fmt.Sprintf("album-thumbs:%s:%s", uid, typeName)

		if cacheData, err := cache.Get(cacheKey); err == nil {
			log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))

			var cached ThumbCache

			if err := json.Unmarshal(cacheData, &cached); err != nil {
				log.Errorf("album-thumbs: %s not found", uid)
				c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
				return
			}

			if !fs.FileExists(cached.FileName) {
				log.Errorf("album-thumbs: %s not found", uid)
				c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
				return
			}

			if c.Query("download") != "" {
				c.FileAttachment(cached.FileName, cached.ShareName)
			} else {
				c.File(cached.FileName)
			}

			return
		}

		f, err := query.AlbumCoverByUID(uid)

		if err != nil {
			log.Debugf("album-thumbs: no photos yet, using generic image for %s", uid)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("album-thumbs: could not find original for %s", fileName)
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			log.Warnf("album-thumbs: %s is missing", txt.Quote(f.FileName))
			logError("album-thumbnail", f.Update("FileMissing", true))
			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsLimit() && c.Query("download") == "" {
			log.Debugf("album-thumbs: using original, size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)
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
			log.Errorf("album-thumbs: %s has empty thumb name - bug?", filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
			return
		}

		if cached, err := json.Marshal(ThumbCache{thumbnail, f.ShareFileName()}); err == nil {
			logError("album-thumbnail", cache.Set(cacheKey, cached))
			log.Debugf("cached %s [%s]", cacheKey, time.Since(start))
		}

		if c.Query("download") != "" {
			c.FileAttachment(thumbnail, f.ShareFileName())
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
func LabelThumb(router *gin.RouterGroup) {
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
			log.Errorf("label-thumbs: invalid type %s", txt.Quote(typeName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		cache := service.Cache()
		cacheKey := fmt.Sprintf("label-thumbs:%s:%s", uid, typeName)

		if cacheData, err := cache.Get(cacheKey); err == nil {
			log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))

			var cached ThumbCache

			if err := json.Unmarshal(cacheData, &cached); err != nil {
				log.Errorf("label-thumbs: %s not found", uid)
				c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
				return
			}

			if !fs.FileExists(cached.FileName) {
				log.Errorf("label-thumbs: %s not found", uid)
				c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
				return
			}

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
			log.Errorf("label-thumbs: file %s is missing", txt.Quote(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logError("label-thumbnail", f.Update("FileMissing", true))

			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsLimit() {
			log.Debugf("label-thumbs: using original, size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)

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
			log.Errorf("label-thumbs: %s", err)
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("label-thumbs: %s has empty thumb name - bug?", filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", labelIconSvg)
			return
		}

		if cached, err := json.Marshal(ThumbCache{thumbnail, f.ShareFileName()}); err == nil {
			logError("label-thumbnail", cache.Set(cacheKey, cached))
			log.Debugf("cached %s [%s]", cacheKey, time.Since(start))
		}

		if c.Query("download") != "" {
			c.FileAttachment(thumbnail, f.ShareFileName())
		} else {
			c.File(thumbnail)
		}
	})
}
