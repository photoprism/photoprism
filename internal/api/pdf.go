package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// GetPDF streams the original bytes of a PDF document for inline viewing.
//
//	@Summary	streams a PDF document for inline viewing
//	@Id			GetPDF
//	@Produce	application/pdf
//	@Tags		Files, Documents
//	@Failure	403,404,415	{object}	i18n.Response
//	@Success	200			{file}		application/pdf
//	@Param		hash		path		string	true	"SHA1 file hash"
//	@Param		token		path		string	true	"user-specific security token provided with session or 'public' when running PhotoPrism in public mode"
//	@Router		/api/v1/pdf/{hash}/{token} [get]
func GetPDF(router *gin.RouterGroup) {
	router.GET("/pdf/:hash/:token", func(c *gin.Context) {
		// A valid preview token is required. This is independent of the share's
		// download permission, so share-link visitors can view PDFs even when
		// downloads are disabled, consistent with thumbnails and video streams.
		if InvalidPreviewToken(c) {
			AbortForbidden(c)
			return
		}

		// Find the indexed file by its hash.
		fileHash := clean.Token(c.Param("hash"))

		f, err := query.FileByHash(fileHash)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		// A document's primary file is its rendered cover image, so when the
		// request uses the cover hash, resolve the related PDF for the photo.
		if f.Type() != fs.DocumentPDF {
			if doc, docErr := query.DocumentByPhotoUID(f.PhotoUID); docErr == nil && doc.Type() == fs.DocumentPDF {
				f = doc
			} else {
				Abort(c, http.StatusUnsupportedMediaType, i18n.ErrUnsupportedType)
				return
			}
		}

		// Resolve the absolute filename and verify the original still exists.
		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		if !fs.FileExists(fileName) {
			log.Errorf("pdf: file %s is missing", clean.Log(f.FileName))

			// Flag as missing so it no longer shows up in search results.
			logErr("pdf", f.Update("FileMissing", true))

			AbortEntityNotFound(c)
			return
		}

		// Serve the document with an immutable cache header. The c.File helper
		// uses http.ServeContent, which adds HTTP Range support for progressive
		// loading and keeps the Content-Type and Content-Disposition set here.
		AddContentTypeHeader(c, header.ContentTypePDF)
		AddImmutableCacheHeader(c)

		// Allow forcing a download via ?download for consistency with other
		// file endpoints; the default is inline viewing.
		if c.Query("download") != "" {
			c.FileAttachment(fileName, f.DownloadName(DownloadName(c), 0))
			return
		}

		c.Header(header.ContentDisposition, "inline")
		c.File(fileName)
	})
}
