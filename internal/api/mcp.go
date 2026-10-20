package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/oauth"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mcp"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// newMcpLibraryServer creates the MCP server for authenticated sessions.
var newMcpLibraryServer = mcp.NewLibraryServer

// McpSessionTimeout configures the idle lifetime of MCP Streamable HTTP sessions.
// Sessions abandoned without a DELETE tear-down free their SDK-side bookkeeping
// promptly; active clients renew the timer on every request.
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

	// The command list of the running edition is only available from the CLI app.
	var opt mcp.Options

	if app := conf.CliApp(); app != nil {
		opt.Commands = app.Commands
	}

	// Data tools apply the current feature settings, as the web app's search does.
	opt.Settings = conf.Settings

	// Public mode gets the reference server, authenticated sessions the library server, whose
	// middleware authorizes each tool call. Each instance is shared across HTTP requests; the SDK
	// isolates callers through its session bookkeeping, keyed by the Mcp-Session-Id header.
	implementation := &sdkmcp.Implementation{
		Name:    "photoprism-mcp",
		Version: conf.Version(),
	}

	referenceServer := mcp.NewServer(implementation, conf.Edition(), opt)
	libraryServer := newMcpLibraryServer(implementation, conf.Edition(), opt)

	// Streamable HTTP handler with warn-level logging and an explicit CrossOriginProtection,
	// which go-sdk does not enable implicitly. The option is deprecated in favor of wrapping
	// the handler and is still honored as of go-sdk v1.8.0, so an upgrade has to confirm the
	// check still runs rather than assume it.
	handler := sdkmcp.NewStreamableHTTPHandler(
		func(r *http.Request) *sdkmcp.Server {
			if conf.Public() {
				return referenceServer
			}

			return libraryServer
		},
		&sdkmcp.StreamableHTTPOptions{
			SessionTimeout:        McpSessionTimeout,
			Logger:                slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
			CrossOriginProtection: &http.CrossOriginProtection{},
		},
	)

	// MCP sessions opened with app access tokens end when their grant is revoked or its client disabled.
	grantSessions := newMcpGrantSessions(libraryServer)
	entity.GrantDeletedHandler = grantSessions.Close
	entity.GrantSuspendedHandler = grantSessions.Close

	// mcpHandler authenticates each request, checks the account setting, caps the JSON-RPC payload size,
	// and asks for the scopes a tool call needs. Public mode reaches the reference server as the default
	// public session; data tools exist only on the library server, which authorizes each call.
	mcpHandler := func(c *gin.Context) {
		s := AuthAudience(c, conf.OAuthResourceUrl(), acl.ResourceMCP, acl.ActionView)

		// Point clients to the protected resource metadata, so they can start the app flow.
		if s.Status == http.StatusUnauthorized && conf.OAuthAppsEnabled() {
			if AuthToken(c) == "" {
				c.Header(header.WWWAuthenticate, McpChallenge(conf, ""))
			} else {
				c.Header(header.WWWAuthenticate, McpChallenge(conf, McpErrInvalidToken))
			}
		}

		// Abort writes the matching 401/403/429 response if the session
		// is invalid and returns true so the handler can exit early.
		if s.Abort(c) {
			return
		}

		// Sessions that act as a user also need the account's MCP setting, except the public session.
		if u := s.GetUser(); !conf.Public() && s.HasUser() && !u.CanUseMCP() {
			event.AuditErr([]string{ClientIP(c), "session %s", "mcp", "disabled for user %s", status.Denied}, s.RefID, clean.Log(u.Username()))
			AbortForbidden(c)
			return
		}

		// Reject oversized requests up front when Content-Length is known; other bodies are capped below.
		if c.Request.ContentLength > MaxMCPRequestBytes {
			event.AuditWarn([]string{ClientIP(c), "session %s", "mcp", "request body too large", status.Failed}, s.RefID)
			AbortRequestTooLarge(c, i18n.ErrBadRequest)
			return
		}

		// Read the capped body of a request, so it can be checked before the SDK runs it; cap other bodies as well.
		if c.Request.Method != http.MethodPost {
			LimitRequestBodyBytes(c, MaxMCPRequestBytes)
		} else {
			calls, ok := mcpReadCalls(c, s)

			if !ok {
				return
			}

			// Ask OAuth clients to re-authorize for the scopes a tool call needs.
			if missing := mcpMissingScopes(s, calls, conf.Public()); len(missing) > 0 {
				event.AuditWarn([]string{ClientIP(c), "session %s", "mcp", "call needs scope %s", status.Denied}, s.RefID, clean.Log(strings.Join(missing, " ")))
				c.Header(header.WWWAuthenticate, McpChallenge(conf, McpErrInsufficientScope, mcpChallengeScopes(s, missing)...))
				AbortForbidden(c)
				return
			}
		}

		// Bind MCP sessions to the authenticated user or client, except in public mode.
		if conf.Public() {
			handler.ServeHTTP(c.Writer, c.Request)
		} else if info := mcpTokenInfo(s); info == nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", "mcp", "session has no owner", status.Denied}, s.RefID)
			AbortForbidden(c)
			return
		} else {
			info.Extra[mcp.ClientIPKey] = ClientIP(c)
			mcpSessionHandler(handler, info).ServeHTTP(c.Writer, mcpBearerRequest(c.Request, AuthToken(c)))
		}

		// Remember which grant the MCP session of an app access token belongs to.
		if s.IsOAuthApp() && c.Writer.Status() < http.StatusBadRequest {
			grantSessions.Add(c.Writer.Header().Get(sdkSessionIdHeader), s.AuthID, mcpGrantActive)
		}

		// Audit requests the SDK refused, e.g. for the MCP session of another user.
		if c.Writer.Status() == http.StatusForbidden {
			event.AuditWarn([]string{ClientIP(c), "session %s", "mcp", "request refused", status.Denied}, s.RefID)
		}
	}

	// Streamable HTTP uses POST for requests, GET for the event stream,
	// and DELETE to tear down a session; register the same handler for
	// all three verbs.
	router.POST("/mcp", mcpHandler)
	router.GET("/mcp", mcpHandler)
	router.DELETE("/mcp", mcpHandler)
}

