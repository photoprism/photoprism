package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api/download"
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TODO: GET /api/v1/dl/file/:hash
// TODO: GET /api/v1/dl/photo/:uid
// TODO: GET /api/v1/dl/album/:uid

// DownloadName returns the download file name type.
func DownloadName(c *gin.Context) customize.DownloadName {
	switch c.Query("name") {
	case "file":
		return customize.DownloadNameFile
	case "share":
		return customize.DownloadNameShare
	case "original":
		return customize.DownloadNameOriginal
	default:
		return get.Config().Settings().Download.Name
	}
}

// GetDownload returns the raw file data.
//
//	@Summary	returns the raw file data
//	@Id			GetDownload
//	@Tags		Images, Files
//	@Produce	application/octet-stream
//	@Failure	403,404	{file}	image/svg+xml
//	@Success	200		{file}	application/octet-stream
//	@Param		file	path	string	true	"file hash or unique download id"
//	@Router		/api/v1/dl/{file} [get]
func GetDownload(router *gin.RouterGroup) {
	router.GET("/dl/:file", func(c *gin.Context) {
		id := clean.Token(c.Param("file"))

		// Check for temporary download if the file is identified by a UUID string.
		if rnd.IsUUID(id) {
			fileName, fileErr := download.Find(id)

			if fileErr != nil {
				AbortForbidden(c)
				return
			} else if !fs.FileExists(fileName) {
				AbortNotFound(c)
				return
			}

			c.FileAttachment(fileName, filepath.Base(fileName))
			return
		}

		// If the file is identified by its hash, a valid download token is required.
		if InvalidDownloadToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		f, err := query.FileByHash(id)

		if err != nil {
			c.AbortWithStatusJSON(404, gin.H{"error": err.Error()})
			return
		}

		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("download: file %s is missing", clean.Log(f.FileName))
			c.Data(404, "image/svg+xml", brokenIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore.
			logErr("download", f.Update("FileMissing", true))

			return
		}

		c.FileAttachment(fileName, f.DownloadName(DownloadName(c), 0))
	})
}
