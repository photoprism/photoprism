package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIssuer(t *testing.T) {
	assert.Equal(t, "", Issuer(""))
	assert.Equal(t, "", Issuer("         "))
	assert.Equal(t, "https://accounts.google.com", Issuer("https://accounts.google.com"))
	assert.Equal(t, "user://123456", Issuer("user://123456"))
	assert.Equal(t, "issuer.example.com", Issuer("issuer.example.com"))
	assert.Equal(t, "example", Issuer(" example "))
}
