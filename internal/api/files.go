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
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// GetFile returns file details as JSON.
//
//	@Summary	returns file details as JSON
//	@Id			GetFile
//	@Tags		Files
//	@Produce	json
//	@Success	200				{object}	entity.File
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		hash			path		string	true	"SHA-1 hash of the file"
//	@Router		/api/v1/files/{hash} [get]
func GetFile(router *gin.RouterGroup) {
	router.GET("/files/:hash", func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.ActionView)

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		hash := clean.Token(c.Param("hash"))

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

		c.JSON(http.StatusOK, f)
	})
}

// GetFileBytes streams the original bytes of a file for inline viewing, selecting the
// served type from the request name's extension (e.g. "file.pdf"). The auth and
// visibility preamble matches GetFile; an unknown or unsupported type returns 415.
//
//	@Summary	streams the original file bytes for inline viewing (e.g. ".pdf")
//	@Id			GetFileBytes
//	@Tags		Files
//	@Produce	application/pdf
//	@Success	200					{file}		binary
//	@Failure	401,403,404,415,429	{object}	i18n.Response
//	@Param		hash				path		string	true	"SHA-1 hash of the file"
//	@Param		name				path		string	true	"file name whose extension selects the served type (e.g. 'file.pdf')"
//	@Router		/api/v1/files/{hash}/{name} [get]
func GetFileBytes(router *gin.RouterGroup) {
	router.GET("/files/:hash/:name", func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.ActionView)

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		hash := clean.Token(c.Param("hash"))

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

		// Serve the bytes based on the type selected by the request name's extension.
		// Future variants (e.g. ".jpg"/".mp4") add their own case here.
		switch fs.FileType(c.Param("name")) {
		case fs.DocumentPDF:
			getFilePDF(c, f)
		default:
			Abort(c, http.StatusUnsupportedMediaType, i18n.ErrUnsupportedType)
		}
	})
}

// getFilePDF streams the original PDF bytes of a document for inline viewing,
// reusing the session-scoped authorization already enforced by the caller.
func getFilePDF(c *gin.Context, f *entity.File) {
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

	// Serve the document inline with HTTP Range support via c.File (http.ServeContent),
	// which keeps the Content-Type and Content-Disposition set here.
	c.Header(header.ContentDisposition, "inline")
	c.File(fileName)
}
