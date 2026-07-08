package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

// newCompressTestRouter builds a router with the configured compression
// preferences applied via NewCompressMiddleware. The optional handler is
// registered at /ok for round-trip assertions; additional routes that share
// the standard handler can be registered by the caller.
func newCompressTestRouter(t *testing.T, prefs string) (*gin.Engine, *config.Config) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	conf := config.TestConfig()
	conf.Options().HttpCompression = prefs

	r := gin.New()
	r.Use(NewCompressMiddleware(conf))
	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "hello world")
	})
	return r, conf
}

func TestNewCompressMiddleware_Zstd(t *testing.T) {
	r, conf := newCompressTestRouter(t, "zstd")

	excludedPath := conf.BaseUri(config.ApiUri + "/dl/test")
	r.GET(excludedPath, func(c *gin.Context) {
		c.String(http.StatusOK, "download")
	})

	doRequest := func(path, accept string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if accept != "" {
			req.Header.Set("Accept-Encoding", accept)
		}
		r.ServeHTTP(w, req)
		return w
	}

	t.Run("EncodesAndRoundTrips", func(t *testing.T) {
		w := doRequest("/ok", "zstd")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "zstd", w.Header().Get("Content-Encoding"))
		assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding")

		zr, err := zstd.NewReader(strings.NewReader(w.Body.String()))
		require.NoError(t, err)
		defer zr.Close()
		decoded, err := io.ReadAll(zr)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(decoded))
	})

	t.Run("ClientWithoutZstdGetsIdentityWithVary", func(t *testing.T) {
		w := doRequest("/ok", "gzip")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding")
		assert.Equal(t, "hello world", w.Body.String())
	})

	t.Run("MissingAcceptEncodingFallsThrough", func(t *testing.T) {
		w := doRequest("/ok", "")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding")
		assert.Equal(t, "hello world", w.Body.String())
	})

	t.Run("ExcludedPathIsNeverCompressed", func(t *testing.T) {
		w := doRequest(excludedPath, "zstd")
		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Empty(t, w.Header().Get("Vary"))
		assert.Equal(t, "download", w.Body.String())
	})

	t.Run("ConnectionUpgradeBypassesCompression", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		req.Header.Set("Accept-Encoding", "zstd")
		req.Header.Set("Connection", "Upgrade")
		req.Header.Set("Upgrade", "websocket")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Empty(t, w.Header().Get("Vary"))
	})
}

func TestNewCompressMiddleware_PrefersZstdOverGzip(t *testing.T) {
	r, _ := newCompressTestRouter(t, "zstd,gzip")

	t.Run("BothAcceptedZstdWins", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		req.Header.Set("Accept-Encoding", "gzip, zstd")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "zstd", w.Header().Get("Content-Encoding"))
		assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding")
	})

	t.Run("OnlyGzipAcceptedFallsBack", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
	})

	t.Run("UnknownEncodingYieldsIdentityWithVary", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		req.Header.Set("Accept-Encoding", "br")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding")
		assert.Equal(t, "hello world", w.Body.String())
	})
}

func TestNewCompressMiddleware_BypassesPartialContent(t *testing.T) {
	// 206 Partial Content responses serve a slice of the identity
	// representation. Compressing them would scramble the byte offsets the
	// client asked for and contradict the Content-Range header. The bundled
	// `/static/*` route is now served by PrecompressedStatic and excluded
	// from the runtime middleware, so this test uses a non-excluded path
	// to verify the middleware-side 206 bypass for any future compressible
	// route that may produce a 206 (e.g. routes_webapp.go).
	for _, prefs := range []string{"gzip", "zstd"} {
		t.Run("Prefs="+prefs, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			conf := config.TestConfig()
			conf.Options().HttpCompression = prefs

			r := gin.New()
			r.Use(NewCompressMiddleware(conf))
			// Mimic what http.ServeContent does for a Range request: set
			// Content-Range, status 206, and write only the requested slice.
			r.GET("/bundle/app.js", func(c *gin.Context) {
				c.Header("Content-Range", "bytes 0-9/100")
				c.Header("Accept-Ranges", "bytes")
				c.Data(http.StatusPartialContent, "application/javascript", []byte("0123456789"))
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/bundle/app.js", nil)
			req.Header.Set("Range", "bytes=0-9")
			req.Header.Set("Accept-Encoding", prefs)
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusPartialContent, w.Code)
			assert.Empty(t, w.Header().Get("Content-Encoding"),
				"206 must not carry Content-Encoding — the slice corresponds to identity bytes")
			assert.Equal(t, "bytes 0-9/100", w.Header().Get("Content-Range"))
			assert.Equal(t, "0123456789", w.Body.String(),
				"206 body must be the raw slice the handler wrote, not encoded bytes")
			assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding",
				"Vary still belongs on the response so caches don't reuse the slice for a non-Range request")
		})
	}
}

