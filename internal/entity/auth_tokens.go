package entity

import (
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Reserved token names used for configuration and public access.
const TokenConfig = "__config__"
const TokenPublic = "public"

var PreviewToken = NewStringMap(Strings{})
var DownloadToken = NewStringMap(Strings{})
var ValidateTokens = true

// GenerateToken returns a short random token for previews or downloads.
func GenerateToken() string {
	return rnd.Base36(8)
}

// InvalidDownloadToken checks if the token is unknown.
func InvalidDownloadToken(t string) bool {
	return ValidateTokens && DownloadToken.Missing(t)
}

// InvalidPreviewToken checks if the preview token is unknown.
func InvalidPreviewToken(t string) bool {
	return ValidateTokens && PreviewToken.Missing(t) && DownloadToken.Missing(t)
}
