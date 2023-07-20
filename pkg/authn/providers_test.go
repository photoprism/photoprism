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
}

func TestProviderType_IsRemote(t *testing.T) {
	assert.Equal(t, false, ProviderLocal.IsRemote())
	assert.Equal(t, true, ProviderLDAP.IsRemote())
	assert.Equal(t, false, ProviderNone.IsRemote())
	assert.Equal(t, false, ProviderDefault.IsRemote())
	assert.Equal(t, false, ProviderUnknown.IsRemote())
}

func TestProviderType_IsLocal(t *testing.T) {
	assert.Equal(t, true, ProviderLocal.IsLocal())
	assert.Equal(t, false, ProviderLDAP.IsLocal())
	assert.Equal(t, false, ProviderNone.IsLocal())
	assert.Equal(t, false, ProviderDefault.IsLocal())
	assert.Equal(t, false, ProviderUnknown.IsLocal())
}

func TestProviderType_IsDefault(t *testing.T) {
	assert.Equal(t, false, ProviderLocal.IsDefault())
	assert.Equal(t, false, ProviderLDAP.IsDefault())
	assert.Equal(t, false, ProviderNone.IsDefault())
	assert.Equal(t, true, ProviderDefault.IsDefault())
	assert.Equal(t, true, ProviderUnknown.IsDefault())
}

func TestProviderType_Pretty(t *testing.T) {
	assert.Equal(t, "Local", ProviderLocal.Pretty())
	assert.Equal(t, "LDAP/AD", ProviderLDAP.Pretty())
	assert.Equal(t, "None", ProviderNone.Pretty())
	assert.Equal(t, "Default", ProviderDefault.Pretty())
	assert.Equal(t, "Default", ProviderUnknown.Pretty())
}

func TestProvider(t *testing.T) {
	assert.Equal(t, ProviderLocal, Provider("pass"))
	assert.Equal(t, ProviderLDAP, Provider("ad"))
	assert.Equal(t, ProviderDefault, Provider(""))
	assert.Equal(t, ProviderLink, Provider("url"))
}
