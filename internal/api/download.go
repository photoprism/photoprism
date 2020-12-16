package api

import (
	"fmt"
	"github.com/photoprism/photoprism/internal/entity"
	"net/http"

	"github.com/photoprism/photoprism/internal/service"

	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gin-gonic/gin"
)

// TODO: GET /api/v1/dl/file/:hash
// TODO: GET /api/v1/dl/photo/:uid
// TODO: GET /api/v1/dl/album/:uid

// GET /api/v1/dl/:hash
//
// Parameters:
//   hash: string The file hash as returned by the search API
func GetDownload(router *gin.RouterGroup) {
	router.GET("/dl/:hash", func(c *gin.Context) {
		if InvalidDownloadToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		fileHash := c.Param("hash")

		f, err := query.FileByHash(fileHash)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": err.Error()})
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("download: file %s is missing", txt.Quote(f.FileName))
			c.Data(404, "image/svg+xml", brokenIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logError("download", f.Update("FileMissing", true))

			return
		}

		name := entity.DownloadNameFile

		switch c.Query("name") {
		case "file":
			name = entity.DownloadNameFile
		case "share":
			name = entity.DownloadNameShare
		case "original":
			name = entity.DownloadNameOriginal
		default:
			name = service.Config().Settings().Download.Name
		}

		var downloadName string

		switch name {
		case entity.DownloadNameFile:
			downloadName = f.Base()
		case entity.DownloadNameOriginal:
			downloadName = f.OriginalBase()
		default:
			downloadName = f.ShareBase()
		}

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", downloadName))

		c.File(fileName)
	})
}
