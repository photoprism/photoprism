package form

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// OAuthCreateToken represents a create token request form.
type OAuthCreateToken struct {
	GrantType    authn.GrantType `form:"grant_type" json:"grant_type,omitempty"`
	ClientID     string          `form:"client_id" json:"client_id,omitempty"`
	ClientName   string          `form:"client_name" json:"client_name,omitempty"`
	ClientSecret string          `form:"client_secret" json:" client_secret,omitempty"`
	Username     string          `form:"username" json:"username,omitempty"`
	Password     string          `form:"password" json:"password,omitempty"`
	RefreshToken string          `form:"refresh_token" json:"refresh_token,omitempty"`
	Code         string          `form:"code" json:"code,omitempty"`
	CodeVerifier string          `form:"code_verifier" json:"code_verifier,omitempty"`
	RedirectURI  string          `form:"redirect_uri" json:"redirect_uri,omitempty"`
	Assertion    string          `form:"assertion" json:"assertion,omitempty"`
	Scope        string          `form:"scope" json:"scope,omitempty"`
	Lifetime     int64           `form:"lifetime" json:"lifetime,omitempty"`
}

// Validate verifies the request parameters depending on the grant type.
func (f OAuthCreateToken) Validate() error {
	switch f.GrantType {
	case authn.GrantClientCredentials, authn.GrantUndefined:
		// Validate client id.
		if f.ClientID == "" {
			return authn.ErrClientIDRequired
		} else if rnd.InvalidUID(f.ClientID, 'c') {
			return authn.ErrInvalidCredentials
		}

		// Validate client secret.
		if f.ClientSecret == "" {
			return authn.ErrClientSecretRequired
		} else if !rnd.IsAlnum(f.ClientSecret) {
			return authn.ErrInvalidCredentials
		}
	case authn.GrantPassword:
		// Validate request credentials.
		if f.Username == "" {
			return authn.ErrUsernameRequired
		} else if len(f.Username) > txt.ClipUsername {
			return authn.ErrInvalidCredentials
		} else if f.Password == "" {
			return authn.ErrPasswordRequired
		} else if len(f.Password) > txt.ClipPassword {
			return authn.ErrInvalidCredentials
		} else if f.ClientName == "" {
			return authn.ErrNameRequired
		} else if f.Scope == "" {
			return authn.ErrScopeRequired
		}
	default:
		// Reject requests with unsupported grant types.
		return authn.ErrInvalidGrantType
	}

	return nil
}

// CleanScope returns the client scopes as sanitized string.
func (f OAuthCreateToken) CleanScope() string {
	return clean.Scope(f.Scope)
}
