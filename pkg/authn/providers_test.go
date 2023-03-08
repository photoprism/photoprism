package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderString(t *testing.T) {
	assert.Equal(t, "default", ProviderString(""))
	assert.Equal(t, "default", ProviderString(ProviderDefault))
	assert.Equal(t, "none", ProviderString(ProviderNone))
	assert.Equal(t, "local", ProviderString(ProviderLocal))
	assert.Equal(t, "ldap", ProviderString(ProviderLDAP))
}
