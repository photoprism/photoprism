package oauth

import (
	"errors"
)

// OAuth error codes (RFC 6749, RFC 8707).
const (
	ErrCodeInvalidRequest         = "invalid_request"
	ErrCodeInvalidClient          = "invalid_client"
	ErrCodeInvalidGrant           = "invalid_grant"
	ErrCodeInvalidScope           = "invalid_scope"
	ErrCodeInvalidTarget          = "invalid_target"
	ErrCodeAccessDenied           = "access_denied"
	ErrCodeUnauthorizedClient     = "unauthorized_client"
	ErrCodeUnsupportedResponse    = "unsupported_response_type"
	ErrCodeServerError            = "server_error"
	ErrCodeTemporarilyUnavailable = "temporarily_unavailable"
)

// ErrInvalidGrant is returned when an authorization code or refresh token cannot be redeemed.
var ErrInvalidGrant = errors.New("invalid grant")

// ErrSigningKey is returned when a code or refresh token cannot be signed because no signing key is configured.
var ErrSigningKey = errors.New("oauth: token signing key unavailable")
