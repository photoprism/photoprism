package api

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/video"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GET /api/v1/videos/:hash/:type
//
// Parameters:
//   hash: string The photo or video file hash as returned by the search API
//   type: string Video type
func GetVideo(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/videos/:hash/:type", func(c *gin.Context) {
		fileHash := c.Param("hash")
		typeName := c.Param("type")

		_, ok := video.Types[typeName]

		if !ok {
			log.Errorf("video: invalid type %s", txt.Quote(typeName))
			c.Data(http.StatusOK, "image/svg+xml", videoIconSvg)
			return
		}

		f, err := query.FileByHash(fileHash)

		if err != nil {
			log.Errorf("video: db error %s", err.Error())
			c.Data(http.StatusOK, "image/svg+xml", videoIconSvg)
			return
		}

		if f.FileError != "" {
			log.Errorf("video: file error %s", f.FileError)
			c.Data(http.StatusOK, "image/svg+xml", videoIconSvg)
			return
		}

		fileName := path.Join(conf.OriginalsPath(), f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("video: could not find file for %s", fileName)
			c.Data(http.StatusOK, "image/svg+xml", videoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			f.FileMissing = true
			entity.Db().Save(&f)
			return
		}

		if c.Query("download") != "" {
			c.FileAttachment(fileName, f.ShareFileName())
		} else {
			c.File(fileName)
		}

		return
	})
}
