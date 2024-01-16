package header

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestBlockCdn(t *testing.T) {
	t.Run("Allow", func(t *testing.T) {
		u, _ := url.Parse("/foo")

		r := &http.Request{
			URL:    u,
			Header: http.Header{CdnHost: []string{"host.cdn.com"}},
			Method: http.MethodGet,
		}

		assert.False(t, BlockCdn(r))
	})
	t.Run("UnsafeMethod", func(t *testing.T) {
		u, _ := url.Parse("/foo")

		r := &http.Request{
			URL:    u,
			Header: http.Header{CdnHost: []string{"host.cdn.com"}},
			Method: http.MethodPost,
		}

		assert.True(t, BlockCdn(r))
	})
	t.Run("Root", func(t *testing.T) {
		u, _ := url.Parse("/")

		r := &http.Request{
			URL:    u,
			Header: http.Header{CdnHost: []string{"host.cdn.com"}},
			Method: http.MethodGet,
		}

		assert.True(t, BlockCdn(r))
	})
	t.Run("NoCdn", func(t *testing.T) {
		u, _ := url.Parse("/foo")

		r := &http.Request{
			URL:    u,
			Header: http.Header{Accept: []string{"application/json"}},
			Method: http.MethodPost,
		}

		assert.False(t, BlockCdn(r))
	})
}
