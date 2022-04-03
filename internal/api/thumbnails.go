package api

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/crop"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/thumb"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// GetThumb returns a thumbnail image matching the file hash, crop area, and type.
//
// GET /api/v1/t/:thumb/:token/:size
//
// Parameters:
//   thumb: string sha1 file hash plus optional crop area
//   token: string url security token, see config
//   size: string thumb type, see thumb.Sizes
func GetThumb(router *gin.RouterGroup) {
	router.GET("/t/:thumb/:token/:size", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		logPrefix := "thumb"

		start := time.Now()
		conf := service.Config()
		download := c.Query("download") != ""
		fileHash, cropArea := crop.ParseThumb(sanitize.Token(c.Param("thumb")))

		// Is cropped thumbnail?
		if cropArea != "" {
			cropName := crop.Name(sanitize.Token(c.Param("size")))

			cropSize, ok := crop.Sizes[cropName]

			if !ok {
				log.Errorf("%s: invalid size %s", logPrefix, sanitize.Log(string(cropName)))
				c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
				return
			}

			fileName, err := crop.FromRequest(fileHash, cropArea, cropSize, conf.ThumbPath())

			if err != nil {
				log.Warnf("%s: %s", logPrefix, err)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			} else if fileName == "" {
				log.Errorf("%s: empty file name, potential bug", logPrefix)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			}

			if download {
				c.FileAttachment(fileName, cropName.Jpeg())
			} else {
				AddThumbCacheHeader(c)
				c.File(fileName)
			}

			return
		}

		thumbName := thumb.Name(sanitize.Token(c.Param("size")))

		size, ok := thumb.Sizes[thumbName]

		if !ok {
			log.Errorf("%s: invalid size %s", logPrefix, sanitize.Log(thumbName.String()))
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		if size.Uncached() && !conf.ThumbUncached() {
			thumbName, size = thumb.Find(conf.ThumbSizePrecached())

			if thumbName == "" {
				log.Errorf("%s: invalid size %d", logPrefix, conf.ThumbSizePrecached())
				c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
				return
			}
		}

		cache := service.ThumbCache()
		cacheKey := CacheKey("thumbs", fileHash, string(thumbName))

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Tracef("api: cache hit for %s [%s]", cacheKey, time.Since(start))

			cached := cacheData.(ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("%s: %s not found", logPrefix, fileHash)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			}

			if c.Query("download") != "" {
				c.FileAttachment(cached.FileName, cached.ShareName)
			} else {
				AddThumbCacheHeader(c)
				c.File(cached.FileName)
			}

			return
		}

		// Return existing thumbs straight away.
		if !download {
			if fileName, err := thumb.FileName(fileHash, conf.ThumbPath(), size.Width, size.Height, size.Options...); err == nil && fs.FileExists(fileName) {
				AddThumbCacheHeader(c)
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
			log.Errorf("%s: file %s is missing", logPrefix, sanitize.Log(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logError(logPrefix, f.Update("FileMissing", true))

			if f.AllFilesMissing() {
				log.Infof("%s: deleting photo, all files missing for %s", logPrefix, sanitize.Log(f.FileName))

				if _, err := f.RelatedPhoto().Delete(false); err != nil {
					log.Errorf("%s: %s while deleting %s", logPrefix, err, sanitize.Log(f.FileName))
				}
			}

			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if size.ExceedsLimit() && c.Query("download") == "" {
			log.Debugf("%s: using original, size exceeds limit (width %d, height %d)", logPrefix, size.Width, size.Height)

			AddThumbCacheHeader(c)
			c.File(fileName)

			return
		}

		var thumbnail string

		if conf.ThumbUncached() || size.Uncached() {
			thumbnail, err = thumb.FromFile(fileName, f.FileHash, conf.ThumbPath(), size.Width, size.Height, f.FileOrientation, size.Options...)
		} else {
			thumbnail, err = thumb.FromCache(fileName, f.FileHash, conf.ThumbPath(), size.Width, size.Height, size.Options...)
		}

		if err != nil {
			log.Errorf("%s: %s", logPrefix, err)
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("%s: %s has empty thumb name - bug?", logPrefix, filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		}

		cache.SetDefault(cacheKey, ThumbCache{thumbnail, f.ShareBase(0)})
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		if download {
			c.FileAttachment(thumbnail, f.DownloadName(DownloadName(c), 0))
		} else {
			AddThumbCacheHeader(c)
			c.File(thumbnail)
		}
	})
}