func TestNewCompressMiddleware_BypassesPartialContentViaGinStatic(t *testing.T) {
	// End-to-end check that the bypass also works when 206 is produced by
	// gin.Static (i.e. http.ServeFile / http.ServeContent), not just by a
	// hand-crafted handler. The bundled `/static/*` and `/c/static/*` routes
	// are now served by PrecompressedStatic and excluded from the runtime
	// middleware via NewShouldCompressFn, so this test uses a route that is
	// still in the runtime middleware's scope to exercise the same code
	// path on any future compressible static-file route.
	for _, prefs := range []string{"gzip", "zstd"} {
		t.Run("Prefs="+prefs, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			conf := config.TestConfig()
			conf.Options().HttpCompression = prefs

			// Write a real file with deterministic content so we can assert
			// the byte slice that comes back through the wire.
			dir := t.TempDir()
			path := filepath.Join(dir, "app.js")
			alphabet := "0123456789abcdefghijklmnopqrstuvwxyz" // 36 bytes
			content := []byte(strings.Repeat(alphabet, 3))[:100]
			require.Len(t, content, 100)
			require.NoError(t, os.WriteFile(path, content, 0o600))

			r := gin.New()
			r.Use(NewCompressMiddleware(conf))
			// Use a route that is not on the predicate's exclusion list so
			// the middleware's 206 bypass — not the static-route exclusion —
			// is what we're verifying here.
			r.Static("/bundle", dir)

			t.Run("RangeRequestReturns206RawBytes", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/bundle/app.js", nil)
				req.Header.Set("Range", "bytes=10-19")
				req.Header.Set("Accept-Encoding", prefs)
				r.ServeHTTP(w, req)

				require.Equal(t, http.StatusPartialContent, w.Code,
					"http.ServeContent should produce 206 for a satisfiable Range request")
				assert.Empty(t, w.Header().Get("Content-Encoding"),
					"206 from gin.Static must not be re-encoded by the compression middleware")
				assert.Equal(t, "bytes 10-19/100", w.Header().Get("Content-Range"))
				assert.Equal(t, "abcdefghij", w.Body.String(),
					"206 body must be the raw byte slice that http.ServeContent wrote, not encoded")
				assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding",
					"Vary still belongs on the response so caches don't reuse the slice for non-Range requests")
			})

			t.Run("FullRequestStillCompresses", func(t *testing.T) {
				// Sanity check: a non-Range GET on the same route still
				// goes through the encoder, so we know the bypass is targeted
				// at 206 rather than blanket-disabling compression for the path.
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/bundle/app.js", nil)
				req.Header.Set("Accept-Encoding", prefs)
				r.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, prefs, w.Header().Get("Content-Encoding"),
					"non-Range GET on a compressible static asset should still be encoded")
			})
		})
	}
}

func TestNewCompressMiddleware_BypassesErrorResponses(t *testing.T) {
	// Both encoders should bypass for 4xx/5xx and bodyless statuses so the
	// server doesn't burn CPU compressing tiny error payloads (e.g. 429 during
	// a rate-limit storm) and so the encoder doesn't emit a stray frame
	// trailer when the handler writes no body (e.g. 204).
	for _, prefs := range []string{"gzip", "zstd"} {
		t.Run("Prefs="+prefs, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			conf := config.TestConfig()
			conf.Options().HttpCompression = prefs
			r := gin.New()
			r.Use(NewCompressMiddleware(conf))
			r.GET("/rate-limited", func(c *gin.Context) {
				c.String(http.StatusTooManyRequests, `{"error":"too many requests"}`)
			})
			r.GET("/server-error", func(c *gin.Context) {
				c.String(http.StatusServiceUnavailable, `{"error":"unavailable"}`)
			})
			r.GET("/no-content", func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})
			r.GET("/not-modified", func(c *gin.Context) {
				c.Status(http.StatusNotModified)
			})

			cases := []struct {
				name string
				path string
				code int
				body string
			}{
				{"429RateLimited", "/rate-limited", http.StatusTooManyRequests, `{"error":"too many requests"}`},
				{"503ServiceUnavailable", "/server-error", http.StatusServiceUnavailable, `{"error":"unavailable"}`},
				{"204NoContent", "/no-content", http.StatusNoContent, ""},
				{"304NotModified", "/not-modified", http.StatusNotModified, ""},
			}

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					w := httptest.NewRecorder()
					req := httptest.NewRequest(http.MethodGet, tc.path, nil)
					req.Header.Set("Accept-Encoding", prefs)
					r.ServeHTTP(w, req)

					require.Equal(t, tc.code, w.Code)
					assert.Empty(t, w.Header().Get("Content-Encoding"),
						"%s response must not advertise Content-Encoding", tc.name)
					assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding",
						"%s response should still carry Vary so caches stay correct", tc.name)
					assert.Equal(t, tc.body, w.Body.String(),
						"%s body must match the handler's raw bytes", tc.name)
				})
			}
		})
	}
}

