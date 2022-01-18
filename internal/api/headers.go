package api

import (
	"fmt"
	"strconv"

	"github.com/photoprism/photoprism/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	ContentTypeAvc = `video/mp4; codecs="avc1"`
)

// AddCacheHeader adds a cache control header to the response.
func AddCacheHeader(c *gin.Context, maxAge MaxAge) {
	c.Header("Cache-Control", fmt.Sprintf("private, max-age=%s, no-transform", maxAge.String()))
}

// AddCoverCacheHeader adds cover image cache control headers to the response.
func AddCoverCacheHeader(c *gin.Context) {
	AddCacheHeader(c, CoverCacheTTL)
}

// AddThumbCacheHeader adds thumbnail cache control headers to the response.
func AddThumbCacheHeader(c *gin.Context) {
	c.Header("Cache-Control", fmt.Sprintf("private, max-age=%s, no-transform, immutable", ThumbCacheTTL.String()))
}

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
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
}

// AddSessionHeader adds a session id header to the response.
func AddSessionHeader(c *gin.Context, id string) {
	c.Header("X-Session-ID", id)
}

// AddContentTypeHeader adds a content type header to the response.
func AddContentTypeHeader(c *gin.Context, contentType string) {
	c.Header("Content-Type", contentType)
}

// AddFileCountHeaders adds file and folder counts to the response.
func AddFileCountHeaders(c *gin.Context, filesCount, foldersCount int) {
	c.Header("X-Files", strconv.Itoa(filesCount))
	c.Header("X-Folders", strconv.Itoa(foldersCount))
}

// AddTokenHeaders adds preview token headers to the response.
func AddTokenHeaders(c *gin.Context) {
	c.Header("X-Preview-Token", service.Config().PreviewToken())
	c.Header("X-Download-Token", service.Config().DownloadToken())
}
