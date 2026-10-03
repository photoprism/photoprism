package safe

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubLookup replaces the host name lookup for the duration of the test.
func stubLookup(t *testing.T, fn func(ctx context.Context, host string) ([]net.IPAddr, error)) {
	t.Helper()

	orig := lookupIPAddr
	lookupIPAddr = fn
	t.Cleanup(func() { lookupIPAddr = orig })
}

func TestDisallowedAddr(t *testing.T) {
	t.Run("Zone", func(t *testing.T) {
		assert.True(t, disallowedAddr(netip.MustParseAddr("fe80::1%eth0")))
		assert.True(t, disallowedAddr(netip.MustParseAddr("::%eth0")))
		assert.True(t, disallowedAddr(netip.MustParseAddr("fec0::1%eth0")))
		assert.True(t, disallowedAddr(netip.MustParseAddr("64:ff9b::7f00:1%eth0")))
		assert.False(t, disallowedAddr(netip.MustParseAddr("2606:4700:4700::1111%eth0")))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.True(t, disallowedAddr(netip.Addr{}))
	})
	t.Run("Embedded", func(t *testing.T) {
		assert.True(t, disallowedAddr(netip.MustParseAddr("64:ff9b::169.254.169.254")))
		assert.True(t, disallowedAddr(netip.MustParseAddr("2002:c0a8:101::1")))
		assert.False(t, disallowedAddr(netip.MustParseAddr("64:ff9b::1.1.1.1")))
	})
}

func TestCheckHost(t *testing.T) {
	t.Run("Literal", func(t *testing.T) {
		assert.ErrorIs(t, checkHost("127.0.0.1"), ErrPrivateIP)
		assert.ErrorIs(t, checkHost("::"), ErrPrivateIP)
		assert.NoError(t, checkHost("8.8.8.8"))
	})
	t.Run("ResolvesPrivate", func(t *testing.T) {
		stubLookup(t, func(context.Context, string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}, {IP: net.ParseIP("64:ff9b::7f00:1")}}, nil
		})
		assert.ErrorIs(t, checkHost("example.com"), ErrPrivateIP)
	})
	t.Run("ResolvesPublic", func(t *testing.T) {
		stubLookup(t, func(context.Context, string) ([]net.IPAddr, error) {
			return []net.IPAddr{{IP: net.ParseIP("2606:4700:4700::1111")}}, nil
		})
		assert.NoError(t, checkHost("example.com"))
	})
	t.Run("LookupError", func(t *testing.T) {
		lookupErr := errors.New("no such host")
		stubLookup(t, func(context.Context, string) ([]net.IPAddr, error) { return nil, lookupErr })
		assert.ErrorIs(t, checkHost("example.com"), lookupErr)
	})
}

func TestDialControl(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		assert.NoError(t, dialControl("tcp4", "8.8.8.8:443", nil))
		assert.NoError(t, dialControl("tcp6", "[2606:4700:4700::1111]:443", nil))
	})
	t.Run("Disallowed", func(t *testing.T) {
		for _, address := range []string{"127.0.0.1:80", "[::1]:80", "[::]:80", "[fe80::1%eth0]:80", "10.0.0.1:443", "[64:ff9b::a00:1]:80"} {
			assert.ErrorIs(t, dialControl("tcp", address, nil), ErrPrivateIP, address)
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.ErrorIs(t, dialControl("tcp", "example.com:80", nil), ErrPrivateIP)
		assert.ErrorIs(t, dialControl("tcp", "8.8.8.8", nil), ErrPrivateIP)
	})
}

func TestNewTransport(t *testing.T) {
	t.Run("AllowPrivate", func(t *testing.T) {
		assert.Same(t, http.DefaultTransport, newTransport(true))
	})
	t.Run("Guarded", func(t *testing.T) {
		base := http.DefaultTransport.(*http.Transport)
		transport, ok := newTransport(false).(*http.Transport)
		require.True(t, ok)
		assert.NotNil(t, transport.DialContext)
		assert.NotNil(t, transport.Proxy)
		assert.NotSame(t, http.DefaultTransport, transport)
		assert.True(t, transport.DisableKeepAlives)
		assert.Nil(t, transport.DialTLSContext)
		assert.Equal(t, base.TLSHandshakeTimeout, transport.TLSHandshakeTimeout)
		assert.Equal(t, base.IdleConnTimeout, transport.IdleConnTimeout)
		assert.Equal(t, base.ForceAttemptHTTP2, transport.ForceAttemptHTTP2)
		assert.Equal(t, base.TLSClientConfig, transport.TLSClientConfig)
	})
	t.Run("CustomDefault", func(t *testing.T) {
		orig := http.DefaultTransport
		http.DefaultTransport = http.RoundTripper(nil)
		t.Cleanup(func() { http.DefaultTransport = orig })

		transport, ok := newTransport(false).(*http.Transport)
		require.True(t, ok)
		assert.NotNil(t, transport.DialContext)
		assert.NotNil(t, transport.Proxy)
	})
}

