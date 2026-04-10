package api

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/internal/auth/acl"
	internalmcp "github.com/photoprism/photoprism/internal/mcp"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// ServeMCP registers the experimental MCP Streamable HTTP endpoint at /mcp.
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

	if conf == nil || !conf.Experimental() {
		return
	}

	mcpServer := internalmcp.NewServer(&sdkmcp.Implementation{
		Name:    "photoprism-mcp",
		Version: conf.Version(),
	}, conf.Edition())

	handler := sdkmcp.NewStreamableHTTPHandler(
		func(r *http.Request) *sdkmcp.Server { return mcpServer },
		&sdkmcp.StreamableHTTPOptions{
			SessionTimeout: 30 * time.Minute,
			Logger:         slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
		},
	)

	mcpHandler := func(c *gin.Context) {
		s := Auth(c, acl.ResourceMCP, acl.ActionView)

		if s.Invalid() || conf.Public() {
			AbortForbidden(c)
			return
		}

		handler.ServeHTTP(c.Writer, c.Request)
	}

	router.POST("/mcp", mcpHandler)
	router.GET("/mcp", mcpHandler)
	router.DELETE("/mcp", mcpHandler)
}
