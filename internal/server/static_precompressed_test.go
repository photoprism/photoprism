package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

// staticFixture writes a deterministic identity file plus precompressed
// gzip and zstd siblings into a fresh temp dir and returns the dir path,
// the relative file name, and the raw identity bytes for assertions.
func staticFixture(t *testing.T, name string, size int) (dir, filename string, identity []byte) {
	t.Helper()

	dir = t.TempDir()
	alphabet := "0123456789abcdefghijklmnopqrstuvwxyz"
	identity = bytes.Repeat([]byte(alphabet), (size/len(alphabet))+1)[:size]
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), identity, 0o600))

	var gz bytes.Buffer
	gw := gzip.NewWriter(&gz)
	_, err := gw.Write(identity)
	require.NoError(t, err)
	require.NoError(t, gw.Close())
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".gz"), gz.Bytes(), 0o600))

	zenc, err := zstd.NewWriter(nil)
	require.NoError(t, err)
	zbytes := zenc.EncodeAll(identity, nil)
	require.NoError(t, zenc.Close())
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".zst"), zbytes, 0o600))

	return dir, name, identity
}

// newPrecompressedTestRouter builds a router that mirrors the real
// production wiring: NewCompressMiddleware is registered globally and
// /static/*filepath is served by PrecompressedStatic. The runtime
// middleware should be a no-op for /static/* because the predicate now
// excludes those paths.
func newPrecompressedTestRouter(t *testing.T, prefs, dir string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	conf := config.TestConfig()
	conf.Options().HttpCompression = prefs

	r := gin.New()
	r.Use(NewCompressMiddleware(conf))
	handler := PrecompressedStatic(conf, dir)
	r.GET("/static/*filepath", handler)
	r.HEAD("/static/*filepath", handler)
	return r
}

func TestPrecompressedStatic_ServesZstdSibling(t *testing.T) {
	dir, name, identity := staticFixture(t, "app.js", 4096)
	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
	req.Header.Set("Accept-Encoding", "zstd, gzip")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "zstd", w.Header().Get("Content-Encoding"),
		"zstd preference + zstd-capable client should pick the .zst sibling")
	assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding",
		"Vary belongs on every response that could vary by Accept-Encoding")
	assert.Contains(t, w.Header().Get("Content-Type"), "javascript",
		"Content-Type must reflect the identity filename, not the .zst suffix")

	dec, err := zstd.NewReader(bytes.NewReader(w.Body.Bytes()))
	require.NoError(t, err)
	defer dec.Close()
	out, err := io.ReadAll(dec)
	require.NoError(t, err)
	assert.Equal(t, identity, out, "zstd-decoded body must round-trip to the identity bytes")
}

func TestPrecompressedStatic_ServesGzipFallback(t *testing.T) {
	dir, name, identity := staticFixture(t, "app.css", 4096)
	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"),
		"gzip-only clients fall back to the .gz sibling even when zstd is preferred")
	assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding")

	gr, err := gzip.NewReader(bytes.NewReader(w.Body.Bytes()))
	require.NoError(t, err)
	defer gr.Close()
	out, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, identity, out)
}

func TestPrecompressedStatic_ServesIdentityWhenNoEncodingMatches(t *testing.T) {
	dir, name, identity := staticFixture(t, "app.js", 4096)
	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
	// Client only accepts brotli, which we don't have a sibling for.
	req.Header.Set("Accept-Encoding", "br")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Content-Encoding"),
		"identity fallback must not set Content-Encoding")
	assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding",
		"Vary still belongs on the response since a sibling existed for some encoding")
	assert.Equal(t, identity, w.Body.Bytes(), "body must be the raw identity bytes")
}

func TestPrecompressedStatic_ServesIdentityWhenNoSiblings(t *testing.T) {
	// File without precompressed siblings — happens during dev (watch mode)
	// and for assets that don't compress meaningfully (skipped by the build
	// script's MIN_RATIO threshold).
	dir := t.TempDir()
	identity := []byte(strings.Repeat("0123456789abcdefghijklmnopqrstuvwxyz", 16))
	name := "raw.bin"
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), identity, 0o600))

	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
	req.Header.Set("Accept-Encoding", "zstd, gzip")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Content-Encoding"),
		"no sibling exists, so the response must be identity")
	assert.Empty(t, w.Header().Get("Vary"),
		"Vary must not be set when no precompressed sibling could ever have been served")
	assert.Equal(t, identity, w.Body.Bytes())
}

