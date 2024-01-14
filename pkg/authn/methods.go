package authn

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// MethodType represents an authentication method.
type MethodType string

// Authentication methods.
const (
	MethodDefault     MethodType = "default"
	MethodAccessToken MethodType = "access_token"
	MethodOAuth2      MethodType = "oauth2"
	MethodOIDC        MethodType = "oidc"
	Method2FA         MethodType = "2fa"
	MethodUnknown     MethodType = ""
)

// IsDefault checks if this is the default method.
func (t MethodType) IsDefault() bool {
	return t.String() == MethodDefault.String()
}

// String returns the provider identifier as a string.
func (t MethodType) String() string {
	switch t {
	case "":
		return string(MethodDefault)
	case "oauth":
		return string(MethodOAuth2)
	case "openid":
		return string(MethodOIDC)
	case "totp":
		return string(Method2FA)
	default:
		return string(t)
	}
}

// Equal checks if the type matches.
func (t MethodType) Equal(s string) bool {
	return strings.EqualFold(s, t.String())
}

// NotEqual checks if the type is different.
func (t MethodType) NotEqual(s string) bool {
	return !t.Equal(s)
}

// Pretty returns the provider identifier in an easy-to-read format.
func (t MethodType) Pretty() string {
	switch t {
	case MethodAccessToken:
		return "Access Token"
	case MethodOAuth2:
		return "OAuth2"
	case MethodOIDC:
		return "OIDC"
	case Method2FA:
		return "2FA"
	default:
		return txt.UpperFirst(t.String())
	}
}

// Method casts a string to a normalized method type.
func Method(s string) MethodType {
	switch s {
	case "", "-", "null", "nil", "0", "false":
		return MethodDefault
	case "oauth2", "oauth":
		return MethodOAuth2
	case "sso":
		return MethodOIDC
	case "two-factor", "totp":
		return Method2FA
	default:
		return MethodType(clean.TypeLower(s))
	}
}
