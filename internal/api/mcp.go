package api

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mcp"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// McpSessionTimeout configures the idle lifetime of MCP Streamable HTTP
// sessions. It is shorter than a typical IDE editing window so that sessions
// abandoned without a DELETE tear-down do not accumulate SDK-side bookkeeping
// for long periods; active clients renew the idle timer on every request, so
// interactive use is unaffected.
var McpSessionTimeout = 5 * time.Minute

// ServeMCP registers the Model Context Protocol (MCP) Streamable HTTP
// endpoint at /api/v1/mcp.
//
//	@Summary	model context protocol endpoint
//	@Id			ServeMCP
//	@Tags		MCP
//	@Accept		json
//	@Produce	json
//	@Success	200					{string}	string	"JSON-RPC 2.0 response"
//	@Failure	401,403,404,413,429	{object}	i18n.Response
//	@Router		/api/v1/mcp [post]
func ServeMCP(router *gin.RouterGroup) {
	if router == nil {
		return
	}

	conf := get.Config()

	// Skip registration when no config is available, so
	// /api/v1/mcp returns the standard 404 in that case.
	if conf == nil {
		return
	}

	// Skip registration when the MCP endpoint has been disabled via
	// --disable-mcp / PHOTOPRISM_DISABLE_MCP / DisableMCP, so requests
	// to /api/v1/mcp return the standard 404 response.
	if conf.DisableMCP() {
		log.Info("mcp: disabled")
		return
	}

	// One server instance is shared across all HTTP requests. The SDK
	// isolates concurrent callers through its own session bookkeeping,
	// keyed by the Mcp-Session-Id response header.
	mcpServer := mcp.NewServer(&sdkmcp.Implementation{
		Name:    "photoprism-mcp",
		Version: conf.Version(),
	}, conf.Edition())

	// Streamable HTTP handler with warn-level logging and an explicit
	// CrossOriginProtection (go-sdk no longer enables it implicitly).
	handler := sdkmcp.NewStreamableHTTPHandler(
		func(r *http.Request) *sdkmcp.Server { return mcpServer },
		&sdkmcp.StreamableHTTPOptions{
			SessionTimeout:        McpSessionTimeout,
			Logger:                slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
			CrossOriginProtection: &http.CrossOriginProtection{},
		},
	)

	// mcpHandler authenticates each request, caps the JSON-RPC payload size,
	// and surfaces 413 on overflow. Public mode reaches read-only tools as
	// the default public session; any tool that touches per-user state must
	// add an explicit gate before being registered on this shared server.
	mcpHandler := func(c *gin.Context) {
		s := Auth(c, acl.ResourceMCP, acl.ActionView)

		// Abort writes the matching 401/403/429 response if the session
		// is invalid and returns true so the handler can exit early.
		if s.Abort(c) {
			return
		}

		// Reject oversized requests up front when Content-Length is
		// known; chunked or unknown-length bodies fall through to
		// MaxBytesReader below.
		if c.Request.ContentLength > MaxMCPRequestBytes {
			event.AuditWarn([]string{ClientIP(c), "session %s", "mcp", "request body too large", status.Failed}, s.RefID)
			AbortRequestTooLarge(c, i18n.ErrBadRequest)
			return
		}

		// Cap the body before the SDK reads it; the wrapper rewrites the
		// SDK's "failed to read body" 400 into a standard 413 on overflow.
		LimitRequestBodyBytes(c, MaxMCPRequestBytes)

		tripped := &atomic.Bool{}
		c.Request.Body = &mcpLimitReader{ReadCloser: c.Request.Body, tripped: tripped}
		writer := &mcpLimitWriter{ResponseWriter: c.Writer, tripped: tripped}

		handler.ServeHTTP(writer, c.Request)

		// Audit the rewritten 413 so the audit log reflects the final
		// status the client received, not the SDK's internal 400.
		if writer.suppress.Load() {
			event.AuditWarn([]string{ClientIP(c), "session %s", "mcp", "request body too large", status.Failed}, s.RefID)
		}
	}

	// Streamable HTTP uses POST for requests, GET for the event stream,
	// and DELETE to tear down a session; register the same handler for
	// all three verbs.
	router.POST("/mcp", mcpHandler)
	router.GET("/mcp", mcpHandler)
	router.DELETE("/mcp", mcpHandler)
}

// mcpLimitReader wraps an http.MaxBytesReader-bounded body and records
// whether the read failed with *http.MaxBytesError so the paired writer
// can translate the SDK's 400 response into the standard 413.
type mcpLimitReader struct {
	io.ReadCloser
	tripped *atomic.Bool
}

// Read delegates to the wrapped body and flips the shared flag when the
// MaxBytesReader cap is exceeded.
func (r *mcpLimitReader) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)

	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			r.tripped.Store(true)
		}
	}

	return n, err
}

// mcpLimitWriter wraps the outgoing response so the SDK's 400 "failed to
// read body" is rewritten to 413 when the body cap was exceeded. Body
// writes that arrive after the rewrite are suppressed so the SDK's
// internal error phrasing does not replace PhotoPrism's standard 413
// payload.
type mcpLimitWriter struct {
	gin.ResponseWriter
	tripped  *atomic.Bool
	suppress atomic.Bool
}

// WriteHeader rewrites 400 to 413 when the request body exceeded the cap
// and marks the response body as suppressed so subsequent Write calls
// cannot leak the SDK's 400 payload.
func (w *mcpLimitWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusBadRequest && w.tripped.Load() {
		statusCode = http.StatusRequestEntityTooLarge
		w.suppress.Store(true)
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

// Write forwards payload bytes unless the response is a rewritten 413,
// in which case the SDK's body is silently dropped so the final response
// stays consistent with AbortRequestTooLarge.
func (w *mcpLimitWriter) Write(b []byte) (int, error) {
	if w.suppress.Load() {
		return len(b), nil
	}

	return w.ResponseWriter.Write(b)
}
