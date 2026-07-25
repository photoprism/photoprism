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

// ValidateTokens enables preview token validation; it is disabled in public mode.
var ValidateTokens = true

// GenerateToken returns a short random token for previews.
func GenerateToken() string {
	return rnd.Base36(8)
}

// InvalidPreviewToken checks if the preview token is unknown. Download-token cross-acceptance is applied
// by the request handler (api.InvalidPreviewToken), which also accepts the coarse download token.
func InvalidPreviewToken(t string) bool {
	return ValidateTokens && PreviewToken.MissingValue(t)
}
