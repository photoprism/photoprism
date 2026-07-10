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
// served type from the request name's extension (e.g. "file.pdf" or "file.svg"). The
// auth and visibility preamble matches GetFile; an unknown or unsupported type returns 415.
//
//	@Summary	streams the original file bytes for inline viewing (e.g. ".pdf" or ".svg")
//	@Id			GetFileBytes
//	@Tags		Files
//	@Produce	application/pdf
//	@Produce	image/svg+xml
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

		// Select the served type from the request name's extension; an unknown or
		// unsupported extension has no descriptor and returns 415.
		spec, ok := servableTypes[fs.FileType(c.Param("name"))]
		if !ok {
			Abort(c, http.StatusUnsupportedMediaType, i18n.ErrUnsupportedType)
			return
		}

		// Resolve the file actually served. It must belong to the SAME photo that passed the
		// visibility check above, so a resolver can never hand back a file past the scope gate.
		target, ok := spec.Resolve(f)
		if !ok || target.PhotoUID != f.PhotoUID {
			AbortEntityNotFound(c)
			return
		}

		serveInlineFile(c, spec, target)
	})
}

// ServableType describes how GetFileBytes streams a file type inline. Fields are
// exported so additional served types can be registered from the same package.
type ServableType struct {
	ContentType string               // response Content-Type
	Disposition string               // response Content-Disposition; defaults to "inline"
	SetHeaders  func(c *gin.Context) // optional per-type response headers (e.g. SVG XSS neutralization)
	// Resolve maps the request file to the file actually served. It must return a
	// file of the SAME photo that passed the visibility check, which GetFileBytes enforces.
	Resolve func(f *entity.File) (*entity.File, bool)
}

// servableTypes lists the file types GetFileBytes streams inline, keyed by fs.Type.
// Adding a type is a map entry, not a new handler branch.
var servableTypes = map[fs.Type]ServableType{
	fs.DocumentPDF: {ContentType: header.ContentTypePDF, Resolve: resolvePDFDocument},
	fs.VectorSVG:   {ContentType: header.ContentTypeSVG, SetHeaders: svgSafeHeaders, Resolve: serveSelf(fs.VectorSVG)},
}

// serveSelf returns a resolver that serves the request file itself when its type
// matches want, used by types whose original is its own primary file.
func serveSelf(want fs.Type) func(*entity.File) (*entity.File, bool) {
	return func(f *entity.File) (*entity.File, bool) {
		if f.Type() == want {
			return f, true
		}
		return nil, false
	}
}

// resolvePDFDocument maps a document request to its PDF original. A document's
// primary file is its rendered cover image, so when the request carries the cover
// hash the related PDF for the same photo is resolved. This cross-resolution is
// deliberate and specific to documents: the original IS the sidecar.
func resolvePDFDocument(f *entity.File) (*entity.File, bool) {
	if f.Type() == fs.DocumentPDF {
		return f, true
	}

	if doc, err := query.DocumentByPhotoUID(f.PhotoUID); err == nil && doc.Type() == fs.DocumentPDF {
		return doc, true
	}

	return nil, false
}

// svgSafeHeaders neutralizes the XSS surface of an inline SVG served same-origin:
// a restrictive CSP sandboxes the document and blocks script execution, and nosniff
// stops content-type sniffing. PDF needs none of this, hence the per-type hook.
func svgSafeHeaders(c *gin.Context) {
	c.Header(header.ContentSecurityPolicy, "default-src 'none'; style-src 'unsafe-inline'; sandbox")
	c.Header(header.ContentTypeOptions, header.PolicyNoSniff)
}

// serveInlineFile streams the original bytes of f inline per the type descriptor,
// with HTTP Range support via c.File. A missing original is flagged so it drops
// out of search results and reported as not found.
func serveInlineFile(c *gin.Context, spec ServableType, f *entity.File) {
	// Resolve the absolute filename and verify the original still exists.
	fileName := photoprism.FileName(f.FileRoot, f.FileName)

	if !fs.FileExists(fileName) {
		log.Errorf("files: %s %s is missing", f.Type().String(), clean.Log(f.FileName))

		// Flag as missing so it no longer shows up in search results.
		logErr("files", f.Update("FileMissing", true))

		AbortEntityNotFound(c)
		return
	}

	// The response is session-scoped, so cache it privately and never on shared CDNs.
	AddContentTypeHeader(c, spec.ContentType)
	header.SetCacheControlImmutable(c, ttl.CacheDefault.Int(), false)

	// Apply any per-type response headers (e.g. SVG XSS neutralization).
	if spec.SetHeaders != nil {
		spec.SetHeaders(c)
	}

	// Serve the file inline with HTTP Range support via c.File (http.ServeContent),
	// which keeps the Content-Type and Content-Disposition set here.
	disposition := spec.Disposition
	if disposition == "" {
		disposition = "inline"
	}

	c.Header(header.ContentDisposition, disposition)
	c.File(fileName)
}
