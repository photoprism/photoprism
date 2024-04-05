package form

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	ClientAccessToken = "access_token"
)

// OAuthRevokeToken represents a token revokation form.
type OAuthRevokeToken struct {
	AuthToken string `form:"token" binding:"required" json:"token,omitempty"`
	TypeHint  string `form:"token_type_hint" json:" token_type_hint,omitempty"`
}

// Empty checks if all form values are unset.
func (f OAuthRevokeToken) Empty() bool {
	switch {
	case f.AuthToken != "":
		return false
	case f.TypeHint != "":
		return false
	}

	return true
}

// Validate checks the token and token type.
func (f OAuthRevokeToken) Validate() error {
	// Check auth token.
	if f.AuthToken == "" {
		return fmt.Errorf("missing token")
	} else if !rnd.IsAlnum(f.AuthToken) {
		return fmt.Errorf("invalid token")
	}

	// Check token type.
	if f.TypeHint != "" && f.TypeHint != ClientAccessToken {
		return fmt.Errorf("unsupported token type")
	}

	return nil
}
