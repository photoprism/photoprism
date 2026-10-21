package form

import (
	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	// RefID indicates a reference token identifier.
	RefID = "ref_id"
	// SessionID indicates a session token identifier.
	SessionID = "session_id"
	// AccessToken indicates a bearer access token identifier.
	AccessToken = "access_token"
	// RefreshToken indicates a refresh token of an app grant.
	RefreshToken = "refresh_token"
)

// grantUIDPrefix is the UID prefix of app grants, whose UID starts their refresh tokens.
const grantUIDPrefix = byte('g')

// OAuthRevokeToken represents a token revocation form.
type OAuthRevokeToken struct {
	Token         string `form:"token" binding:"required" json:"token,omitempty"`
	TokenTypeHint string `form:"token_type_hint" json:"token_type_hint,omitempty"`
}

// Empty checks if all form values are unset.
func (f *OAuthRevokeToken) Empty() bool {
	switch {
	case f.Token != "":
		return false
	case f.TokenTypeHint != "":
		return false
	}

	return true
}

// Validate checks the revoke token form values and returns an error if invalid. A hint that does not match
// the token is ignored, and the type is then inferred from the token (RFC 7009, section 2.1).
func (f *OAuthRevokeToken) Validate() error {
	// Require a token.
	if f.Token == "" {
		return authn.ErrTokenRequired
	}

	// Validate token type.
	isRefID := rnd.IsRefID(f.Token)
	isSessionID := rnd.IsSessionID(f.Token)
	isAuthAny := rnd.IsAuthAny(f.Token)
	isRefreshToken := isRefreshTokenShape(f.Token)

	switch f.TokenTypeHint {
	case RefID:
		if isRefID {
			return nil
		}
	case SessionID:
		if isSessionID {
			return nil
		}
	case AccessToken:
		if isAuthAny {
			return nil
		}
	case RefreshToken:
		if isRefreshToken {
			return nil
		}
	case "":
	default:
		return authn.ErrInvalidTokenType
	}

	switch {
	case isRefreshToken:
		f.TokenTypeHint = RefreshToken
	case isRefID:
		f.TokenTypeHint = RefID
	case isSessionID:
		f.TokenTypeHint = SessionID
	case isAuthAny:
		f.TokenTypeHint = AccessToken
	default:
		return authn.ErrInvalidToken
	}

	return nil
}

// isRefreshTokenShape reports whether s has the shape of a grant refresh token, with or without an expiry
// and signature.
func isRefreshTokenShape(s string) bool {
	if value, _, _, ok := tokens.SplitValue(s); ok {
		s = value
	}

	_, _, ok := rnd.ParseRefreshToken(s, grantUIDPrefix)

	return ok
}
