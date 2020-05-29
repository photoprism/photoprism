package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
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
	Data  []byte
}

// GET /api/v1/t/:hash/:token/:type
//
// Parameters:
//   hash: string The file hash as returned by the search API
//   type: string Thumbnail type, see photoprism.ThumbnailTypes
func GetThumbnail(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/t/:hash/:token/:type", func(c *gin.Context) {
		if InvalidToken(c, conf) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		start := time.Now()
		fileHash := c.Param("hash")
		typeName := c.Param("type")

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("thumbnail: invalid type %s", txt.Quote(typeName))
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		cache := service.Cache()
		cacheKey := fmt.Sprintf("thumbnail:%s:%s", fileHash, typeName)

		if cacheData, err := cache.Get(cacheKey); err == nil {
			log.Debugf("cache hit for %s [%s]", cacheKey, time.Since(start))

			var cached ThumbCache

			if err := json.Unmarshal(cacheData, &cached); err != nil {
				log.Errorf("thumbnail: %s not found", fileHash)
				c.Data(http.StatusOK, "image/svg+xml", albumIconSvg)
				return
			}

			if !fs.FileExists(cached.FileName) {
				log.Errorf("thumbnail: %s not found", fileHash)
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

		fileName := path.Join(conf.OriginalsPath(), f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("thumbnail: file %s is missing", txt.Quote(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logError("thumbnail", f.Update("FileMissing", true))

			if f.AllFilesMissing() {
				log.Infof("thumbnail: deleting photo, all files missing for %s", txt.Quote(f.FileName))

				logError("thumbnail", f.RelatedPhoto().Delete(false))
			}

			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsLimit() && c.Query("download") == "" {
			log.Debugf("thumbnail: using original, size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)

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
			log.Errorf("thumbnail: %s", err)
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("thumbnail: %s has empty thumb name - bug?", filepath.Base(fileName))
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
