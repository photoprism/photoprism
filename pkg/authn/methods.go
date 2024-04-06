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
	MethodUndefined MethodType = ""
	MethodDefault   MethodType = "default"
	MethodSession   MethodType = "session"
	MethodOAuth2    MethodType = "oauth2"
	MethodOIDC      MethodType = "oidc"
	Method2FA       MethodType = "2fa"
)

// Method casts a string to a normalized method type.
func Method(s string) MethodType {
	s = clean.TypeLowerUnderscore(s)
	switch s {
	case "":
		return MethodUndefined
	case "_", "-", "null", "nil", "0", "false":
		return MethodDefault
	case "oauth2", "oauth":
		return MethodOAuth2
	case "sso":
		return MethodOIDC
	case "2fa", "mfa", "otp", "totp":
		return Method2FA
	case "access_token":
		return MethodDefault
	default:
		return MethodType(s)
	}
}

// Pretty returns the provider identifier in an easy-to-read format.
func (t MethodType) Pretty() string {
	switch t {
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

// String returns the provider identifier as a string.
func (t MethodType) String() string {
	switch t {
	case "", "access_token":
		return string(MethodDefault)
	case "oauth":
		return string(MethodOAuth2)
	case "openid":
		return string(MethodOIDC)
	case "2fa", "otp", "totp":
		return string(Method2FA)
	default:
		return string(t)
	}
}

// Equal checks if the type matches the specified string.
func (t MethodType) Equal(s string) bool {
	return strings.EqualFold(s, t.String())
}

// NotEqual checks if the type does not match the specified string.
func (t MethodType) NotEqual(s string) bool {
	return !t.Equal(s)
}

// Is compares the method with another type.
func (t MethodType) Is(methodType MethodType) bool {
	return t == methodType
}

// IsNot checks if the method is not the specified type.
func (t MethodType) IsNot(methodType MethodType) bool {
	return t != methodType
}

// IsUndefined checks if the method is undefined.
func (t MethodType) IsUndefined() bool {
	return t == ""
}

// IsDefault checks if this is the default method.
func (t MethodType) IsDefault() bool {
	return t.String() == MethodDefault.String()
}

// IsSession checks if this is the session method.
func (t MethodType) IsSession() bool {
	return t.String() == MethodSession.String()
}
