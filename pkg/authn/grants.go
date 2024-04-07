package authn

import (
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GrantType represents an authentication grant type.
type GrantType string

// Standard authentication grant types.
const (
	GrantUndefined         GrantType = ""
	GrantCLI               GrantType = "cli"
	GrantImplicit          GrantType = "implicit"
	GrantSession           GrantType = "session"
	GrantPassword          GrantType = "password"
	GrantClientCredentials GrantType = "client_credentials"
	GrantShareToken        GrantType = "share_token"
	GrantRefreshToken      GrantType = "refresh_token"
	GrantAuthorizationCode GrantType = "authorization_code"
	GrantJwtBearer         GrantType = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	GrantSamlBearer        GrantType = "urn:ietf:params:oauth:grant-type:saml2-bearer"
	GrantTokenExchange     GrantType = "urn:ietf:params:oauth:grant-type:token-exchange"
)

// Grant casts a string to a normalized grant type.
func Grant(s string) GrantType {
	s = clean.TypeLowerUnderscore(s)
	switch s {
	case "", "_", "-", "null", "nil", "0", "false":
		return GrantUndefined
	case "cli", "terminal", "command":
		return GrantCLI
	case "implicit":
		return GrantImplicit
	case "session":
		return GrantSession
	case "password", "passwd", "pass":
		return GrantPassword
	case "client_credentials", "client":
		return GrantClientCredentials
	case "share_token", "share":
		return GrantShareToken
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

// Pretty returns the grant type in a human-readable format.
func (t GrantType) Pretty() string {
	switch t {
	case GrantCLI:
		return "CLI"
	case GrantImplicit:
		return "Implicit"
	case GrantSession:
		return "Session"
	case GrantPassword:
		return "Password"
	case GrantClientCredentials:
		return "Client Credentials"
	case GrantShareToken:
		return "Share Token"
	case GrantRefreshToken:
		return "Refresh Token"
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

// String returns the grant type as a string.
func (t GrantType) String() string {
	return clean.TypeLowerUnderscore(string(t))
}

// Equal checks if the type matches the specified string.
func (t GrantType) Equal(s string) bool {
	return t == Grant(s)
}

// NotEqual checks if the type does mot match the specified string.
func (t GrantType) NotEqual(s string) bool {
	return !t.Equal(s)
}

// Is compares the grant with another type.
func (t GrantType) Is(grantType GrantType) bool {
	return t == grantType
}

// IsNot checks if the grant is not the specified type.
func (t GrantType) IsNot(grantType GrantType) bool {
	return t != grantType
}

// IsUndefined checks if the grant is undefined.
func (t GrantType) IsUndefined() bool {
	return t == ""
}