func TestNewCompressMiddleware_HeadRequests(t *testing.T) {
	// HEAD requests carry no body — Go's net/http discards body writes for
	// HEAD on the wire — so the encoder must be skipped to save pool slots
	// and CPU. But Vary: Accept-Encoding must still be set when the
	// corresponding GET would be content-coding negotiated, so caches don't
	// reuse the response for clients that asked for a different encoding
	// (RFC 9110 §15.4.5: HEAD may legitimately omit content-determined
	// fields like Content-Length, but Vary still belongs on the response).
	for _, prefs := range []string{"gzip", "zstd", "zstd,gzip"} {
		t.Run("Prefs="+strings.ReplaceAll(prefs, ",", "_"), func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			conf := config.TestConfig()
			conf.Options().HttpCompression = prefs
			r := gin.New()
			r.Use(NewCompressMiddleware(conf))
			r.Any("/ok", func(c *gin.Context) {
				c.String(http.StatusOK, "hello world")
			})
			r.Any(conf.BaseUri(config.ApiUri+"/dl/test"), func(c *gin.Context) {
				c.String(http.StatusOK, "download")
			})

			t.Run("CompressiblePathSetsVaryButSkipsEncoder", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodHead, "/ok", nil)
				req.Header.Set("Accept-Encoding", "zstd, gzip")
				r.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				assert.Empty(t, w.Header().Get("Content-Encoding"),
					"HEAD response must not advertise Content-Encoding (encoder is skipped)")
				assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding",
					"HEAD must still set Vary so caches know the GET varies by Accept-Encoding")
			})

			t.Run("ExcludedPathSetsNoVary", func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodHead, conf.BaseUri(config.ApiUri+"/dl/test"), nil)
				req.Header.Set("Accept-Encoding", "zstd, gzip")
				r.ServeHTTP(w, req)

				require.Equal(t, http.StatusOK, w.Code)
				assert.Empty(t, w.Header().Get("Content-Encoding"))
				assert.Empty(t, w.Header().Get("Vary"),
					"HEAD on a predicate-excluded path does not vary by Accept-Encoding")
			})
		})
	}
}

func TestNewCompressMiddleware_WeakensETagOnEncodedResponses(t *testing.T) {
	// When the response is compressed, strong ETags must be rewritten to
	// W/<value> so caches don't reuse a compressed body for a client that
	// asked for a different content-coding (RFC 9110 §13.2.4). Weak ETags
	// and missing ETags must be left untouched. Bypassed responses
	// (4xx/5xx/204/304) keep their original ETag because the wire bytes
	// match the strong validator.
	cases := []struct {
		name     string
		prefs    string
		accept   string
		status   int
		etagIn   string
		etagOut  string
		encoding string
	}{
		{"GzipWeakensStrongETag", "gzip", "gzip", http.StatusOK, `"abc123"`, `W/"abc123"`, "gzip"},
		{"ZstdWeakensStrongETag", "zstd", "zstd", http.StatusOK, `"abc123"`, `W/"abc123"`, "zstd"},
		{"PreservesAlreadyWeakETag", "gzip", "gzip", http.StatusOK, `W/"abc123"`, `W/"abc123"`, "gzip"},
		{"DoesNothingWhenNoETag", "gzip", "gzip", http.StatusOK, "", "", "gzip"},
		{"PreservesETagOnBypass4xx", "gzip", "gzip", http.StatusBadRequest, `"abc123"`, `"abc123"`, ""},
		{"PreservesETagOnBypass5xx", "zstd", "zstd", http.StatusInternalServerError, `"abc123"`, `"abc123"`, ""},
		{"DoesNothingWhenIdentityNegotiated", "zstd", "br", http.StatusOK, `"abc123"`, `"abc123"`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			conf := config.TestConfig()
			conf.Options().HttpCompression = tc.prefs
			r := gin.New()
			r.Use(NewCompressMiddleware(conf))
			r.GET("/ok", func(c *gin.Context) {
				if tc.etagIn != "" {
					c.Header("ETag", tc.etagIn)
				}
				c.String(tc.status, "payload")
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/ok", nil)
			req.Header.Set("Accept-Encoding", tc.accept)
			r.ServeHTTP(w, req)

			require.Equal(t, tc.status, w.Code)
			assert.Equal(t, tc.etagOut, w.Header().Get("ETag"),
				"ETag header should match expected value after middleware processing")
			assert.Equal(t, tc.encoding, w.Header().Get("Content-Encoding"),
				"Content-Encoding should match expected encoding")
		})
	}
}

func TestNewCompressMiddleware_NoOpWhenDisabled(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(NewCompressMiddleware(nil))
		r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "hi") })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Empty(t, w.Header().Get("Vary"))
		assert.Equal(t, "hi", w.Body.String())
	})

	t.Run("NoneDisablesCompression", func(t *testing.T) {
		r, _ := newCompressTestRouter(t, "none")

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		req.Header.Set("Accept-Encoding", "gzip, zstd")
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
		assert.Empty(t, w.Header().Get("Vary"))
		assert.Equal(t, "hello world", w.Body.String())
	})
}
