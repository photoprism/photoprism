package api

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// GetThumb returns a thumbnail image matching the file hash, crop area, and type.
//
//	@Summary	returns a thumbnail image with the requested size
//	@Id			GetThumb
//	@Produce	image/jpeg
//	@Tags		Images, Files
//	@Param		thumb path string true "SHA1 file hash, optionally with a crop area suffixed, e.g. '-016014058037'"
//	@Param		token path string true "user-specific security token provided with session"
//	@Param		size path string true "thumbnail size"
//	@Router		/api/v1/t/{hash}/{token}/{size} [get]
func GetThumb(router *gin.RouterGroup) {
	router.GET("/t/:thumb/:token/:size", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		logPrefix := "thumb"

		start := time.Now()
		conf := get.Config()
		download := c.Query("download") != ""
		fileHash, cropArea := crop.ParseThumb(clean.Token(c.Param("thumb")))

		// Is cropped thumbnail?
		if cropArea != "" {
			cropName := crop.Name(clean.Token(c.Param("size")))

			cropSize, ok := crop.Sizes[cropName]

			if !ok {
				log.Errorf("%s: invalid size %s", logPrefix, clean.Log(string(cropName)))
				c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
				return
			}

			fileName, err := crop.FromRequest(fileHash, cropArea, cropSize, conf.ThumbCachePath())

			if err != nil {
				log.Warnf("%s: %s", logPrefix, err)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			} else if fileName == "" {
				log.Errorf("%s: empty file name - you may have found a bug", logPrefix)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			}

			// Add HTTP cache header.
			AddImmutableCacheHeader(c)

			if download {
				c.FileAttachment(fileName, cropName.Jpeg())
			} else {
				c.File(fileName)
			}

			return
		}

		sizeName := thumb.Name(clean.Token(c.Param("size")))

		size, ok := thumb.Sizes[sizeName]

		if !ok {
			log.Errorf("%s: invalid size %s", logPrefix, clean.Log(sizeName.String()))
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		if size.Uncached() && !conf.ThumbUncached() {
			sizeName, size = thumb.Find(conf.ThumbSizePrecached())

			if sizeName == "" {
				log.Errorf("%s: invalid size %d", logPrefix, conf.ThumbSizePrecached())
				c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
				return
			}
		}

		cache := get.ThumbCache()
		cacheKey := CacheKey("thumbs", fileHash, string(sizeName))

		if cacheData, ok := cache.Get(cacheKey); ok {
			log.Tracef("api-v1: cache hit for %s [%s]", cacheKey, time.Since(start))

			cached := cacheData.(ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("%s: %s not found", logPrefix, fileHash)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			}

			// Add HTTP cache header.
			AddImmutableCacheHeader(c)

			if download {
				c.FileAttachment(cached.FileName, cached.ShareName)
			} else {
				c.File(cached.FileName)
			}

			return
		}

		// Return existing thumbs straight away.
		if !download {
			if fileName, err := size.ResolvedName(fileHash, conf.ThumbCachePath()); err == nil {
				// Add HTTP cache header.
				AddImmutableCacheHeader(c)

				// Return requested content.
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

		// Find supported preview image if media file is not a JPEG or PNG.
		if f.NoJPEG() && f.NoPNG() {
			if f, err = query.FileByPhotoUID(f.PhotoUID); err != nil {
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

		if fileName, err = fs.Resolve(fileName); err != nil {
			log.Errorf("%s: file %s is missing", logPrefix, clean.Log(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logErr(logPrefix, f.Update("FileMissing", true))

			if f.AllFilesMissing() {
				log.Infof("%s: deleting photo, all files missing for %s", logPrefix, clean.Log(f.FileName))

				if _, err := f.RelatedPhoto().Delete(false); err != nil {
					log.Errorf("%s: %s while deleting %s", logPrefix, err, clean.Log(f.FileName))
				}
			}

			return
		}

		// Choose the smallest fitting size if the original image is smaller.
		if size.Fit && f.Bounds().In(size.Bounds()) {
			size = thumb.FitBounds(f.Bounds())
			log.Tracef("%s: smallest fitting size for %s is %s (width %d, height %d)", logPrefix, clean.Log(f.FileName), size.Name, size.Width, size.Height)
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if size.ExceedsLimit() && !download {
			log.Debugf("%s: using original, size exceeds limit (width %d, height %d)", logPrefix, size.Width, size.Height)

			// Add HTTP cache header.
			AddImmutableCacheHeader(c)

			// Return requested content.
			c.File(fileName)
			return
		}

		// thumbName is the thumbnail filename.
		var thumbName string

		// Try to find or create thumbnail image.
		if conf.ThumbUncached() || size.Uncached() {
			thumbName, err = size.FromFile(fileName, f.FileHash, conf.ThumbCachePath(), f.FileOrientation)
		} else {
			thumbName, err = size.FromCache(fileName, f.FileHash, conf.ThumbCachePath())
		}

		// Failed?
		if err != nil {
			log.Errorf("%s: %s", logPrefix, err)
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		} else if thumbName == "" {
			log.Errorf("%s: %s has empty thumb name - you may have found a bug", logPrefix, filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		}

		// Cache thumbnail filename to reduce the number of index queries.
		cache.SetDefault(cacheKey, ThumbCache{thumbName, f.ShareBase(0)})
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		// Add HTTP cache header.
		AddImmutableCacheHeader(c)

		// Return requested content.
		if download {
			c.FileAttachment(thumbName, f.DownloadName(DownloadName(c), 0))
		} else {
			c.File(thumbName)
		}
	})
}
