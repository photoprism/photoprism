package form

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	RefID       = "ref_id"
	SessionID   = "session_id"
	AccessToken = "access_token"
)

// OAuthRevokeToken represents a token revokation form.
type OAuthRevokeToken struct {
	Token         string `form:"token" binding:"required" json:"token,omitempty"`
	TokenTypeHint string `form:"token_type_hint" json:" token_type_hint,omitempty"`
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

// Validate checks the revoke token form values and returns an error if invalid.
func (f *OAuthRevokeToken) Validate() error {
	// Require a token.
	if f.Token == "" {
		return authn.ErrTokenRequired
	}

	// Validate token type.
	isRefID := rnd.IsRefID(f.Token)
	isSessionID := rnd.IsSessionID(f.Token)
	isAuthAny := rnd.IsAuthAny(f.Token)

	switch f.TokenTypeHint {
	case "":
		if !isRefID && !isSessionID && !isAuthAny {
			return authn.ErrInvalidToken
		} else if isRefID {
			f.TokenTypeHint = RefID
		} else if isSessionID {
			f.TokenTypeHint = SessionID
		} else {
			f.TokenTypeHint = AccessToken
		}
	case RefID:
		if !isRefID {
			return authn.ErrInvalidToken
		}
	case SessionID:
		if !isSessionID {
			return authn.ErrInvalidToken
		}
	case AccessToken:
		if !isAuthAny {
			return authn.ErrInvalidToken
		}
	default:
		return authn.ErrInvalidTokenType
	}

	return nil
}
