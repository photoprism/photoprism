package entity

import (
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TokenConfig is the reserved key under which the instance-wide default tokens are registered.
const TokenConfig = "__config__" //nolint:gosec // G101: Reserved token keyword, not a credential.

// TokenPublic is the token value that grants access while the instance runs in public mode.
const TokenPublic = "public"

// PreviewToken maps each active session ID to its preview token value for lookup.
var PreviewToken = NewStringMap(Strings{})

// DownloadToken maps each active session ID to its download token value for lookup.
var DownloadToken = NewStringMap(Strings{})

// ValidateTokens enables preview and download token validation; it is disabled in public mode.
var ValidateTokens = true

// GenerateToken returns a short random token for previews or downloads.
func GenerateToken() string {
	return rnd.Base36(8)
}

// InvalidDownloadToken checks if the token is unknown.
func InvalidDownloadToken(t string) bool {
	return ValidateTokens && DownloadToken.MissingValue(t)
}

// InvalidPreviewToken checks if the preview token is unknown.
func InvalidPreviewToken(t string) bool {
	return ValidateTokens && PreviewToken.MissingValue(t) && DownloadToken.MissingValue(t)
}
