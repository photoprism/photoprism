package oauth

import (
	"errors"
	"fmt"
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

// ErrInvalidToken is returned for a code or refresh token with an invalid shape or signature.
var ErrInvalidToken = fmt.Errorf("%w: invalid token", ErrInvalidGrant)

// ErrTokenExpired is returned for a code or refresh token with a valid signature whose signed expiry has passed.
var ErrTokenExpired = fmt.Errorf("%w: token expired", ErrInvalidGrant)

// ErrGrantNotFound is returned for a valid refresh token whose grant does not exist.
var ErrGrantNotFound = fmt.Errorf("%w: grant not found", ErrInvalidGrant)

// ErrSigningKey is returned when a code or refresh token cannot be signed because no signing key is configured.
var ErrSigningKey = errors.New("oauth: token signing key unavailable")