// TestDownload_RefusesPrivateDial checks that a connection to a loopback address is refused before the
// server receives a request, independent of the host name check.
func TestDownload_RefusesPrivateDial(t *testing.T) {
	var hits atomic.Int32

	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte("ok"))
	})

	stubLookup(t, func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})

	rawURL := strings.Replace(ts.URL, "127.0.0.1", "localhost", 1)
	err := Download(filepath.Join(t.TempDir(), "out"), rawURL, &Options{Timeout: 5 * time.Second, AllowPrivate: false})

	assert.ErrorIs(t, err, ErrPrivateIP)
	assert.Equal(t, int32(0), hits.Load())
}

func TestCheckRedirect(t *testing.T) {
	newRequest := func(t *testing.T, rawURL string) *http.Request {
		t.Helper()
		req, err := http.NewRequest(http.MethodGet, rawURL, nil)
		require.NoError(t, err)
		return req
	}

	first := newRequest(t, "http://example.com/")
	first.Header.Set("Accept", "image/*")

	t.Run("Unspecified", func(t *testing.T) {
		assert.ErrorIs(t, checkRedirect(false)(newRequest(t, "http://[::]:80/"), []*http.Request{first}), ErrPrivateIP)
	})
	t.Run("Nat64Loopback", func(t *testing.T) {
		assert.ErrorIs(t, checkRedirect(false)(newRequest(t, "http://[64:ff9b::7f00:1]/"), []*http.Request{first}), ErrPrivateIP)
	})
	t.Run("Public", func(t *testing.T) {
		req := newRequest(t, "http://8.8.8.8/avatar.jpg")
		assert.NoError(t, checkRedirect(false)(req, []*http.Request{first}))
		assert.Equal(t, "image/*", req.Header.Get("Accept"))
	})
	t.Run("AllowPrivate", func(t *testing.T) {
		assert.NoError(t, checkRedirect(true)(newRequest(t, "http://127.0.0.1/"), []*http.Request{first}))
	})
	t.Run("Limit", func(t *testing.T) {
		via := make([]*http.Request, maxRedirects-1)
		for i := range via {
			via[i] = first
		}
		assert.NoError(t, checkRedirect(true)(newRequest(t, "http://127.0.0.1/"), via))
	})
	t.Run("TooMany", func(t *testing.T) {
		via := make([]*http.Request, maxRedirects)
		for i := range via {
			via[i] = first
		}
		assert.EqualError(t, checkRedirect(true)(newRequest(t, "http://127.0.0.1/"), via), "stopped after 10 redirects")
	})
}

// allowLoopbackDial lets the guarded transport connect to local test servers for the duration of the test.
func allowLoopbackDial(t *testing.T) {
	t.Helper()

	origDial, origPeer := dialControlFunc, disallowedPeer
	dialControlFunc = func(string, string, syscall.RawConn) error { return nil }
	disallowedPeer = func(ip net.IP) bool { return ip == nil }
	t.Cleanup(func() { dialControlFunc, disallowedPeer = origDial, origPeer })
}

// TestDownload_GuardedTransport checks downloads through the transport used when private networks are
// disallowed, with its connection check replaced so that a local test server can be reached.
func TestDownload_GuardedTransport(t *testing.T) {
	stubLookup(t, func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})
	allowLoopbackDial(t)

	var hits atomic.Int32

	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)

		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "http://127.0.0.1:"+r.URL.Query().Get("port")+"/ok", http.StatusFound) //nolint:gosec // G710: test server
			return
		}

		_, _ = w.Write([]byte("ok"))
	})

	host := strings.Replace(ts.URL, "127.0.0.1", "localhost", 1)
	port := ts.URL[strings.LastIndex(ts.URL, ":")+1:]

	t.Run("Success", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "out")
		require.NoError(t, Download(dest, host+"/ok", &Options{Timeout: 5 * time.Second, AllowPrivate: false}))
	})
	t.Run("TargetChecked", func(t *testing.T) {
		hits.Store(0)
		err := Download(filepath.Join(t.TempDir(), "out"), ts.URL+"/ok", &Options{Timeout: 5 * time.Second, AllowPrivate: false})
		assert.ErrorIs(t, err, ErrPrivateIP)
		assert.Equal(t, int32(0), hits.Load())
	})
	t.Run("RedirectChecked", func(t *testing.T) {
		hits.Store(0)
		err := Download(filepath.Join(t.TempDir(), "out"), host+"/redirect?port="+port, &Options{Timeout: 5 * time.Second, AllowPrivate: false})
		assert.ErrorIs(t, err, ErrPrivateIP)
		assert.Equal(t, int32(1), hits.Load())
	})
}
