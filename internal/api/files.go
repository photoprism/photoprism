package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// GetFile returns file details as JSON, or streams the document bytes when the
// request adds a known file extension (e.g. ".pdf" for the inline PDF viewer).
//
//	@Summary	returns file details as JSON, or the document bytes for a ".pdf" request
//	@Id			GetFile
//	@Tags		Files
//	@Produce	json
//	@Produce	application/pdf
//	@Success	200					{object}	entity.File
//	@Failure	401,403,404,415,429	{object}	i18n.Response
//	@Param		hash				path		string	true	"SHA-1 hash of the file, optionally suffixed with '.pdf' to stream the document"
//	@Router		/api/v1/files/{hash} [get]
func GetFile(router *gin.RouterGroup) {
	router.GET("/files/:hash", func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.ActionView)

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		// Parse the file extension off the raw param before cleaning, since clean.Token
		// strips dots. A ".pdf" suffix streams the document bytes instead of the JSON
		// details; no/unknown extension keeps the original, fully back-compatible response.
		param := c.Param("hash")
		ext := fs.FileType(param)
		hash := clean.Token(fs.StripExt(param))

		// Limit results to files within the session's shared scope, consistent with how photo
		// search filters results. Files outside the scope are reported as not found.
		if visible, err := search.FileVisibleToSession(hash, s); err != nil || !visible {
			AbortEntityNotFound(c)
			return
		}

		f, err := query.FileByHash(hash)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		// Stream the original PDF bytes for a ".pdf" request (used by the inline viewer).
		if ext == fs.DocumentPDF {
			getFilePDF(c, s, f)
			return
		}

		c.JSON(http.StatusOK, f)
	})
}

// getFilePDF streams the original PDF bytes of a document for inline viewing,
// reusing the session-scoped authorization already enforced by GetFile.
func getFilePDF(c *gin.Context, s *entity.Session, f *entity.File) {
	// A document's primary file is its rendered cover image, so when the request
	// carries the cover hash, resolve the related PDF for the same photo. This is
	// deliberate and specific to documents: the original IS the sidecar. A future
	// ".jpg"/".mp4" variant should serve its own type directly, not cross-resolve.
	if f.Type() != fs.DocumentPDF {
		if doc, err := query.DocumentByPhotoUID(f.PhotoUID); err == nil && doc.Type() == fs.DocumentPDF {
			f = doc
		} else {
			Abort(c, http.StatusUnsupportedMediaType, i18n.ErrUnsupportedType)
			return
		}
	}

	// Resolve the absolute filename and verify the original still exists.
	fileName := photoprism.FileName(f.FileRoot, f.FileName)

	if !fs.FileExists(fileName) {
		log.Errorf("files: pdf %s is missing", clean.Log(f.FileName))

		// Flag as missing so it no longer shows up in search results.
		logErr("files", f.Update("FileMissing", true))

		AbortEntityNotFound(c)
		return
	}

	// The response is session-scoped, so cache it privately and never on shared CDNs.
	AddContentTypeHeader(c, header.ContentTypePDF)
	header.SetCacheControlImmutable(c, ttl.CacheDefault.Int(), false)

	// Forcing a download is an actual download, so honor the download permission and the
	// "downloads disabled" setting even though inline viewing only requires ActionView.
	// Mirrors the gate used by the zip and album-download endpoints (ResourcePhotos + ActionDownload).
	if c.Query("download") != "" {
		conf := get.Config()

		if !acl.Rules.Allow(acl.ResourcePhotos, s.GetUserRole(), acl.ActionDownload) ||
			!conf.Settings().Features.Download || conf.Settings().Download.Disabled {
			AbortFeatureDisabled(c)
			return
		}

		c.FileAttachment(fileName, f.DownloadName(DownloadName(c), 0))
		return
	}

	// Serve the document inline with HTTP Range support via c.File (http.ServeContent),
	// which keeps the Content-Type and Content-Disposition set here.
	c.Header(header.ContentDisposition, "inline")
	c.File(fileName)
}
