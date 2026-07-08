package scheme

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPort(t *testing.T) {
	t.Run("Http", func(t *testing.T) {
		assert.Equal(t, "80", DefaultPort(Http))
	})
	t.Run("Https", func(t *testing.T) {
		assert.Equal(t, "443", DefaultPort(Https))
	})
	t.Run("Websocket", func(t *testing.T) {
		assert.Equal(t, "443", DefaultPort(Websocket))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", DefaultPort(""))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, "", DefaultPort("ftp"))
	})
}

func TestStripDefaultPort(t *testing.T) {
	t.Run("HttpsDefault", func(t *testing.T) {
		u, err := url.Parse("https://example.com:443/path")
		assert.NoError(t, err)
		StripDefaultPort(u)
		assert.Equal(t, "https://example.com/path", u.String())
	})
	t.Run("HttpDefault", func(t *testing.T) {
		u, err := url.Parse("http://example.com:80/")
		assert.NoError(t, err)
		StripDefaultPort(u)
		assert.Equal(t, "http://example.com/", u.String())
	})
	t.Run("NonDefaultPortPreserved", func(t *testing.T) {
		u, err := url.Parse("https://example.com:8443/")
		assert.NoError(t, err)
		StripDefaultPort(u)
		assert.Equal(t, "https://example.com:8443/", u.String())
	})
	t.Run("MismatchedScheme", func(t *testing.T) {
		u, err := url.Parse("http://example.com:443/")
		assert.NoError(t, err)
		StripDefaultPort(u)
		assert.Equal(t, "http://example.com:443/", u.String())
	})
	t.Run("NoPort", func(t *testing.T) {
		u, err := url.Parse("https://example.com/")
		assert.NoError(t, err)
		StripDefaultPort(u)
		assert.Equal(t, "https://example.com/", u.String())
	})
	t.Run("IPv6Default", func(t *testing.T) {
		u, err := url.Parse("https://[::1]:443/")
		assert.NoError(t, err)
		StripDefaultPort(u)
		assert.Equal(t, "https://[::1]/", u.String())
	})
	t.Run("Nil", func(t *testing.T) {
		assert.NotPanics(t, func() { StripDefaultPort(nil) })
	})
	t.Run("UnknownScheme", func(t *testing.T) {
		u, err := url.Parse("ftp://example.com:21/")
		assert.NoError(t, err)
		StripDefaultPort(u)
		assert.Equal(t, "ftp://example.com:21/", u.String())
	})
}