func TestPrecompressedStatic_RangeRequestServesIdentity(t *testing.T) {
	dir, name, identity := staticFixture(t, "app.js", 256)
	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
	req.Header.Set("Range", "bytes=10-19")
	req.Header.Set("Accept-Encoding", "zstd, gzip")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusPartialContent, w.Code,
		"Range requests on existing siblings must still produce 206")
	assert.Empty(t, w.Header().Get("Content-Encoding"),
		"Range responses must serve the identity slice — byte offsets are over identity bytes")
	assert.Contains(t, w.Header().Get("Vary"), "Accept-Encoding",
		"Vary still belongs on Range responses so caches don't reuse the slice for non-Range requests")
	assert.Equal(t, identity[10:20], w.Body.Bytes(),
		"206 body must be the raw identity slice the client requested")
}

func TestPrecompressedStatic_RuntimeMiddlewareDoesNotDoubleEncode(t *testing.T) {
	dir, name, identity := staticFixture(t, "app.js", 4096)
	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
	req.Header.Set("Accept-Encoding", "zstd, gzip")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "zstd", w.Header().Get("Content-Encoding"),
		"a single Content-Encoding header must be present — no double encoding")
	assert.Len(t, w.Header().Values("Content-Encoding"), 1,
		"Content-Encoding must not be wrapped twice when the runtime middleware is also active")

	// Decode once and confirm the identity bytes pop out — if the runtime
	// middleware had wrapped the body again, this single zstd decode would
	// fail or yield encoded bytes instead of the original.
	dec, err := zstd.NewReader(bytes.NewReader(w.Body.Bytes()))
	require.NoError(t, err)
	defer dec.Close()
	out, err := io.ReadAll(dec)
	require.NoError(t, err, "single zstd decode must succeed; double encoding would fail or return encoded bytes")
	assert.Equal(t, identity, out, "single zstd decode must yield the identity bytes")
}

func TestPrecompressedStatic_NoneDisablesEncodedSelection(t *testing.T) {
	dir, name, identity := staticFixture(t, "app.js", 4096)
	r := newPrecompressedTestRouter(t, "none", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
	req.Header.Set("Accept-Encoding", "zstd, gzip")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Content-Encoding"),
		"PHOTOPRISM_HTTP_COMPRESSION=none must disable precompressed-sibling selection")
	assert.Empty(t, w.Header().Get("Vary"),
		"with compression disabled the response never varies by Accept-Encoding — "+
			"setting Vary would mislead shared caches into keeping redundant per-encoding entries")
	assert.Equal(t, identity, w.Body.Bytes())
}

func TestPrecompressedStatic_PrecompressedResponseHasLastModified(t *testing.T) {
	// http.ServeContent emits Last-Modified on every successful response
	// from a non-zero modtime, which is what enables clients to revalidate
	// later via If-Modified-Since. Confirm it survives the precompressed
	// path so caches don't lose validators when zstd/gzip is selected.
	dir, name, _ := staticFixture(t, "app.js", 4096)

	mtime := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(filepath.Join(dir, name), mtime, mtime))

	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
	req.Header.Set("Accept-Encoding", "zstd")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "zstd", w.Header().Get("Content-Encoding"))
	assert.NotEmpty(t, w.Header().Get("Last-Modified"),
		"precompressed responses must carry Last-Modified so clients can revalidate")
}

func TestPrecompressedStatic_PrecompressedReturns304OnIfModifiedSince(t *testing.T) {
	// Conditional revalidation must work for precompressed siblings; before
	// the http.ServeContent refactor servePrecompressed wrote 200 OK
	// unconditionally, which would have broken cache freshness checks for
	// every bundled JS/CSS asset in the wild.
	dir, name, _ := staticFixture(t, "app.js", 4096)

	mtime := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	require.NoError(t, os.Chtimes(filepath.Join(dir, name), mtime, mtime))

	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	t.Run("ServerSideMatchReturns304", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
		req.Header.Set("Accept-Encoding", "zstd")
		// If-Modified-Since equal to or after mtime — serve must 304.
		req.Header.Set("If-Modified-Since", mtime.Add(time.Hour).UTC().Format(http.TimeFormat))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotModified, w.Code,
			"a fresh If-Modified-Since must short-circuit to 304 even on the precompressed path")
		assert.Empty(t, w.Body.Bytes(), "304 responses must not carry a body")
		// http.ServeContent's writeNotModified strips Content-Encoding /
		// Content-Length / Content-Type; assert that to lock in the contract.
		assert.Empty(t, w.Header().Get("Content-Encoding"),
			"304 must not advertise an encoding for an empty body")
		assert.Empty(t, w.Header().Get("Content-Length"),
			"304 must not advertise a body size for an empty body")
	})

	t.Run("StaleIfModifiedSinceStillReturns200", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
		req.Header.Set("Accept-Encoding", "zstd")
		// If-Modified-Since strictly before mtime — serve must 200 with the
		// encoded body, just like the non-conditional case.
		req.Header.Set("If-Modified-Since", mtime.Add(-time.Hour).UTC().Format(http.TimeFormat))
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code,
			"a stale If-Modified-Since must not suppress the body")
		assert.Equal(t, "zstd", w.Header().Get("Content-Encoding"))
		assert.NotEmpty(t, w.Body.Bytes(),
			"200 responses must carry the encoded body")
	})
}

