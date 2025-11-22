package header

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCdn(t *testing.T) {
	t.Run("Header", func(t *testing.T) {
		assert.Equal(t, "Cdn-Host", CdnHost)
		assert.Equal(t, "Cdn-Mobiledevice", CdnMobileDevice)
		assert.Equal(t, "Cdn-Serverzone", CdnServerZone)
		assert.Equal(t, "Cdn-Serverid", CdnServerID)
		assert.Equal(t, "Cdn-Connectionid", CdnConnectionID)
	})
}

func TestIsCdn(t *testing.T) {
	t.Run("Header", func(t *testing.T) {
		u, _ := url.Parse("/foo")

		r := &http.Request{
			URL:    u,
			Header: http.Header{CdnHost: []string{"host.cdn.com"}},
			Method: http.MethodGet,
		}

		assert.True(t, IsCdn(r))
	})
	t.Run("EmptyHeader", func(t *testing.T) {
		u, _ := url.Parse("/foo")

		r := &http.Request{
			URL:    u,
			Header: http.Header{CdnHost: []string{""}},
			Method: http.MethodPost,
		}

		assert.False(t, IsCdn(r))
	})
	t.Run("NoHeader", func(t *testing.T) {
		u, _ := url.Parse("/foo")

		r := &http.Request{
			URL:    u,
			Header: http.Header{Accept: []string{"application/json"}},
			Method: http.MethodPost,
		}

		assert.False(t, IsCdn(r))
	})
}

func TestAbortCdnRequest(t *testing.T) {
	t.Run("Allow", func(t *testing.T) {
		u, _ := url.Parse("/foo")

		r := &http.Request{
			URL:    u,
			Header: http.Header{CdnHost: []string{"host.cdn.com"}},
			Method: http.MethodGet,
		}

		assert.False(t, AbortCdnRequest(r))
	})
	t.Run("UnsafeMethod", func(t *testing.T) {
		u, _ := url.Parse("/foo")

		r := &http.Request{
			URL:    u,
			Header: http.Header{CdnHost: []string{"host.cdn.com"}},
			Method: http.MethodPost,
		}

		assert.True(t, AbortCdnRequest(r))
	})
	t.Run("Root", func(t *testing.T) {
		u, _ := url.Parse("/")

		r := &http.Request{
			URL:    u,
			Header: http.Header{CdnHost: []string{"host.cdn.com"}},
			Method: http.MethodGet,
		}

		assert.True(t, AbortCdnRequest(r))
	})
	t.Run("NoCdn", func(t *testing.T) {
		u, _ := url.Parse("/foo")

		r := &http.Request{
			URL:    u,
			Header: http.Header{Accept: []string{"application/json"}},
			Method: http.MethodPost,
		}

		assert.False(t, AbortCdnRequest(r))
	})
}
