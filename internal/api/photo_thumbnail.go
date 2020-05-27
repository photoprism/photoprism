package api

import (
	"net/http"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

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

		fileHash := c.Param("hash")
		typeName := c.Param("type")

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("photo: invalid thumb type %s", txt.Quote(typeName))
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
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
			log.Errorf("photo: file %s is missing", txt.Quote(f.FileName))
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			report("photo", f.Update("FileMissing", true))

			if f.AllFilesMissing() {
				log.Infof("photo: deleting photo, all files missing for %s", txt.Quote(f.FileName))

				report("photo", f.RelatedPhoto().Delete(false))
			}

			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsLimit() && c.Query("download") == "" {
			log.Debugf("photo: using original, thumbnail size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)

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
			log.Errorf("photo: %s", err)
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		} else if thumbnail == "" {
			log.Errorf("photo: thumbnail name for %s is empty - bug?", filepath.Base(fileName))
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		}

		if c.Query("download") != "" {
			c.FileAttachment(thumbnail, f.ShareFileName())
		} else {
			c.File(thumbnail)
		}
	})
}