// MaxMCPBatchMessages is the maximum number of messages in a JSON-RPC batch, which the SDK still accepts for
// protocol version 2025-03-26.
const MaxMCPBatchMessages = 10

// mcpCall is the part of a JSON-RPC message that identifies a tool call.
type mcpCall struct {
	Method string `json:"method"`
	Params struct {
		Name string `json:"name"`
	} `json:"params"`
}

// mcpReadCalls reads at most MaxMCPRequestBytes of the request body, restores it, and returns the JSON-RPC messages
// in it, which may be a batch. It aborts and returns false if the body is too large or cannot be read, is not a
// single JSON value, or holds more than MaxMCPBatchMessages.
func mcpReadCalls(c *gin.Context, s *entity.Session) (calls []mcpCall, ok bool) {
	LimitRequestBodyBytes(c, MaxMCPRequestBytes)

	body, err := io.ReadAll(c.Request.Body)

	if err != nil {
		if IsRequestBodyTooLarge(err) {
			event.AuditWarn([]string{ClientIP(c), "session %s", "mcp", "request body too large", status.Failed}, s.RefID)
			AbortRequestTooLarge(c, i18n.ErrBadRequest)
		} else {
			AbortBadRequest(c, err)
		}

		return nil, false
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	trimmed := bytes.TrimSpace(body)

	if len(trimmed) == 0 {
		return nil, true
	}

	// Decode the first JSON value, as the SDK does, and refuse data after it, so both read the same messages.
	var first json.RawMessage

	dec := json.NewDecoder(bytes.NewReader(trimmed))

	if err = dec.Decode(&first); err != nil || len(bytes.TrimSpace(trimmed[dec.InputOffset():])) > 0 {
		AbortBadRequest(c, errors.New("malformed request body"))
		return nil, false
	}

	if first[0] == '[' {
		var batch []json.RawMessage

		if err = json.Unmarshal(first, &batch); err != nil {
			AbortBadRequest(c, errors.New("malformed request body"))
			return nil, false
		} else if len(batch) > MaxMCPBatchMessages {
			event.AuditWarn([]string{ClientIP(c), "session %s", "mcp", "batch of %d messages", status.Denied}, s.RefID, len(batch))
			AbortBadRequest(c, fmt.Errorf("batch must not exceed %d messages", MaxMCPBatchMessages))
			return nil, false
		}

		for _, raw := range batch {
			var call mcpCall

			if json.Unmarshal(raw, &call) == nil {
				calls = append(calls, call)
			}
		}
	} else {
		var call mcpCall

		if json.Unmarshal(first, &call) == nil {
			calls = append(calls, call)
		}
	}

	return calls, true
}

// mcpMissingScopes returns the scopes the session's token lacks for the tool calls, if its role and account may
// use the tools; none in public mode, which serves only the reference tools. The middleware of the library server
// authorizes each call again, so this check only lets OAuth clients ask for the scopes in advance.
func mcpMissingScopes(s *entity.Session, calls []mcpCall, public bool) (scopes []string) {
	if public {
		return nil
	}

	for _, call := range calls {
		if call.Method != "tools/call" {
			continue
		}

		if decision, needed := mcp.AuthorizeTool(s, call.Params.Name); decision == mcp.ToolMissingScope {
			for _, scope := range needed {
				if !slices.Contains(scopes, scope) {
					scopes = append(scopes, scope)
				}
			}
		}
	}

	return scopes
}

// mcpChallengeScopes returns the scopes an OAuth client should request: the app scopes its token already has
// and the missing ones, so that re-authorizing does not drop the scopes of other tools.
func mcpChallengeScopes(s *entity.Session, missing []string) (scopes []string) {
	for _, scope := range strings.Fields(s.Scope()) {
		if slices.Contains(oauth.AppScopes, scope) && !slices.Contains(scopes, scope) {
			scopes = append(scopes, scope)
		}
	}

	for _, scope := range missing {
		if !slices.Contains(scopes, scope) {
			scopes = append(scopes, scope)
		}
	}

	return scopes
}
