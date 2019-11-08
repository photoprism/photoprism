package api

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
)

// GET /api/v1/download/:hash
//
// Parameters:
//   hash: string The file hash as returned by the search API
func GetDownload(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/download/:hash", func(c *gin.Context) {
		fileHash := c.Param("hash")

		search := photoprism.NewSearch(conf.OriginalsPath(), conf.Db())
		file, err := search.FindFileByHash(fileHash)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": err.Error()})
			return
		}

		fileName := fmt.Sprintf("%s/%s", conf.OriginalsPath(), file.FileName)

		if !util.Exists(fileName) {
			log.Errorf("could not find original: %s", fileHash)
			c.Data(404, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			file.FileMissing = true
			conf.Db().Save(&file)
			return
		}

		downloadFileName := file.DownloadFileName()

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadFileName))

		c.File(fileName)
	})
}
