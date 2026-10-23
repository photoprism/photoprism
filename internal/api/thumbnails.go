package api

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// GetThumb returns a thumbnail image matching the file hash, crop area, and type.
//
//	@Summary		returns a thumbnail image with the requested size
//	@Description	For more information see:
//	@Description	- https://docs.photoprism.app/developer-guide/api/thumbnails/#image-endpoint-uri
//	@Id				GetThumb
//	@Produce		image/jpeg, image/svg+xml
//	@Tags			Images, Files
//	@Failure		400,403,404,500	{file}	image/svg+xml
//	@Success		200				{file}	image/svg+xml
//	@Success		200				{file}	image/jpg
//	@Param			thumb			path	string	true	"SHA1 file hash, optionally with a crop area suffixed, e.g. '-016014058037'"
//	@Param			token			path	string	true	"user-specific security token provided with session or 'public' when running PhotoPrism in public mode"
//	@Param			size			path	string	true	"thumbnail size; sizes above the server limit are reduced"	Enums(tile_50, tile_100, left_224, right_224, tile_224, tile_500, fit_720, tile_1080, fit_1280, fit_1600, fit_1920, fit_2048, fit_2560, fit_3840, fit_4096, fit_5120, fit_7680, fit_15360)
//	@Param			download		query	string	false	"returns the thumbnail as an attachment named <stem>_<size> with the extension of the thumbnail, or after the crop size for a cropped thumbnail; placeholders then use 400, 404, or 500"
//	@Param			name			query	string	false	"attachment name type (file, share, original)"
//	@Router			/api/v1/t/{thumb}/{token}/{size} [get]
func GetThumb(router *gin.RouterGroup) {
	router.GET("/t/:thumb/:token/:size", func(c *gin.Context) {
		if InvalidPreviewToken(c, acl.Resources{acl.ResourcePhotos, acl.ResourceFiles}) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		logPrefix := "thumb"

		start := time.Now()
		conf := get.Config()
		attachment := c.Query("download") != ""
		fileHash, cropArea := crop.ParseThumb(clean.Token(c.Param("thumb")))

		// Is cropped thumbnail?
		if cropArea != "" {
			cropName := crop.Name(clean.Token(c.Param("size")))

			cropSize, ok := crop.Sizes[cropName]

			if !ok {
				log.Errorf("%s: invalid size %s", logPrefix, clean.Log(string(cropName)))
				abortThumb(c, attachment, http.StatusBadRequest, photoIconSvg)
				return
			}

			fileName, err := crop.FromRequest(fileHash, cropArea, cropSize, conf.ThumbCachePath())

			if err != nil {
				log.Warnf("%s: %s in %s", logPrefix, clean.Error(err), thumbLogName(fileHash))
				abortThumb(c, attachment, http.StatusNotFound, brokenIconSvg)
				return
			} else if fileName == "" {
				log.Errorf("%s: empty file name - you may have found a bug", logPrefix)
				abortThumb(c, attachment, http.StatusInternalServerError, brokenIconSvg)
				return
			}

			// Add HTTP cache header.
			AddImmutableCacheHeader(c)

			if attachment {
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
			abortThumb(c, attachment, http.StatusBadRequest, photoIconSvg)
			return
		}

		if size.Uncached() && !conf.ThumbUncached() {
			sizeName, size = thumb.Find(conf.ThumbSizePrecached())

			if sizeName == "" {
				log.Errorf("%s: invalid size %d", logPrefix, conf.ThumbSizePrecached())
				abortThumb(c, attachment, http.StatusInternalServerError, photoIconSvg)
				return
			}
		}

		// Reduce the size to the largest one that can be rendered, if needed.
		if size.ExceedsLimit() {
			size = size.Clamp()
			sizeName = size.Name
			log.Debugf("%s: using %s, requested size exceeds limit", logPrefix, sizeName)
		}

		cache := get.ThumbCache()

		// Keyed by the requested size, which the rendered size may be smaller than.
		// That mapping only depends on the file, so the entry stays valid.
		cacheKey := CacheKey("thumbs", fileHash, string(sizeName))

		// Downloads are named from the indexed file, so they skip the cache.
		if cacheData, ok := cache.Get(cacheKey); ok && !attachment {
			log.Tracef("api: cache hit for %s [%s]", cacheKey, time.Since(start))

			cached := cacheData.(ThumbCache)

			if !fs.FileExists(cached.FileName) {
				log.Errorf("%s: cached %s of %s is missing", logPrefix, sizeName, thumbLogName(fileHash))
				event.SystemWarn([]string{"thumb", "cached file %s is missing"}, clean.Log(cached.FileName))
				cache.Delete(cacheKey)
				c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
				return
			}

			// Add HTTP cache header.
			AddImmutableCacheHeader(c)
			c.File(cached.FileName)

			return
		}

		// Return existing thumbs straight away.
		if !attachment {
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
			abortThumb(c, attachment, http.StatusNotFound, photoIconSvg)
			return
		}

		// Find supported preview image if media file is not a JPEG or PNG.
		if f.NoJpeg() && f.NoPng() {
			if f, err = query.FileByPhotoUID(f.PhotoUID); err != nil {
				abortThumb(c, attachment, http.StatusNotFound, fileIconSvg)
				return
			}
		}

		// Return SVG icon as placeholder if file has errors.
		if f.FileError != "" {
			abortThumb(c, attachment, http.StatusNotFound, brokenIconSvg)
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if fileName, err = fs.Resolve(fileName); err != nil {
			log.Errorf("%s: file %s is missing", logPrefix, clean.Log(f.FileName))
			abortThumb(c, attachment, http.StatusNotFound, brokenIconSvg)

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
			// The smallest fitting size can resolve above the requested one, so this is clamped again.
			size = thumb.FitBounds(f.Bounds()).Clamp()
			log.Tracef("%s: smallest fitting size for %s is %s (width %d, height %d)", logPrefix, clean.Log(f.FileName), size.Name, size.Width, size.Height)
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
		switch {
		case errors.Is(err, thumb.ErrNotCached):
			log.Errorf("%s: %s in %s", logPrefix, clean.Error(err), clean.Log(f.FileName))
			abortThumb(c, attachment, http.StatusNotFound, brokenIconSvg)
			return
		case err != nil:
			log.Errorf("%s: %s in %s", logPrefix, clean.Error(err), clean.Log(f.FileName))
			abortThumb(c, attachment, http.StatusInternalServerError, brokenIconSvg)
			return
		case thumbName == "":
			log.Errorf("%s: %s has empty thumb name - you may have found a bug", logPrefix, filepath.Base(fileName))
			abortThumb(c, attachment, http.StatusInternalServerError, brokenIconSvg)
			return
		}

		// Cache thumbnail filename to reduce the number of index queries.
		cache.SetDefault(cacheKey, ThumbCache{FileName: thumbName})
		log.Debugf("cached %s [%s]", cacheKey, time.Since(start))

		// Add HTTP cache header.
		AddImmutableCacheHeader(c)

		// Return requested content.
		if attachment {
			c.FileAttachment(thumbName, thumbDownloadName(f, DownloadName(c), sizeName, thumbName))
		} else {
			c.File(thumbName)
		}
	})
}

// abortThumb sends a placeholder icon, with the given status code only if the thumbnail was requested as a download.
func abortThumb(c *gin.Context, download bool, code int, icon []byte) {
	if !download {
		code = http.StatusOK
	}

	c.Data(code, "image/svg+xml", icon)
}

// thumbDownloadName returns the attachment name of a thumbnail: the selected download name of the original it
// belongs to without its last extension, followed by the size and the extension of the thumbnail file.
func thumbDownloadName(f *entity.File, n customize.DownloadName, size thumb.Name, thumbName string) string {
	return clean.AttachmentName(fs.StripExt(thumbOwner(f).DownloadName(n, 0)) + "_" + size.String() + filepath.Ext(thumbName))
}

// thumbOwner returns the original of the same picture that a preview outside the originals folder was made from,
// matched by its full name or, if unambiguous, by its stem, or the file itself otherwise.
func thumbOwner(f *entity.File) *entity.File {
	if f.FileRoot == entity.RootOriginals || f.PhotoID == 0 {
		return f
	}

	var originals entity.Files

	if err := entity.Db().Where("photo_id = ? AND file_root = ? AND file_sidecar = ?", f.PhotoID, entity.RootOriginals, false).
		Preload("Photo").Find(&originals).Error; err != nil {
		return f
	}

	stem := strings.TrimSuffix(f.FileName, filepath.Ext(f.FileName))
	var match *entity.File

	for i := range originals {
		if originals[i].FileName == stem {
			return &originals[i]
		}
	}

	for i := range originals {
		if name := originals[i].FileName; strings.TrimSuffix(name, filepath.Ext(name)) != stem {
			continue
		} else if match != nil {
			return f
		}

		match = &originals[i]
	}

	if match != nil {
		return match
	}

	return f
}

// thumbLogName returns the name of the original a thumbnail belongs to, for naming it in a log message
// instead of its file hash.
func thumbLogName(fileHash string) string {
	if f, err := query.FileByHash(fileHash); err == nil && f.FileName != "" {
		return clean.Log(f.FileName)
	}

	return "unknown file"
}
