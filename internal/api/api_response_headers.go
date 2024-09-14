package api

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/header"
)

// AddCountHeader adds the actual result count to the response.
func AddCountHeader(c *gin.Context, count int) {
	c.Header("X-Count", strconv.Itoa(count))
}

// AddLimitHeader adds the max result count to the response.
func AddLimitHeader(c *gin.Context, limit int) {
	c.Header("X-Limit", strconv.Itoa(limit))
}

// AddOffsetHeader adds the result offset to the response.
func AddOffsetHeader(c *gin.Context, offset int) {
	c.Header("X-Offset", strconv.Itoa(offset))
}

// AddDownloadHeader adds a header indicating the response is expected to be downloaded.
func AddDownloadHeader(c *gin.Context, fileName string) {
	c.Header(header.ContentDisposition, fmt.Sprintf("attachment; filename=%s", fileName))
}

// AddAuthTokenHeader adds an X-Auth-Token header to the response.
func AddAuthTokenHeader(c *gin.Context, authToken string) {
	c.Header(header.XAuthToken, authToken)
}

// AddContentTypeHeader adds a content type header to the response.
func AddContentTypeHeader(c *gin.Context, contentType string) {
	c.Header(header.ContentType, contentType)
}

// AddFileCountHeaders adds file and folder counts to the response.
func AddFileCountHeaders(c *gin.Context, filesCount, foldersCount int) {
	c.Header("X-Files", strconv.Itoa(filesCount))
	c.Header("X-Folders", strconv.Itoa(foldersCount))
}

// AddTokenHeaders adds preview token headers to the response.
func AddTokenHeaders(c *gin.Context, s *entity.Session) {
	if get.Config().Public() {
		return
	}

	if s.PreviewToken != "" {
		c.Header("X-Preview-Token", s.PreviewToken)
	}
	if s.DownloadToken != "" {
		c.Header("X-Download-Token", s.DownloadToken)
	}
}
