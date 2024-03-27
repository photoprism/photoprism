package entity

import (
	"github.com/photoprism/photoprism/pkg/rnd"
)

const TokenConfig = "__config__"
const TokenPublic = "public"

var PreviewToken = NewStringMap(Strings{})
var DownloadToken = NewStringMap(Strings{})
var CheckTokens = true

// GenerateToken returns a random string token.
func GenerateToken() string {
	return rnd.Base36(8)
}

// InvalidDownloadToken checks if the token is unknown.
func InvalidDownloadToken(t string) bool {
	return CheckTokens && DownloadToken.Missing(t)
}

// InvalidPreviewToken checks if the preview token is unknown.
func InvalidPreviewToken(t string) bool {
	return CheckTokens && PreviewToken.Missing(t) && DownloadToken.Missing(t)
}
