package api

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/mcp"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// ServeMCP registers the Model Context Protocol (MCP) Streamable HTTP
// endpoint at /api/v1/mcp. The endpoint is a no-op until the operator
// enables experimental features (--experimental / PHOTOPRISM_EXPERIMENTAL)
// because the route returns early before any handler is attached.
//
//	@Summary	model context protocol endpoint (experimental)
//	@Id			ServeMCP
//	@Tags		MCP
//	@Accept		json
//	@Produce	json
//	@Success	200				{string}	string	"JSON-RPC 2.0 response"
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/mcp [post]
func ServeMCP(router *gin.RouterGroup) {
	if router == nil {
		return
	}

	conf := get.Config()

	// Skip registration when no config is available or experimental mode
	// is off, so /api/v1/mcp returns the standard 404 in that case.
	if conf == nil || !conf.Experimental() {
		return
	}

	// One server instance is shared across all HTTP requests. The SDK
	// isolates concurrent callers through its own session bookkeeping,
	// keyed by the Mcp-Session-Id response header.
	mcpServer := mcp.NewServer(&sdkmcp.Implementation{
		Name:    "photoprism-mcp",
		Version: conf.Version(),
	}, conf.Edition())

	// Streamable HTTP handler. Warn-level logging keeps the default log
	// quiet under normal operation while still surfacing SDK warnings.
	// The 30-minute session timeout matches typical IDE idle windows.
	handler := sdkmcp.NewStreamableHTTPHandler(
		func(r *http.Request) *sdkmcp.Server { return mcpServer },
		&sdkmcp.StreamableHTTPOptions{
			SessionTimeout: 30 * time.Minute,
			Logger:         slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
		},
	)

	// mcpHandler authenticates each request and delegates to the SDK
	// handler. In public mode Session() returns the default public session
	// (treated as admin by the ACL), so the read-only tools registered
	// today are reachable anonymously — intentional to support demo
	// deployments such as demo.photoprism.app. Any future tool that
	// touches per-user state, the database, or mutates anything MUST NOT
	// be registered on this shared server without an additional per-tool
	// gate; see internal/mcp/README.md for the recommended patterns.
	mcpHandler := func(c *gin.Context) {
		s := Auth(c, acl.ResourceMCP, acl.ActionView)

		// Abort writes the matching 401/403/429 response if the session
		// is invalid and returns true so the handler can exit early.
		if s.Abort(c) {
			return
		}

		handler.ServeHTTP(c.Writer, c.Request)
	}

	// Streamable HTTP uses POST for requests, GET for the event stream,
	// and DELETE to tear down a session; register the same handler for
	// all three verbs.
	router.POST("/mcp", mcpHandler)
	router.GET("/mcp", mcpHandler)
	router.DELETE("/mcp", mcpHandler)
}
