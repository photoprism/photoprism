package proxy

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestParseTrustedProxies(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, ParseTrustedProxies(nil))
		assert.Nil(t, ParseTrustedProxies([]string{}))
	})
	t.Run("Success", func(t *testing.T) {
		result := ParseTrustedProxies([]string{"172.16.0.0/12", "192.168.1.5", "fd00::1/64", "::ffff:10.0.0.0/104", "172.16.0.0/012"})
		require.Len(t, result, 5)
		assert.Equal(t, "172.16.0.0/12", result[0].String())
		assert.Equal(t, "192.168.1.5/32", result[1].String())
		assert.Equal(t, "fd00::/64", result[2].String())
		assert.Equal(t, "10.0.0.0/8", result[3].String())
		assert.Equal(t, "172.16.0.0/12", result[4].String())
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Nil(t, ParseTrustedProxies([]string{"172.16.0.0/12", "proxy.example.com"}))
		assert.Nil(t, ParseTrustedProxies([]string{"10.0.0.0/8", "10.0.0.0/33"}))
		assert.Nil(t, ParseTrustedProxies([]string{"10.0.0.0/8", "fe80::1%eth0"}))
		assert.Nil(t, ParseTrustedProxies([]string{"172.18.0.2", ""}))
		assert.Nil(t, ParseTrustedProxies([]string{" 10.0.0.1 "}))
	})
}

func TestTrustedIP(t *testing.T) {
	trusted := ParseTrustedProxies([]string{"172.16.0.0/12", "2001:db8::1"})

	t.Run("InRange", func(t *testing.T) {
		assert.True(t, TrustedIP("172.18.0.5", trusted))
		assert.True(t, TrustedIP("::ffff:172.18.0.5", trusted))
		assert.True(t, TrustedIP("2001:db8::1", trusted))
	})
	t.Run("OutOfRange", func(t *testing.T) {
		assert.False(t, TrustedIP("203.0.113.7", trusted))
		assert.False(t, TrustedIP("2001:db8::2", trusted))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, TrustedIP("", trusted))
		assert.False(t, TrustedIP("172.18.0.5:41234", trusted))
		assert.False(t, TrustedIP("172.18.0.5", nil))
	})
}

func TestTrustedPeer(t *testing.T) {
	trusted := ParseTrustedProxies([]string{"172.16.0.0/12", "2001:db8::1"})

	t.Run("InRange", func(t *testing.T) {
		assert.True(t, TrustedPeer("172.18.0.5:41234", trusted))
		assert.True(t, TrustedPeer(" [::ffff:172.18.0.5]:41234", trusted))
		assert.True(t, TrustedPeer("[2001:db8::1]:443", trusted))
	})
	t.Run("OutOfRange", func(t *testing.T) {
		assert.False(t, TrustedPeer("203.0.113.7:41234", trusted))
		assert.False(t, TrustedPeer("[2001:db8::2]:443", trusted))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, TrustedPeer("", trusted))
		assert.False(t, TrustedPeer("172.18.0.5", trusted))
		assert.False(t, TrustedPeer("@", trusted))
		assert.False(t, TrustedPeer("172.18.0.5:41234", nil))
	})
	t.Run("MatchesGin", func(t *testing.T) {
		lists := [][]string{
			{"172.16.0.0/12"},
			{"172.16.0.0/012"},
			{"::ffff:172.16.0.0/108"},
			{"172.18.0.2", ""},
			{" 172.18.0.2 "},
			{"172.18.0.2", "fe80::1%eth0"},
			{"2001:db8::/32", "10.0.0.1"},
			{"::ffff:172.18.0.2"},
			{},
		}
		peers := []string{"172.18.0.2:1", "[::ffff:172.18.0.2]:1", "10.0.0.1:1", "[2001:db8::5]:1", "203.0.113.7:1", "[fe80::1]:1", "[::1]:1"}

		for _, list := range lists {
			var clientIP string

			engine := gin.New()
			if err := engine.SetTrustedProxies(list); err != nil {
				require.NoError(t, engine.SetTrustedProxies(nil))
			}
			engine.GET("/", func(c *gin.Context) { clientIP = c.ClientIP() })

			trusted := ParseTrustedProxies(list)

			for _, peer := range peers {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.RemoteAddr = peer
				req.Header.Set(header.XForwardedFor, "198.51.100.1")
				engine.ServeHTTP(httptest.NewRecorder(), req)

				assert.Equal(t, clientIP == "198.51.100.1", TrustedPeer(peer, trusted), "list %q, peer %s", list, peer)
			}
		}
	})
}

