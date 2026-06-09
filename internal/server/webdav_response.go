package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// ServeWebDAV serves the request using a response writer that normalizes
// selected WebDAV response headers.
func ServeWebDAV(w gin.ResponseWriter, r *http.Request, srv *webdav.Handler) {
	if w == nil || r == nil || srv == nil {
		return
	}

	srv.ServeHTTP(&webDAVResponseWriter{
		ResponseWriter: w,
		method:         r.Method,
	}, r)
}

// webDAVResponseWriter adjusts selected WebDAV response headers.
type webDAVResponseWriter struct {
	gin.ResponseWriter
	method      string
	wroteHeader bool
}

// WriteHeader writes the status code after normalizing headers.
func (w *webDAVResponseWriter) WriteHeader(statusCode int) {
	w.applyWebDAVHeaders(statusCode)
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write writes response bytes after normalizing headers for implicit 200 responses.
func (w *webDAVResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.applyWebDAVHeaders(http.StatusOK)
		w.wroteHeader = true
	}

	return w.ResponseWriter.Write(data)
}

// WriteString writes string data after normalizing headers for implicit 200 responses.
func (w *webDAVResponseWriter) WriteString(s string) (int, error) {
	if !w.wroteHeader {
		w.applyWebDAVHeaders(http.StatusOK)
		w.wroteHeader = true
	}

	return w.ResponseWriter.WriteString(s)
}

// applyWebDAVHeaders adjusts the XML content type for PROPFIND multi-status responses.
func (w *webDAVResponseWriter) applyWebDAVHeaders(statusCode int) {
	if w.method != header.MethodPropfind || statusCode != http.StatusMultiStatus {
		return
	}

	contentType := strings.ToLower(w.ResponseWriter.Header().Get(header.ContentType))
	if strings.HasPrefix(contentType, header.ContentTypeXml) {
		w.ResponseWriter.Header().Set(header.ContentType, "application/xml; charset=utf-8")
	}
}
