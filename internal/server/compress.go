package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/zstd"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// compressFlusher is the optional interface implemented by encoder writers
// that buffer output internally and need to be flushed before the underlying
// http.ResponseWriter is flushed.
type compressFlusher interface {
	Flush() error
}

// statusBypassesCompression reports whether responses with the given status
// should be sent uncompressed: 4xx/5xx (tiny error bodies aren't worth the CPU),
// 206 Partial Content (Range offsets are over the identity representation), and
// 204/304 (no body to encode). Status 0 means the handler hasn't written yet
// and defaults to 200, which is compressible.
func statusBypassesCompression(status int) bool {
	switch {
	case status >= 400:
		return true
	case status == http.StatusPartialContent:
		return true
	case status == http.StatusNoContent, status == http.StatusNotModified:
		return true
	}
	return false
}

// weakenETag rewrites a strong ETag to its weak form (W/"…") so caches don't
// reuse a compressed representation across content-codings. Per RFC 9110
// §13.2.4 strong validators must change when the bytes change. Already-weak
// or missing ETags are left as-is.
func weakenETag(h http.Header) {
	etag := h.Get(header.ETag)
	if etag == "" || strings.HasPrefix(etag, "W/") {
		return
	}
	h.Set(header.ETag, "W/"+etag)
}

// compressWriter wraps a Gin ResponseWriter and routes body writes through the
// configured encoder. On the first Write it either bypasses the encoder (for
// 4xx/5xx/204/304 — Content-Encoding is removed) or clears Content-Length so
// subsequent writes can stream through. `encoded` tracks whether any byte
// reached the encoder so the middleware can suppress an empty frame trailer.
type compressWriter struct {
	gin.ResponseWriter
	encoder    io.Writer
	encoding   string
	headerDone bool
	bypass     bool
	encoded    bool
}

// Write encodes data through the wrapper or bypasses the encoder for error/
// bodyless responses. The first call also clears Content-Length, weakens any
// strong ETag (so caches don't mix encodings), and removes Content-Encoding
// in the bypass path.
func (cw *compressWriter) Write(p []byte) (int, error) {
	if !cw.headerDone {
		cw.headerDone = true
		if statusBypassesCompression(cw.Status()) {
			cw.bypass = true
			cw.ResponseWriter.Header().Del(header.ContentEncoding)
		} else {
			cw.ResponseWriter.Header().Del(header.ContentLength)
			weakenETag(cw.Header())
		}
	}
	if cw.bypass {
		return cw.ResponseWriter.Write(p)
	}
	cw.encoded = true
	return cw.encoder.Write(p)
}

// WriteString delegates to Write so bypass and Content-Length-reset logic stays in one place.
func (cw *compressWriter) WriteString(s string) (int, error) {
	return cw.Write([]byte(s))
}

// Flush forwards to the encoder (when not bypassed) and then to the underlying
// ResponseWriter so streaming consumers receive partial output.
func (cw *compressWriter) Flush() {
	if !cw.bypass {
		if f, ok := cw.encoder.(compressFlusher); ok {
			_ = f.Flush()
		}
	}
	cw.ResponseWriter.Flush()
}

// NewCompressMiddleware returns a Gin middleware that negotiates gzip or zstd
// content-encoding per request based on the operator's preferences and the
// client's Accept-Encoding header. Connection upgrades, paths excluded by
// NewShouldCompressFn, HEAD requests, and 4xx/5xx/204/206/304 bodies are
// skipped. Eligible responses always set `Vary: Accept-Encoding`. Returns a
// pass-through middleware when conf is nil or no encodings are configured.
func NewCompressMiddleware(conf *config.Config) gin.HandlerFunc {
	if conf == nil {
		return func(c *gin.Context) { c.Next() }
	}

	prefs := conf.HttpCompressionPreferences()
	if len(prefs) == 0 {
		return func(c *gin.Context) { c.Next() }
	}

	shouldCompress := NewShouldCompressFn(conf)

	gzipPool := &sync.Pool{
		New: func() interface{} {
			gz, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
			return gz
		},
	}
	zstdPool := &sync.Pool{
		New: func() interface{} {
			zw, _ := zstd.NewWriter(io.Discard, zstd.WithEncoderLevel(zstd.SpeedDefault))
			return zw
		},
	}

	return func(c *gin.Context) {
		if c == nil || c.Request == nil {
			c.Next()
			return
		}

		// Connection upgrades (e.g. WebSockets) must not be wrapped.
		if strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade") {
			c.Next()
			return
		}

		// Path is excluded from compression — pass through and skip Vary too,
		// since the response shape does not depend on Accept-Encoding here.
		if !shouldCompress(c) {
			c.Next()
			return
		}

		// Vary: Accept-Encoding on every eligible response so shared caches keep
		// separate entries even when the negotiator selects identity. Also set
		// on HEAD per RFC 9110 §15.4.5 since the corresponding GET would vary.
		c.Writer.Header().Add(header.Vary, header.AcceptEncoding)

		// HEAD bodies are discarded by net/http — skip the encoder pool entirely.
		if c.Request.Method == http.MethodHead {
			c.Next()
			return
		}

		switch NegotiateEncoding(c.GetHeader(header.AcceptEncoding), prefs) {
		case EncodingGzip:
			gz := gzipPool.Get().(*gzip.Writer)
			gz.Reset(c.Writer)
			c.Header(header.ContentEncoding, EncodingGzip)
			cw := &compressWriter{ResponseWriter: c.Writer, encoder: gz, encoding: EncodingGzip}
			c.Writer = cw
			defer func() {
				// Nothing encoded — drop Content-Encoding and discard the trailer.
				if !cw.encoded {
					cw.ResponseWriter.Header().Del(header.ContentEncoding)
					gz.Reset(io.Discard)
				}
				_ = gz.Close()
				gzipPool.Put(gz)
			}()
		case EncodingZstd:
			zw := zstdPool.Get().(*zstd.Encoder)
			zw.Reset(c.Writer)
			c.Header(header.ContentEncoding, EncodingZstd)
			cw := &compressWriter{ResponseWriter: c.Writer, encoder: zw, encoding: EncodingZstd}
			c.Writer = cw
			defer func() {
				if !cw.encoded {
					cw.ResponseWriter.Header().Del(header.ContentEncoding)
					zw.Reset(io.Discard)
				}
				_ = zw.Close()
				zstdPool.Put(zw)
			}()
		}

		c.Next()
	}
}
