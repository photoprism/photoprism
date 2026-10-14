package config

import (
	"net/http"
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

func TestConfig_initHttpsProxy(t *testing.T) {
	transport := http.DefaultTransport.(*http.Transport)
	prevTLS := transport.TLSClientConfig
	t.Cleanup(func() { transport.TLSClientConfig = prevTLS })
	t.Setenv("HTTPS_PROXY", "")

	t.Run("NotSet", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.HttpsProxy = ""
		c.initHttpsProxy()
		assert.Same(t, prevTLS, transport.TLSClientConfig)
		assert.Equal(t, "", os.Getenv("HTTPS_PROXY"))
	})
	t.Run("Set", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.HttpsProxy = "https://proxy.example.com:3128"
		c.options.HttpsProxyInsecure = true
		c.initHttpsProxy()
		assert.Equal(t, "https://proxy.example.com:3128", os.Getenv("HTTPS_PROXY"))
		if assert.NotNil(t, transport.TLSClientConfig) {
			assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
		}
	})
	t.Run("InitCore", func(t *testing.T) {
		// Commands that skip the database, such as cluster register, still use the proxy.
		t.Setenv("HTTPS_PROXY", "")
		c := NewConfig(CliTestContext())
		c.options.HttpsProxy = "https://core.example.com:3128"
		assert.NoError(t, c.InitCore())
		assert.Equal(t, "https://core.example.com:3128", os.Getenv("HTTPS_PROXY"))
	})
}
