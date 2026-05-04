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

// statusBypassesCompression reports whether response bodies with the given
// status code should be sent uncompressed. Bypassed:
//
//   - 4xx and 5xx error responses — bodies are typically tiny JSON envelopes
//     and spending CPU compressing them works against the rate limiter when
//     429s are flying (or makes overload worse when the server is already
//     in 5xx territory).
//   - 206 Partial Content — Range responses serve a slice of the identity
//     representation. Compressing them would change the byte sequence the
//     client asked for, breaking the offsets in any subsequent Range
//     continuations and contradicting the Content-Range header. This
//     matters for static assets and service-worker files served by
//     internal/server/routes_static.go and routes_webapp.go.
//   - 204 No Content and 304 Not Modified — there is no body to encode and
//     we don't want the encoder to emit a stray frame trailer.
//
// A status of 0 means the handler hasn't called WriteHeader yet, in which
// case Gin will default to 200 — treat it as compressible.
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

// weakenETag rewrites a strong ETag header (e.g. `"abc123"`) to its weak
// form (e.g. `W/"abc123"`) so caches don't reuse a compressed representation
// for a client that asked for a different content-coding. Already-weak
// ETags and missing ETags are left untouched.
//
// Per RFC 9110 §13.2.4 strong validators must change when the representation
// bytes change, and content-coding negotiation produces distinct byte
// sequences for the same selected representation. Weak ETags are not usable
// for strong-comparison contexts such as `If-Range`, but PhotoPrism already
// excludes download/video/proxy paths from compression so this does not
// regress practical range-download flows.
func weakenETag(h http.Header) {
	etag := h.Get(header.ETag)
	if etag == "" || strings.HasPrefix(etag, "W/") {
		return
	}
	h.Set(header.ETag, "W/"+etag)
}

// compressWriter wraps a Gin ResponseWriter so that body writes pass through
// the configured encoder when the response is eligible for compression. On
// the first Write it inspects the recorded HTTP status and either:
//
//   - bypasses the encoder for error responses (4xx/5xx) and bodyless
//     responses (204/304), removing the Content-Encoding header so the
//     client receives the raw body, or
//   - clears Content-Length and routes subsequent writes through the encoder.
//
// The encoded flag tracks whether any byte actually flowed through the
// encoder so the middleware can suppress the encoder's frame trailer when
// nothing was written.
type compressWriter struct {
	gin.ResponseWriter
	encoder    io.Writer
	encoding   string
	headerDone bool
	bypass     bool
	encoded    bool
}

// Write encodes data through the wrapped encoder, or — for error and
// bodyless responses — bypasses the encoder and writes the raw bytes
// straight to the underlying ResponseWriter. On the first call it also
// clears any handler-supplied Content-Length when compression is taking
// effect (the post-encoding length is unknown up front), weakens any
// strong ETag the handler set so caches don't mix encodings, and removes
// the Content-Encoding header in the bypass path so the client doesn't
// see a stale "gzip"/"zstd" advertisement.
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

// WriteString encodes a string through the wrapped encoder by delegating to
// Write so the bypass and Content-Length-reset logic stays in one place.
func (cw *compressWriter) WriteString(s string) (int, error) {
	return cw.Write([]byte(s))
}

// Flush forwards a flush request to the encoder (if buffered) and then to the
// underlying ResponseWriter so streaming consumers receive partial output.
// In bypass mode it skips the encoder flush, since no encoded bytes are
// pending.
func (cw *compressWriter) Flush() {
	if !cw.bypass {
		if f, ok := cw.encoder.(compressFlusher); ok {
			_ = f.Flush()
		}
	}
	cw.ResponseWriter.Flush()
}

// NewCompressMiddleware returns a Gin middleware that negotiates an HTTP
// content-encoding for each response based on the operator's configured
// preference list and the client's Accept-Encoding header. Supported
// encodings are "gzip" (compress/gzip at DefaultCompression) and "zstd"
// (klauspost/compress/zstd at SpeedDefault).
//
// Requests on connections that are being upgraded (e.g. WebSockets) are
// skipped, and so are paths that the shared NewShouldCompressFn predicate
// excludes (already-compressed media, health endpoints, photo originals,
// theme zip, share previews, portal proxy, etc.). Eligible responses always
// receive a "Vary: Accept-Encoding" header so caches behave correctly even
// when the negotiator selects the identity encoding.
//
// Error responses (4xx/5xx) and bodyless responses (204/304) bypass the
// encoder per statusBypassesCompression so the server doesn't burn CPU on
// small error payloads — this matters most during rate-limit storms (429)
// and when the server is already under pressure (5xx).
//
// When conf is nil or no encodings are configured, the middleware is a
// no-op and behaves identically to passing the request through unchanged.
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

		// Always advertise that Accept-Encoding affects the response, even when
		// the negotiator selects identity, so shared caches keep separate entries.
		// HEAD responses also carry this header per RFC 9110 §15.4.5: the
		// corresponding GET would be content-coding negotiated, and caches need
		// to know that — even though the HEAD itself sends no body and may
		// legitimately omit other content-determined fields like Content-Length.
		c.Writer.Header().Add(header.Vary, header.AcceptEncoding)

		// HEAD requests carry no body — net/http silently discards anything
		// the handler writes. Wrapping c.Writer would consume an encoder pool
		// slot and burn CPU compressing bytes that never reach the wire, so
		// skip the encoder while still advertising that the GET would vary by
		// Accept-Encoding (set above).
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
				// When no body was encoded (bypass path, or handler wrote
				// nothing), drop the optimistic Content-Encoding header and
				// redirect the encoder's trailer to /dev/null so it doesn't
				// land on the wire.
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
