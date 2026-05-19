package proxy

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestForwardedProto(t *testing.T) {
	t.Run("NilRequest", func(t *testing.T) {
		assert.Equal(t, "", ForwardedProto(nil))
	})
	t.Run("HeaderValue", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
		require.NoError(t, err)
		req.Header.Set(header.XForwardedProto, "https, http")
		assert.Equal(t, "https", ForwardedProto(req))
	})
	t.Run("TLSFallback", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
		require.NoError(t, err)
		req.TLS = &tls.ConnectionState{}
		assert.Equal(t, "https", ForwardedProto(req))
	})
	t.Run("HTTPFallback", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		require.NoError(t, err)
		assert.Equal(t, "http", ForwardedProto(req))
	})
}

func TestRewriteLocation(t *testing.T) {
	prefix := "/i/acme/"
	host := "portal.example.com"

	assert.Equal(t, "/i/acme/library", RewriteLocation("/library", prefix, host))
	assert.Equal(t, "/i/acme/", RewriteLocation("/", prefix, host))
	assert.Equal(t, "https://portal.example.com/i/acme/library", RewriteLocation("https://portal.example.com/library", prefix, host))
	assert.Equal(t, "https://other.example.com/library", RewriteLocation("https://other.example.com/library", prefix, host))
}

func TestRewriteSetCookiePath(t *testing.T) {
	prefix := "/i/acme/"

	assert.Equal(t, "session=1; Path=/i/acme/; HttpOnly", RewriteSetCookiePath("session=1; Path=/; HttpOnly", prefix))
	assert.Equal(t, "session=1; HttpOnly; Path=/i/acme/", RewriteSetCookiePath("session=1; HttpOnly", prefix))
	assert.Equal(t, "session=1; Path=/i/acme/; HttpOnly", RewriteSetCookiePath("session=1; Path=/i/acme/; HttpOnly", prefix))
}

func TestHostMatch(t *testing.T) {
	assert.True(t, HostMatch("portal.example.com", "portal.example.com"))
	assert.True(t, HostMatch("portal.example.com:443", "portal.example.com"))
	assert.True(t, HostMatch("Portal.Example.Com:443", "portal.example.com"))
	assert.False(t, HostMatch("node.example.com", "portal.example.com"))
}

func TestRewriteDestinationHost(t *testing.T) {
	upstream, err := url.Parse("http://instance.internal:2342")
	require.NoError(t, err)

	t.Run("RewritesMatchingHost", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://portal.example.com", nil)
		require.NoError(t, err)
		req.Header.Set("Destination", "http://portal.example.com/i/acme/import/dst.txt")

		RewriteDestinationHost(req, "portal.example.com", upstream)

		assert.Equal(t, "http://instance.internal:2342/i/acme/import/dst.txt", req.Header.Get("Destination"))
	})
	t.Run("SkipsDifferentHost", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://portal.example.com", nil)
		require.NoError(t, err)
		req.Header.Set("Destination", "http://other.example.com/i/acme/import/dst.txt")

		RewriteDestinationHost(req, "portal.example.com", upstream)

		assert.Equal(t, "http://other.example.com/i/acme/import/dst.txt", req.Header.Get("Destination"))
	})
}
