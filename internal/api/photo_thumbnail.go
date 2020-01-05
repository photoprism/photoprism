package api

import (
	"fmt"
	"net/http"
	"path"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
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

		thumbType, ok := photoprism.ThumbnailTypes[typeName]

		if !ok {
			log.Errorf("invalid type: %s", typeName)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
			return
		}

		q := query.New(conf.OriginalsPath(), conf.Db())
		file, err := q.FindFileByHash(fileHash)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		fileName := path.Join(conf.OriginalsPath(), file.FileName)

		if !util.Exists(fileName) {
			log.Errorf("could not find original for thumbnail: %s", fileName)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			file.FileMissing = true
			conf.Db().Save(&file)
			return
		}

		if thumbnail, err := photoprism.ThumbnailFromFile(fileName, file.FileHash, conf.ThumbnailsPath(), thumbType.Width, thumbType.Height, thumbType.Options...); err == nil {
			if c.Query("download") != "" {
				downloadFileName := file.DownloadFileName()

				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadFileName))
			}

			c.File(thumbnail)
		} else {
			log.Errorf("could not create thumbnail: %s", err)
			c.Data(http.StatusBadRequest, "image/svg+xml", photoIconSvg)
		}
	})
}