func TestPrecompressedStatic_ServiceWorkerKeepsJavaScriptMimeType(t *testing.T) {
	// Browsers reject service workers served with a non-JavaScript
	// Content-Type, so registering /static/build/sw.js must keep the .js
	// MIME type when the encoded sibling is selected — otherwise the PWA
	// would silently break on every refresh once the bundle is precompressed.
	// This also covers the importScripts() shim (sw-scope-cleanup.js) and
	// every workbox-*.js helper, since they share the same code path.
	dir, name, _ := staticFixture(t, "sw.js", 4096)
	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/"+name, nil)
	req.Header.Set("Accept-Encoding", "zstd")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "zstd", w.Header().Get("Content-Encoding"))
	assert.Contains(t, w.Header().Get("Content-Type"), "javascript",
		"sw.js Content-Type must be derived from the identity filename, not the .zst suffix — "+
			"otherwise the browser refuses to register the service worker")
}

func TestPrecompressedStatic_HeadRequest(t *testing.T) {
	dir, name, identity := staticFixture(t, "app.js", 4096)
	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/static/"+name, nil)
	req.Header.Set("Accept-Encoding", "zstd")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "zstd", w.Header().Get("Content-Encoding"),
		"HEAD must advertise the same encoding the GET would use")
	assert.NotEmpty(t, w.Header().Get("Content-Length"),
		"HEAD must report the precompressed sibling's size in Content-Length")
	assert.Empty(t, w.Body.Bytes(),
		"HEAD must not write a body even when a sibling exists")
	_ = identity
}

func TestPrecompressedStatic_NotFound(t *testing.T) {
	dir, _, _ := staticFixture(t, "app.js", 4096)
	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/missing.js", nil)
	req.Header.Set("Accept-Encoding", "zstd")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code,
		"unknown identity files must 404, not silently fall through to the sibling")
}

func TestPrecompressedStatic_RejectsPathTraversal(t *testing.T) {
	// Sanity check that http.Dir.Open rejects "..". Without this guard,
	// a crafted /static/..%2Fetc%2Fpasswd could leak files outside the
	// static root.
	dir, _, _ := staticFixture(t, "app.js", 4096)
	// Place a sentinel file outside the static root.
	outside := filepath.Join(filepath.Dir(dir), "secret.txt")
	require.NoError(t, os.WriteFile(outside, []byte("super secret"), 0o600))
	t.Cleanup(func() { _ = os.Remove(outside) })

	r := newPrecompressedTestRouter(t, "zstd,gzip", dir)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/../secret.txt", nil)
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code,
		"path traversal must not return a successful response")
	assert.NotContains(t, w.Body.String(), "super secret",
		"path traversal must not leak file contents from outside the static root")
}

func TestSiblingExists(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "yes.js"), []byte("x"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "subdir"), 0o750))

	fs := http.Dir(dir)

	t.Run("FileExists", func(t *testing.T) {
		assert.True(t, siblingExists(fs, "yes.js"))
	})
	t.Run("FileMissing", func(t *testing.T) {
		assert.False(t, siblingExists(fs, "no.js"))
	})
	t.Run("DirectoryNotConsideredFile", func(t *testing.T) {
		// Directories are valid filesystem entries but not valid siblings.
		assert.False(t, siblingExists(fs, "subdir"))
	})
}

func TestContentTypeForName(t *testing.T) {
	t.Run("Js", func(t *testing.T) {
		assert.Contains(t, contentTypeForName("app.js"), "javascript")
	})
	t.Run("Css", func(t *testing.T) {
		assert.Contains(t, contentTypeForName("app.css"), "css")
	})
	t.Run("UnknownExtensionReturnsEmpty", func(t *testing.T) {
		assert.Equal(t, "", contentTypeForName("blob.xyzzy"))
	})
	t.Run("NoExtensionReturnsEmpty", func(t *testing.T) {
		assert.Equal(t, "", contentTypeForName("README"))
	})
}
