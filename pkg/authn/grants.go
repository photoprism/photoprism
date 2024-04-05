package authn

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GrantType represents an authentication grant type.
type GrantType string

// Standard authentication grant types.
const (
	GrantUndefined         GrantType = ""
	GrantClientCredentials GrantType = "client_credentials"
	GrantPassword          GrantType = "password"
	GrantShareToken        GrantType = "share_token"
	GrantRefreshToken      GrantType = "refresh_token"
	GrantAuthorizationCode GrantType = "authorization_code"
	GrantJwtBearer         GrantType = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	GrantSamlBearer        GrantType = "urn:ietf:params:oauth:grant-type:saml2-bearer"
	GrantTokenExchange     GrantType = "urn:ietf:params:oauth:grant-type:token-exchange"
)

// String returns the provider identifier as a string.
func (t GrantType) String() string {
	return clean.TypeLowerUnderscore(string(t))
}

// Is compares the method with another type.
func (t GrantType) Is(method GrantType) bool {
	return t == method
}

// IsNot checks if the method is not the specified type.
func (t GrantType) IsNot(method GrantType) bool {
	return t != method
}

// IsUndefined checks if the method is undefined.
func (t GrantType) IsUndefined() bool {
	return t == ""
}

// Equal checks if the type matches.
func (t GrantType) Equal(s string) bool {
	return strings.EqualFold(s, t.String())
}

// NotEqual checks if the type is different.
func (t GrantType) NotEqual(s string) bool {
	return !t.Equal(s)
}

// Pretty returns the provider identifier in an easy-to-read format.
func (t GrantType) Pretty() string {
	switch t {
	case GrantShareToken:
		return "Share Token"
	case GrantRefreshToken:
		return "Refresh Token"
	case GrantClientCredentials:
		return "Client Credentials"
	case GrantAuthorizationCode:
		return "Authorization Code"
	case GrantJwtBearer:
		return "JWT Bearer Assertion"
	case GrantSamlBearer:
		return "SAML2 Bearer Assertion"
	case GrantTokenExchange:
		return "Token Exchange"
	default:
		return txt.UpperFirst(t.String())
	}
}

// Grant casts a string to a normalized grant type.
func Grant(s string) GrantType {
	s = clean.TypeLowerUnderscore(s)
	switch s {
	case "", "-", "null", "nil", "0", "false":
		return GrantUndefined
	case "client_credentials", "client":
		return GrantClientCredentials
	case "password", "passwd", "pass", "user", "username":
		return GrantPassword
	case "share_token", "share":
		return GrantClientCredentials
	case "refresh_token", "refresh":
		return GrantRefreshToken
	case "authorization_code", "auth_code":
		return GrantAuthorizationCode
	case "jwt-bearer", "jwt", "jwt_bearer":
		return GrantJwtBearer
	case "saml2-bearer", "saml2_bearer", "saml2", "saml":
		return GrantSamlBearer
	case "token-exchange", "token_exchange":
		return GrantTokenExchange
	default:
		return GrantType(s)
	}
}
