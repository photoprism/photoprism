package api

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/http/header"
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

// AddContentTypeHeader adds a "Content-Type" header to the response.
func AddContentTypeHeader(c *gin.Context, contentType string) {
	c.Header(header.ContentType, contentType)
}

// AddFileCountHeaders adds file and folder counts to the response.
func AddFileCountHeaders(c *gin.Context, filesCount, foldersCount int) {
	c.Header("X-Files", strconv.Itoa(filesCount))
	c.Header("X-Folders", strconv.Itoa(foldersCount))
}

// AddTokenHeaders adds the preview and download tokens to the response so the client can refresh them
// while browsing instead of polling. Both mirror what ClientSession puts in the client config: a session
// without its own preview token receives neither, since the download token also authorizes originals.
func AddTokenHeaders(c *gin.Context, s *entity.Session) {
	if get.Config().Public() || s.PreviewToken == "" {
		return
	}

	c.Header("X-Preview-Token", s.PreviewToken)

	// The download token is the "?t=" value the client appends to a download URL: a signed,
	// session-bound token so header-less endpoints resolve back to this session.
	if v := tokens.DownloadToken(s.ID); v != "" {
		c.Header("X-Download-Token", v)
	}
}
