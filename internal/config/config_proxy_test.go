package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_HttpsProxy(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.HttpsProxy())

	_ = os.Setenv("HTTPS_PROXY", "https://foo.bar:8081")

	assert.Equal(t, "https://foo.bar:8081", c.HttpsProxy())

	_ = os.Setenv("HTTPS_PROXY", "")

	assert.Equal(t, "", c.HttpsProxy())
}

func TestConfig_HttpsProxyInsecure(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.HttpsProxyInsecure())

	_ = os.Setenv("HTTPS_PROXY", "https://foo.bar:8081")

	assert.False(t, c.HttpsProxyInsecure())

	_ = os.Setenv("HTTPS_PROXY", "")

	assert.False(t, c.HttpsProxyInsecure())
}
