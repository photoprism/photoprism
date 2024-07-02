package api

import (
	"archive/zip"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// DownloadAlbum streams the album contents as zip archive.
//
// GET /api/v1/albums/:uid/dl
func DownloadAlbum(router *gin.RouterGroup) {
	router.GET("/albums/:uid/dl", func(c *gin.Context) {
		if InvalidDownloadToken(c) {
			AbortForbidden(c)
			return
		}

		conf := get.Config()

		if !conf.Settings().Features.Download {
			AbortFeatureDisabled(c)
			return
		}

		start := time.Now()
		a, err := query.AlbumByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		files, err := search.AlbumPhotos(a, 10000, true)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		zipFileName := a.ZipName()

		AddDownloadHeader(c, zipFileName)

		zipWriter := zip.NewWriter(c.Writer)
		defer func(w *zip.Writer) {
			logErr("zip", w.Close())
		}(zipWriter)

		var aliases = make(map[string]int)

		for _, file := range files {
			if file.FileName == "" {
				log.Warnf("album: %s cannot be downloaded (empty file name)", clean.Log(file.FileUID))
				continue
			} else if file.FileHash == "" {
				log.Warnf("album: %s cannot be downloaded (empty file hash)", clean.Log(file.FileName))
				continue
			}

			if file.FileSidecar {
				log.Debugf("album: sidecar file %s not included in download", clean.Log(file.FileName))
				continue
			}

			fileName := photoprism.FileName(file.FileRoot, file.FileName)
			alias := file.ShareBase(0)
			key := strings.ToLower(alias)

			if seq := aliases[key]; seq > 0 {
				alias = file.ShareBase(seq)
			}

			aliases[key] += 1

			if fs.FileExists(fileName) {
				if zipErr := fs.ZipFile(zipWriter, fileName, alias, false); zipErr != nil {
					log.Errorf("download: failed to add %s (%s)", clean.Log(file.FileName), zipErr)
					Abort(c, http.StatusInternalServerError, i18n.ErrZipFailed)
					return
				}

				log.Infof("download: added %s as %s", clean.Log(file.FileName), clean.Log(alias))
			} else {
				log.Warnf("download: %s not found", clean.Log(file.FileName))
			}
		}

		log.Infof("album: %s has been downloaded [%s]", clean.Log(a.AlbumTitle), time.Since(start))
	})
}
