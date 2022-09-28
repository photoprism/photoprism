package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

// InvalidPreviewToken checks if the token found in the request is valid for image thumbnails and video streams.
func InvalidPreviewToken(c *gin.Context) bool {
	token := clean.UrlToken(c.Param("token"))

	if token == "" {
		token = clean.UrlToken(c.Query("t"))
	}

	return service.Config().InvalidPreviewToken(token)
}

// InvalidDownloadToken checks if the token found in the request is valid for file downloads.
func InvalidDownloadToken(c *gin.Context) bool {
	return service.Config().InvalidDownloadToken(clean.UrlToken(c.Query("t")))
}