func TestTrustedRequest(t *testing.T) {
	trusted := ParseTrustedProxies([]string{"172.16.0.0/12"})

	t.Run("NilRequest", func(t *testing.T) {
		assert.False(t, TrustedRequest(nil, trusted))
	})
	t.Run("TrustedPeer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "172.18.0.5:41234"
		assert.True(t, TrustedRequest(req, trusted))
		req.RemoteAddr = "203.0.113.7:41234"
		assert.False(t, TrustedRequest(req, trusted))
	})
	t.Run("UnixSocket", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "@"
		assert.False(t, TrustedRequest(req, trusted))
		req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, &net.UnixAddr{Name: "/run/photoprism.sock", Net: "unix"}))
		assert.True(t, TrustedRequest(req, nil))
	})
	t.Run("TcpListener", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "203.0.113.7:41234"
		req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, &net.TCPAddr{IP: net.ParseIP("172.18.0.2"), Port: 2342}))
		assert.False(t, TrustedRequest(req, trusted))
	})
}

func TestForwardedProto(t *testing.T) {
	httpsHeaders := map[string]string{header.XForwardedProto: "https", header.XForwardedSsl: "on"}

	newRequest := func(t *testing.T, headers map[string]string) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		require.NoError(t, err)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req
	}

	t.Run("NilRequest", func(t *testing.T) {
		assert.Equal(t, "", ForwardedProto(nil, true, httpsHeaders))
	})
	t.Run("TLS", func(t *testing.T) {
		req := newRequest(t, map[string]string{header.XForwardedProto: "http"})
		req.TLS = &tls.ConnectionState{}
		assert.Equal(t, "https", ForwardedProto(req, false, httpsHeaders))
	})
	t.Run("TrustedHeader", func(t *testing.T) {
		assert.Equal(t, "https", ForwardedProto(newRequest(t, map[string]string{header.XForwardedProto: "http, HTTPS"}), true, httpsHeaders))
		assert.Equal(t, "https", ForwardedProto(newRequest(t, map[string]string{header.XForwardedSsl: "on"}), true, httpsHeaders))
	})
	t.Run("TrustedHeaderLines", func(t *testing.T) {
		req := newRequest(t, nil)
		req.Header.Add(header.XForwardedProto, "http")
		req.Header.Add(header.XForwardedProto, "https")
		assert.Equal(t, "https", ForwardedProto(req, true, httpsHeaders))
		req = newRequest(t, nil)
		req.Header.Add(header.XForwardedProto, "https")
		req.Header.Add(header.XForwardedProto, "http")
		assert.Equal(t, "http", ForwardedProto(req, true, httpsHeaders))
	})
	t.Run("TrustedHttp", func(t *testing.T) {
		assert.Equal(t, "http", ForwardedProto(newRequest(t, map[string]string{header.XForwardedProto: "https, http"}), true, httpsHeaders))
		assert.Equal(t, "http", ForwardedProto(newRequest(t, map[string]string{header.XForwardedProto: "https,"}), true, httpsHeaders))
		assert.Equal(t, "http", ForwardedProto(newRequest(t, map[string]string{header.XForwardedSsl: "off"}), true, httpsHeaders))
		assert.Equal(t, "http", ForwardedProto(newRequest(t, nil), true, httpsHeaders))
	})
	t.Run("UntrustedHeader", func(t *testing.T) {
		assert.Equal(t, "http", ForwardedProto(newRequest(t, map[string]string{header.XForwardedProto: "https"}), false, httpsHeaders))
	})
	t.Run("UnconfiguredHeader", func(t *testing.T) {
		assert.Equal(t, "http", ForwardedProto(newRequest(t, map[string]string{header.XUrlScheme: "https"}), true, httpsHeaders))
	})
}
