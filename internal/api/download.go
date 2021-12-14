package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// TODO: GET /api/v1/dl/file/:hash
// TODO: GET /api/v1/dl/photo/:uid
// TODO: GET /api/v1/dl/album/:uid

// DownloadName returns the download file name type.
func DownloadName(c *gin.Context) entity.DownloadName {
	switch c.Query("name") {
	case "file":
		return entity.DownloadNameFile
	case "share":
		return entity.DownloadNameShare
	case "original":
		return entity.DownloadNameOriginal
	default:
		return service.Config().Settings().Download.Name
	}
}

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

		fileHash := sanitize.Token(c.Param("hash"))

		f, err := query.FileByHash(fileHash)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": err.Error()})
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("download: file %s is missing", sanitize.Log(f.FileName))
			c.Data(404, "image/svg+xml", brokenIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logError("download", f.Update("FileMissing", true))

			return
		}

		c.FileAttachment(fileName, f.DownloadName(DownloadName(c), 0))
	})
}
