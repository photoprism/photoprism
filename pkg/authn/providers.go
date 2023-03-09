package authn

import (
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/txt"
)

// ProviderType represents an authentication provider type.
type ProviderType string

// Authentication providers.
const (
	ProviderDefault ProviderType = "default"
	ProviderLocal   ProviderType = "local"
	ProviderLDAP    ProviderType = "ldap"
	ProviderToken   ProviderType = "token"
	ProviderNone    ProviderType = "none"
	ProviderUnknown ProviderType = ""
)

// RemoteProviders lists all remote auth providers.
var RemoteProviders = list.List{
	string(ProviderLDAP),
}

// LocalProviders lists all local auth providers.
var LocalProviders = list.List{
	string(ProviderLocal),
}

// IsRemote checks if the provider is external.
func (t ProviderType) IsRemote() bool {
	return list.Contains(RemoteProviders, string(t))
}

// IsLocal checks if local authentication is possible.
func (t ProviderType) IsLocal() bool {
	return list.Contains(LocalProviders, string(t))
}

// String returns the provider identifier as a string.
func (t ProviderType) String() string {
	switch t {
	case "":
		return string(ProviderDefault)
	case "password":
		return string(ProviderLocal)
	default:
		return string(t)
	}
}

// Pretty returns the provider identifier in an easy-to-read format.
func (t ProviderType) Pretty() string {
	switch t {
	case ProviderLDAP:
		return "LDAP/AD"
	default:
		return txt.UpperFirst(t.String())
	}
}

// Provider casts a string to a normalized provider type.
func Provider(s string) ProviderType {
	switch s {
	case "", "-", "null", "nil", "0", "false":
		return ProviderDefault
	case "pass", "passwd", "password":
		return ProviderLocal
	case "ldap", "ad", "ldap/ad", "ldap\\ad":
		return ProviderLDAP
	default:
		return ProviderType(clean.TypeLower(s))
	}
}
