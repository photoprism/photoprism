package api

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

// GET /api/v1/thumbnails/:hash/:type
//
// Parameters:
//   hash: string The file hash as returned by the search API
//   type: string Thumbnail type, see photoprism.ThumbnailTypes
func GetThumbnail(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/thumbnails/:hash/:type", func(c *gin.Context) {
		fileHash := c.Param("hash")
		typeName := c.Param("type")

		thumbType, ok := thumb.Types[typeName]

		if !ok {
			log.Errorf("thumbs: invalid type \"%s\"", typeName)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
			return
		}

		db := conf.Db()
		q := query.New(conf.OriginalsPath(), db)
		f, err := q.FindFileByHash(fileHash)

		if err != nil {
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)
			return
		}

		if f.FileError != "" {
			c.Data(http.StatusBadRequest, "image/svg+xml", brokenIconSvg)
			return
		}

		fileName := path.Join(conf.OriginalsPath(), f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("could not find original for thumbnail: %s", fileName)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			f.FileMissing = true
			db.Save(&f)
			return
		}

		// Use original file if thumb size exceeds limit, see https://github.com/photoprism/photoprism/issues/157
		if thumbType.ExceedsLimit() && c.Query("download") == "" {
			log.Debugf("photo: using original, thumbnail size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)

			c.File(fileName)

			return
		}

		if thumbnail, err := thumb.FromFile(fileName, f.FileHash, conf.ThumbnailsPath(), thumbType.Width, thumbType.Height, thumbType.Options...); err == nil {
			if c.Query("download") != "" {
				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.DownloadFileName()))
			}

			c.File(thumbnail)
		} else {
			f.FileError = err.Error()
			db.Save(&f)

			c.Data(http.StatusBadRequest, "image/svg+xml", brokenIconSvg)
		}
	})
}
