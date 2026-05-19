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

func TestNormalizeBaseURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		// Trailing-slash policy.
		{"AlreadyNormalized", "https://example.com/", "https://example.com/"},
		{"NoTrailingSlash", "https://example.com", "https://example.com/"},
		{"ExtraTrailingSlashes", "https://example.com:443////", "https://example.com/"},

		// Default-port stripping.
		{"HttpsDefaultPort", "https://example.com:443/", "https://example.com/"},
		{"HttpDefaultPort", "http://example.com:80/sub", "http://example.com/sub/"},
		{"NonDefaultPortPreserved", "https://example.com:8443/", "https://example.com:8443/"},
		{"MismatchedScheme", "http://example.com:443/", "http://example.com:443/"},

		// Uncommon but well-formed inputs.
		{"IPv6DefaultPort", "https://[::1]:443/", "https://[::1]/"},
		{"IPv6NonDefaultPort", "https://[2001:db8::1]:8443/path", "https://[2001:db8::1]:8443/path/"},
		{"PathPreserved", "https://example.com:443/i/pro-1/", "https://example.com/i/pro-1/"},
		{"QueryStripped", "https://example.com:443/i/?lang=de&page=2", "https://example.com/i/"},
		{"ForceQueryStripped", "https://example.com/?", "https://example.com/"},
		{"FragmentStripped", "https://example.com/library/#photo123", "https://example.com/library/"},

		// Policy: userinfo is preserved verbatim.
		{"UserinfoPreserved", "https://user:secret@example.com:443/", "https://user:secret@example.com/"},

		// Unix-socket schemes: port stripping must stay a no-op.
		{"UnixScheme", "unix:///var/run/photoprism.sock", "unix:///var/run/photoprism.sock/"},
		{"HttpUnixScheme", "http+unix:///var/run/photoprism.sock", "http+unix:///var/run/photoprism.sock/"},

		// Parse failure falls back to TrimRight + "/".
		{"ParseError", ":foo", ":foo/"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, NormalizeBaseURL(tc.in))
		})
	}
}
