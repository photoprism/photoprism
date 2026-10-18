package safe

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetch(t *testing.T) {
	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/large":
			_, _ = w.Write([]byte(strings.Repeat("x", 2048)))
		case "/redirect":
			http.Redirect(w, r, "/ok", http.StatusFound)
		case "/slow":
			time.Sleep(500 * time.Millisecond)
			_, _ = w.Write([]byte("slow"))
		case "/missing":
			http.NotFound(w, r)
		default:
			w.Header().Set("Cache-Control", "max-age=60")
			_, _ = w.Write([]byte(r.Header.Get("Accept"))) //nolint:gosec // test server echoes the request header
		}
	})

	t.Run("Success", func(t *testing.T) {
		body, header, err := Fetch(ts.URL+"/ok", &Options{AllowPrivate: true, Accept: "application/json"})
		require.NoError(t, err)
		assert.Equal(t, "application/json", string(body))
		assert.Equal(t, "max-age=60", header.Get("Cache-Control"))
	})
	t.Run("TooLarge", func(t *testing.T) {
		_, _, err := Fetch(ts.URL+"/large", &Options{AllowPrivate: true, MaxSizeBytes: 1024})
		assert.ErrorIs(t, err, ErrSizeExceeded)
	})
	t.Run("NoRedirects", func(t *testing.T) {
		_, _, err := Fetch(ts.URL+"/redirect", &Options{AllowPrivate: true, NoRedirects: true})
		assert.ErrorIs(t, err, ErrRedirect)

		body, _, err := Fetch(ts.URL+"/redirect", &Options{AllowPrivate: true})
		require.NoError(t, err)
		assert.Equal(t, "*/*", string(body))
	})
	t.Run("Timeout", func(t *testing.T) {
		_, _, err := Fetch(ts.URL+"/slow", &Options{AllowPrivate: true, Timeout: 100 * time.Millisecond})
		assert.Error(t, err)
	})
	t.Run("Status", func(t *testing.T) {
		_, _, err := Fetch(ts.URL+"/missing", &Options{AllowPrivate: true})
		assert.Error(t, err)
	})
	t.Run("PrivateRefused", func(t *testing.T) {
		_, _, err := Fetch(ts.URL+"/ok", &Options{AllowPrivate: false})
		assert.ErrorIs(t, err, ErrPrivateIP)
	})
	t.Run("InvalidScheme", func(t *testing.T) {
		_, _, err := Fetch("file:///etc/passwd", nil)
		assert.Error(t, err)
	})
}

func TestFetch_GuardedTransport(t *testing.T) {
	stubLookup(t, func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})
	allowLoopbackDial(t)

	var hits atomic.Int32

	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte("ok"))
	})

	host := strings.Replace(ts.URL, "127.0.0.1", "localhost", 1)

	t.Run("Success", func(t *testing.T) {
		body, _, err := Fetch(host+"/ok", &Options{Timeout: 5 * time.Second, AllowPrivate: false})
		require.NoError(t, err)
		assert.Equal(t, "ok", string(body))
	})
	t.Run("TargetChecked", func(t *testing.T) {
		hits.Store(0)
		_, _, err := Fetch(ts.URL+"/ok", &Options{Timeout: 5 * time.Second, AllowPrivate: false})
		assert.ErrorIs(t, err, ErrPrivateIP)
		assert.Equal(t, int32(0), hits.Load())
	})
}

func TestFetch_RefusesPrivateDial(t *testing.T) {
	var hits atomic.Int32

	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte("ok"))
	})

	stubLookup(t, func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})

	_, _, err := Fetch(strings.Replace(ts.URL, "127.0.0.1", "localhost", 1), &Options{Timeout: 5 * time.Second, AllowPrivate: false})

	assert.ErrorIs(t, err, ErrPrivateIP)
	assert.Equal(t, int32(0), hits.Load())
}

func TestFetch_LookupDeadline(t *testing.T) {
	stubLookup(t, func(ctx context.Context, _ string) ([]net.IPAddr, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})

	start := time.Now()
	_, _, err := Fetch("https://example.com/client.json", &Options{Timeout: 100 * time.Millisecond, AllowPrivate: false})

	assert.Error(t, err)
	assert.Less(t, time.Since(start), 2*time.Second)
}

func TestCheckHostContext(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		assert.NoError(t, checkHostContext(context.Background(), "8.8.8.8"))
	})
	t.Run("Private", func(t *testing.T) {
		assert.ErrorIs(t, checkHostContext(context.Background(), "10.0.0.1"), ErrPrivateIP)
	})
	t.Run("Canceled", func(t *testing.T) {
		stubLookup(t, func(ctx context.Context, _ string) ([]net.IPAddr, error) {
			return nil, ctx.Err()
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.Error(t, checkHostContext(ctx, "example.com"))
	})
}

func TestStatusError(t *testing.T) {
	ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	_, _, err := Fetch(ts.URL+"/missing", &Options{AllowPrivate: true})

	var statusErr *StatusError
	require.ErrorAs(t, err, &statusErr)
	assert.Equal(t, http.StatusNotFound, statusErr.Code)
	assert.Equal(t, "unexpected status 404", statusErr.Error())
}
