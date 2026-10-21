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
// The name is written unquoted, so it must be a server-generated ASCII name.
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
// without a preview token of its own receives neither, since the download token also authorizes originals,
// except a client session without an account, which receives the download token its scope permits.
func AddTokenHeaders(c *gin.Context, s *entity.Session) {
	if s == nil || get.Config().Public() {
		return
	}

	preview := SessionPreviewToken(s)

	if preview != "" {
		c.Header("X-Preview-Token", preview)
	} else if !s.SignedPreview() {
		return
	}

	if v := SessionDownloadToken(s); v != "" {
		c.Header("X-Download-Token", v)
	}
}

// SessionPreviewToken returns the preview token to deliver to the session: for a client session without
// an account, a signed token if its scope and role cover previews and none otherwise; else its stored token.
func SessionPreviewToken(s *entity.Session) string {
	switch {
	case s == nil:
		return ""
	case !s.SignedPreview():
		return s.PreviewToken
	case !s.ScopePermitsPreview() || !s.RolePermitsPreview():
		return ""
	}

	return tokens.PreviewToken(s.ID, s.ExpiresAt())
}

// SessionContentUri returns the base URI of the session's thumbnail URLs, which is the API for an OAuth2
// client without an account, as in Config.ClientOAuth.
func SessionContentUri(s *entity.Session) string {
	if s != nil && s.NoUser() && s.IsClient() {
		return get.Config().ApiUri()
	}

	return get.Config().ContentUri()
}

// SessionDownloadToken returns the "?t=" value a client appends to a download URL: a signed,
// session-bound token so header-less endpoints resolve back to this session. It is empty when the
// session's scope covers no downloadable resource; the endpoints apply their own scope check.
func SessionDownloadToken(s *entity.Session) string {
	if s == nil || !s.ScopePermitsDownload() {
		return ""
	}

	return tokens.DownloadToken(s.ID)
}
