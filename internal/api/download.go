package api

import (
	"fmt"
	"path"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/fs"

	"github.com/gin-gonic/gin"
)

// TODO: GET /api/v1/dl/file/:hash
// TODO: GET /api/v1/dl/photo/:uuid
// TODO: GET /api/v1/dl/album/:uuid

// GET /api/v1/download/:hash
//
// Parameters:
//   hash: string The file hash as returned by the search API
func GetDownload(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/download/:hash", func(c *gin.Context) {
		fileHash := c.Param("hash")

		f, err := query.FileByHash(fileHash)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": err.Error()})
			return
		}

		fileName := path.Join(conf.OriginalsPath(), f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("could not find original: %s", fileHash)
			c.Data(404, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			f.FileMissing = true
			entity.Db().Save(&f)
			return
		}

		downloadFileName := f.ShareFileName()

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadFileName))

		c.File(fileName)
	})
}
