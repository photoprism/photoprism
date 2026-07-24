package api

import (
	"archive/zip"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// AlbumDownloadName returns the album download file name type.
func AlbumDownloadName(c *gin.Context) customize.DownloadName {
	switch c.Query("name") {
	case "file":
		return customize.DownloadNameFile
	case "share":
		return customize.DownloadNameShare
	case "original":
		return customize.DownloadNameOriginal
	default:
		return get.Config().Settings().Albums.Download.Name
	}
}

// DownloadAlbum streams the album contents as zip archive.
//
//	@Summary	streams the album contents as zip archive
//	@Id			DownloadAlbum
//	@Tags		Albums, Download
//	@Produce	application/zip
//	@Failure	403,404,500	{object}	i18n.Response
//	@Success	200			{file}		application/zip
//	@Param		uid			path		string	true	"Album UID"
//	@Router		/api/v1/albums/{uid}/dl [get]
func DownloadAlbum(router *gin.RouterGroup) {
	router.GET("/albums/:uid/dl", func(c *gin.Context) {
		conf := get.Config()

		// Reject up front when downloads are disabled, before the more expensive token/session resolution.
		if !conf.Settings().Features.Download || conf.Settings().Albums.Download.Disabled {
			AbortFeatureDisabled(c)
			return
		}

		sess, valid := AuthDownload(c)
		if !valid {
			AbortForbidden(c)
			return
		}

		start := time.Now()
		a, err := query.AlbumByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortAlbumNotFound(c)
			return
		}

		// The session that owns the download token (resolved by AuthDownload above) scopes the archive to
		// pictures it may see. A nil session means an unidentified requester (a coarse/instance token),
		// which is limited to public, non-private content below.

		// Deny access to albums the session may not view, mirroring GetAlbum. A nil session is a coarse
		// (public/static) token, handled by the public-content scoping below rather than this gate.
		if sess != nil && !albumViewableBySession(sess, a) {
			AbortAlbumNotFound(c)
			return
		}

		// Enumerate at the file level (not via a photo search) so real sidecar files (e.g. XMP) are
		// included when the Sidecar option is on. Archived and hidden pictures are always excluded;
		// private ones only when the session may view them, never for an unidentified requester.
		dl := conf.Settings().Albums.Download
		selection := query.AlbumDownloadSelection(dl.MediaRaw, dl.MediaSidecar, dl.Originals, search.PhotoSessionSeesPrivate(sess))
		files, err := query.SelectedFilesForSession(form.Selection{Albums: []string{a.AlbumUID}}, selection, sess)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		// Configure file names.
		dlName := AlbumDownloadName(c)
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

			fileName := photoprism.FileName(file.FileRoot, file.FileName)
			alias := file.DownloadName(dlName, 0)
			key := strings.ToLower(alias)

			if seq := aliases[key]; seq > 0 {
				alias = file.DownloadName(dlName, seq)
			}

			aliases[key]++

			if fs.FileExists(fileName) {
				if zipErr := fs.ZipFile(zipWriter, fileName, alias, false); zipErr != nil {
					log.Errorf("album: failed to zip %s (%s)", clean.Log(file.FileName), zipErr)
					Abort(c, http.StatusInternalServerError, i18n.ErrZipFailed)
					return
				}

				log.Infof("album: zipped %s as %s", clean.Log(file.FileName), clean.Log(alias))
			} else {
				log.Warnf("album: %s not found", clean.Log(file.FileName))
			}
		}

		log.Infof("album: %s has been downloaded [%s]", clean.Log(a.AlbumTitle), time.Since(start))
	})
}
