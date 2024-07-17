package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderType_String(t *testing.T) {
	assert.Equal(t, "default", ProviderUndefined.String())
	assert.Equal(t, "default", ProviderDefault.String())
	assert.Equal(t, "none", ProviderNone.String())
	assert.Equal(t, "local", ProviderLocal.String())
	assert.Equal(t, "oidc", ProviderOIDC.String())
	assert.Equal(t, "ldap", ProviderLDAP.String())
	assert.Equal(t, "link", ProviderLink.String())
	assert.Equal(t, "access_token", ProviderAccessToken.String())
	assert.Equal(t, "client", ProviderClient.String())
}

func TestProviderType_Is(t *testing.T) {
	assert.False(t, ProviderLocal.Is(ProviderLDAP))
	assert.True(t, ProviderOIDC.Is(ProviderOIDC))
	assert.False(t, ProviderOIDC.Is(ProviderLDAP))
	assert.True(t, ProviderLDAP.Is(ProviderLDAP))
	assert.False(t, ProviderClient.Is(ProviderLDAP))
	assert.False(t, ProviderApplication.Is(ProviderLDAP))
	assert.False(t, ProviderAccessToken.Is(ProviderLDAP))
	assert.False(t, ProviderNone.Is(ProviderLDAP))
	assert.False(t, ProviderDefault.Is(ProviderLDAP))
	assert.False(t, ProviderUndefined.Is(ProviderLDAP))
}

func TestProviderType_IsNot(t *testing.T) {
	assert.False(t, ProviderLocal.IsNot(ProviderLocal))
	assert.False(t, ProviderOIDC.IsNot(ProviderOIDC))
	assert.True(t, ProviderOIDC.IsNot(ProviderLDAP))
	assert.True(t, ProviderLDAP.IsNot(ProviderLocal))
	assert.False(t, ProviderClient.IsNot(ProviderClient))
	assert.False(t, ProviderApplication.IsNot(ProviderApplication))
	assert.False(t, ProviderAccessToken.IsNot(ProviderAccessToken))
	assert.False(t, ProviderNone.IsNot(ProviderNone))
	assert.False(t, ProviderDefault.IsNot(ProviderDefault))
	assert.False(t, ProviderUndefined.IsNot(ProviderUndefined))
}

func TestProviderType_IsUndefined(t *testing.T) {
	assert.True(t, ProviderUndefined.IsUndefined())
	assert.True(t, ProviderUndefined.IsDefault())
	assert.False(t, ProviderLocal.IsUndefined())
	assert.False(t, ProviderOIDC.IsUndefined())
}

func TestProviderType_IsLocal(t *testing.T) {
	assert.True(t, ProviderLocal.IsLocal())
	assert.True(t, ProviderOIDC.IsLocal())
	assert.False(t, ProviderLDAP.IsLocal())
	assert.False(t, ProviderClient.IsLocal())
	assert.False(t, ProviderApplication.IsLocal())
	assert.False(t, ProviderAccessToken.IsLocal())
	assert.False(t, ProviderNone.IsLocal())
	assert.False(t, ProviderDefault.IsLocal())
	assert.False(t, ProviderUndefined.IsLocal())
}

func TestProviderType_IsOIDC(t *testing.T) {
	assert.False(t, ProviderLocal.IsOIDC())
	assert.True(t, ProviderOIDC.IsOIDC())
	assert.False(t, ProviderLDAP.IsOIDC())
	assert.False(t, ProviderClient.IsOIDC())
	assert.False(t, ProviderApplication.IsOIDC())
	assert.False(t, ProviderAccessToken.IsOIDC())
	assert.False(t, ProviderNone.IsOIDC())
	assert.False(t, ProviderDefault.IsOIDC())
	assert.False(t, ProviderUndefined.IsOIDC())
}

func TestProviderType_SupportsPasscode(t *testing.T) {
	assert.True(t, ProviderLocal.SupportsPasscodeAuthentication())
	assert.True(t, ProviderOIDC.SupportsPasscodeAuthentication())
	assert.True(t, ProviderLDAP.SupportsPasscodeAuthentication())
	assert.False(t, ProviderClient.SupportsPasscodeAuthentication())
	assert.False(t, ProviderApplication.SupportsPasscodeAuthentication())
	assert.False(t, ProviderAccessToken.SupportsPasscodeAuthentication())
	assert.False(t, ProviderNone.SupportsPasscodeAuthentication())
	assert.True(t, ProviderDefault.SupportsPasscodeAuthentication())
	assert.False(t, ProviderUndefined.SupportsPasscodeAuthentication())
}

