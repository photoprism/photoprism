package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderType_String(t *testing.T) {
	assert.Equal(t, "default", ProviderUnknown.String())
	assert.Equal(t, "default", ProviderDefault.String())
	assert.Equal(t, "none", ProviderNone.String())
	assert.Equal(t, "local", ProviderLocal.String())
	assert.Equal(t, "ldap", ProviderLDAP.String())
	assert.Equal(t, "link", ProviderLink.String())
	assert.Equal(t, "access_token", ProviderAccessToken.String())
	assert.Equal(t, "client_credentials", ProviderClientCredentials.String())
}

func TestProviderType_IsRemote(t *testing.T) {
	assert.False(t, ProviderLocal.IsRemote())
	assert.True(t, ProviderLDAP.IsRemote())
	assert.False(t, ProviderNone.IsRemote())
	assert.False(t, ProviderDefault.IsRemote())
	assert.False(t, ProviderUnknown.IsRemote())
}

func TestProviderType_IsLocal(t *testing.T) {
	assert.True(t, ProviderLocal.IsLocal())
	assert.False(t, ProviderLDAP.IsLocal())
	assert.False(t, ProviderNone.IsLocal())
	assert.False(t, ProviderDefault.IsLocal())
	assert.False(t, ProviderUnknown.IsLocal())
}

func TestProviderType_IsDefault(t *testing.T) {
	assert.False(t, ProviderLocal.IsDefault())
	assert.False(t, ProviderLDAP.IsDefault())
	assert.False(t, ProviderNone.IsDefault())
	assert.True(t, ProviderDefault.IsDefault())
	assert.True(t, ProviderUnknown.IsDefault())
}

func TestProviderType_IsClient(t *testing.T) {
	assert.False(t, ProviderLocal.IsClient())
	assert.False(t, ProviderLDAP.IsClient())
	assert.False(t, ProviderNone.IsClient())
	assert.False(t, ProviderDefault.IsClient())
	assert.True(t, ProviderClient.IsClient())
	assert.True(t, ProviderClientCredentials.IsClient())
}

func TestProviderType_Equal(t *testing.T) {
	assert.True(t, ProviderClient.Equal("Client"))
	assert.False(t, ProviderLocal.Equal("Client"))
}

func TestProviderType_NotEqual(t *testing.T) {
	assert.False(t, ProviderClient.NotEqual("Client"))
	assert.True(t, ProviderLocal.NotEqual("Client"))
}

func TestProviderType_Pretty(t *testing.T) {
	assert.Equal(t, "Local", ProviderLocal.Pretty())
	assert.Equal(t, "LDAP/AD", ProviderLDAP.Pretty())
	assert.Equal(t, "None", ProviderNone.Pretty())
	assert.Equal(t, "Default", ProviderDefault.Pretty())
	assert.Equal(t, "Default", ProviderUnknown.Pretty())
	assert.Equal(t, "Client", ProviderClient.Pretty())
	assert.Equal(t, "Access Token", ProviderAccessToken.Pretty())
	assert.Equal(t, "Client Credentials", ProviderClientCredentials.Pretty())
}

func TestProvider(t *testing.T) {
	assert.Equal(t, ProviderLocal, Provider("pass"))
	assert.Equal(t, ProviderLDAP, Provider("ad"))
	assert.Equal(t, ProviderDefault, Provider(""))
	assert.Equal(t, ProviderLink, Provider("url"))
	assert.Equal(t, ProviderDefault, Provider("default"))
	assert.Equal(t, ProviderClientCredentials, Provider("oauth2"))
}

func TestProviderType_IsUnknown(t *testing.T) {
	assert.True(t, ProviderUnknown.IsUnknown())
	assert.False(t, ProviderLocal.IsUnknown())
}

func TestProviderType_IsApplication(t *testing.T) {
	assert.True(t, ProviderApplication.IsApplication())
	assert.False(t, ProviderLocal.IsApplication())
}
