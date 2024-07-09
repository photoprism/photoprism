package authn

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/txt"
)

// ProviderType represents an authentication provider type.
type ProviderType string

// Standard authentication provider types.
const (
	ProviderUndefined   ProviderType = ""
	ProviderDefault     ProviderType = "default"
	ProviderClient      ProviderType = "client"
	ProviderApplication ProviderType = "application"
	ProviderAccessToken ProviderType = "access_token"
	ProviderLocal       ProviderType = "local"
	ProviderOIDC        ProviderType = "oidc"
	ProviderLDAP        ProviderType = "ldap"
	ProviderLink        ProviderType = "link"
	ProviderNone        ProviderType = "none"
)

// LocalProviders contains local authentication providers (signing up with OIDC creates a local user account).
var LocalProviders = list.List{
	string(ProviderLocal),
	string(ProviderOIDC),
}

// LocalPasswordRequiredProviders contains authentication providers which require a local password.
var LocalPasswordRequiredProviders = list.List{
	string(ProviderUndefined),
	string(ProviderDefault),
	string(ProviderLocal),
}

// PasswordProviders contains authentication providers which support password authentication (local and remote).
var PasswordProviders = list.List{
	string(ProviderDefault),
	string(ProviderLocal),
	string(ProviderLDAP),
}

// PasscodeProviders contains authentication providers that support 2-Factor Authentication (2FA) with a TOTP passcode.
var PasscodeProviders = list.List{
	string(ProviderDefault),
	string(ProviderLocal),
	string(ProviderLDAP),
}

// ClientProviders contains all client authentication providers.
var ClientProviders = list.List{
	string(ProviderClient),
	string(ProviderApplication),
	string(ProviderAccessToken),
}

// Provider casts a string to a normalized provider type.
func Provider(s string) ProviderType {
	s = clean.TypeLowerUnderscore(s)
	switch s {
	case "", "_", "-", "null", "nil", "0", "false":
		return ProviderDefault
	case "token", "url":
		return ProviderLink
	case "pass", "passwd", "password":
		return ProviderLocal
	case "app", "application":
		return ProviderApplication
	case "oidc", "openid":
		return ProviderOIDC
	case "ldap", "ad", "ldap/ad", "ldap\\ad":
		return ProviderLDAP
	case "client", "client_credentials", "oauth2":
		return ProviderClient
	default:
		return ProviderType(s)
	}
}

// Providers casts a string to normalized provider type strings.
func Providers(s string) []ProviderType {
	items := strings.Split(s, ",")
	result := make([]ProviderType, 0, len(items))

	for i := range items {
		result = append(result, Provider(items[i]))
	}

	return result
}

// Pretty returns the provider identifier in an easy-to-read format.
func (t ProviderType) Pretty() string {
	switch t {
	case ProviderOIDC:
		return "OIDC"
	case ProviderLDAP:
		return "LDAP/AD"
	case ProviderClient:
		return "Client"
	case ProviderAccessToken:
		return "Access Token"
	default:
		return txt.UpperFirst(t.String())
	}
}

// String returns the provider identifier as a string.
func (t ProviderType) String() string {
	switch t {
	case "":
		return string(ProviderDefault)
	case "token":
		return string(ProviderLink)
	case "password":
		return string(ProviderLocal)
	case "client", "client credentials", "client_credentials", "oauth2":
		return string(ProviderClient)
	default:
		return string(t)
	}
}

// Equal checks if the type matches the specified string.
func (t ProviderType) Equal(s string) bool {
	return t == Provider(s)
}

// NotEqual checks if the type does not match the specified string.
func (t ProviderType) NotEqual(s string) bool {
	return !t.Equal(s)
}

// Is compares the provider with another type.
func (t ProviderType) Is(providerType ProviderType) bool {
	return t == providerType
}

// IsNot checks if the provider is not the specified type.
func (t ProviderType) IsNot(providerType ProviderType) bool {
	return t != providerType
}

// IsUndefined checks if the provider is undefined.
func (t ProviderType) IsUndefined() bool {
	return t == ""
}

// IsOIDC checks if the provider is OpenID Connect (OIDC).
func (t ProviderType) IsOIDC() bool {
	return t == ProviderOIDC
}

// IsLocal checks if local authentication is possible.
func (t ProviderType) IsLocal() bool {
	return list.Contains(LocalProviders, string(t))
}

// IsClient checks if the authentication is provided for a client.
func (t ProviderType) IsClient() bool {
	return list.Contains(ClientProviders, string(t))
}

// IsApplication checks if the authentication is provided for an application.
func (t ProviderType) IsApplication() bool {
	return t == ProviderApplication
}

// IsDefault checks if this is the default provider.
func (t ProviderType) IsDefault() bool {
	return t.String() == ProviderDefault.String()
}

// RequiresLocalPassword checks if the provider allows a password to be checked for authentication.
func (t ProviderType) RequiresLocalPassword() bool {
	return list.Contains(LocalPasswordRequiredProviders, string(t))
}

// SupportsPasswordAuthentication checks if the provider allows a password to be checked for authentication.
func (t ProviderType) SupportsPasswordAuthentication() bool {
	return list.Contains(PasswordProviders, string(t))
}

// SupportsPasscodeAuthentication checks if the provider supports two-factor authentication with a passcode.
func (t ProviderType) SupportsPasscodeAuthentication() bool {
	return list.Contains(PasscodeProviders, string(t))
}