func TestProviderType_RequiresLocalPassword(t *testing.T) {
	assert.True(t, ProviderLocal.RequiresLocalPassword())
	assert.False(t, ProviderOIDC.RequiresLocalPassword())
	assert.False(t, ProviderLDAP.RequiresLocalPassword())
	assert.False(t, ProviderClient.RequiresLocalPassword())
	assert.False(t, ProviderApplication.RequiresLocalPassword())
	assert.False(t, ProviderAccessToken.RequiresLocalPassword())
	assert.False(t, ProviderNone.RequiresLocalPassword())
	assert.True(t, ProviderDefault.RequiresLocalPassword())
	assert.False(t, ProviderUndefined.RequiresLocalPassword())
}

func TestProviderType_SupportsPasswordAuthentication(t *testing.T) {
	assert.True(t, ProviderLocal.SupportsPasswordAuthentication())
	assert.False(t, ProviderOIDC.SupportsPasswordAuthentication())
	assert.True(t, ProviderLDAP.SupportsPasswordAuthentication())
	assert.False(t, ProviderClient.SupportsPasswordAuthentication())
	assert.False(t, ProviderApplication.SupportsPasswordAuthentication())
	assert.False(t, ProviderAccessToken.SupportsPasswordAuthentication())
	assert.False(t, ProviderNone.SupportsPasswordAuthentication())
	assert.True(t, ProviderDefault.SupportsPasswordAuthentication())
	assert.False(t, ProviderUndefined.SupportsPasswordAuthentication())
}

func TestProviderType_IsDefault(t *testing.T) {
	assert.False(t, ProviderLocal.IsDefault())
	assert.False(t, ProviderOIDC.IsDefault())
	assert.False(t, ProviderLDAP.IsDefault())
	assert.False(t, ProviderNone.IsDefault())
	assert.True(t, ProviderDefault.IsDefault())
	assert.True(t, ProviderUndefined.IsDefault())
}

func TestProviderType_IsClient(t *testing.T) {
	assert.False(t, ProviderLocal.IsClient())
	assert.False(t, ProviderOIDC.IsClient())
	assert.False(t, ProviderLDAP.IsClient())
	assert.False(t, ProviderNone.IsClient())
	assert.False(t, ProviderDefault.IsClient())
	assert.True(t, ProviderClient.IsClient())
}

func TestProviderType_Equal(t *testing.T) {
	assert.True(t, ProviderOIDC.Equal("OIDC"))
	assert.True(t, ProviderLDAP.Equal("LDAP"))
	assert.True(t, ProviderClient.Equal("Client"))
	assert.True(t, ProviderClient.Equal("Client Credentials"))
	assert.False(t, ProviderLocal.Equal("Client"))
}

func TestProviderType_NotEqual(t *testing.T) {
	assert.False(t, ProviderOIDC.NotEqual("OIDC"))
	assert.False(t, ProviderLDAP.NotEqual("LDAP"))
	assert.False(t, ProviderClient.NotEqual("Client"))
	assert.False(t, ProviderClient.NotEqual("Client Credentials"))
	assert.True(t, ProviderLocal.NotEqual("Client"))
}

func TestProviderType_Pretty(t *testing.T) {
	assert.Equal(t, "Local", ProviderLocal.Pretty())
	assert.Equal(t, "OIDC", ProviderOIDC.Pretty())
	assert.Equal(t, "LDAP/AD", ProviderLDAP.Pretty())
	assert.Equal(t, "None", ProviderNone.Pretty())
	assert.Equal(t, "Default", ProviderDefault.Pretty())
	assert.Equal(t, "Default", ProviderUndefined.Pretty())
	assert.Equal(t, "Access Token", ProviderAccessToken.Pretty())
	assert.Equal(t, "Client", ProviderClient.Pretty())
}

func TestProvider(t *testing.T) {
	assert.Equal(t, ProviderLocal, Provider("pass"))
	assert.Equal(t, ProviderOIDC, Provider("oidc"))
	assert.Equal(t, ProviderLDAP, Provider("ad"))
	assert.Equal(t, ProviderDefault, Provider(""))
	assert.Equal(t, ProviderLink, Provider("url"))
	assert.Equal(t, ProviderApplication, Provider("app"))
	assert.Equal(t, ProviderDefault, Provider("default"))
	assert.Equal(t, ProviderClient, Provider("oauth2"))
}

func TestProviders(t *testing.T) {
	types := Providers("pass, oidc")
	assert.Equal(t, ProviderLocal, types[0])
	assert.Equal(t, ProviderOIDC, types[1])
}

func TestProviderType_IsApplication(t *testing.T) {
	assert.True(t, ProviderApplication.IsApplication())
	assert.False(t, ProviderLocal.IsApplication())
}
