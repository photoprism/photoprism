package mcp

import (
	"context"
	"fmt"
	"sync"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// Keys of the values the HTTP handler passes in the Extra map of the SDK's token info.
const (
	// SessionKey is the key of the authenticated *entity.Session.
	SessionKey = "photoprism.session"
	// ClientIPKey is the key of the client IP address of the request.
	ClientIPKey = "photoprism.ip"
)

// ToolAccess declares what a tool call requires: an ACL resource and action, and the token scope.
type ToolAccess struct {
	Resource acl.Resource
	Action   acl.Permission
	// Scope is the token scope the call needs besides mcp; empty means the resource name.
	Scope string
	// Static marks reference data that the public and stdio servers may offer.
	Static bool
}

// TokenScope returns the token scope the call needs besides mcp.
func (a ToolAccess) TokenScope() string {
	if a.Scope != "" {
		return a.Scope
	}

	return a.Resource.String()
}

// TokenScopes returns the scopes a token needs to call the tool.
func (a ToolAccess) TokenScopes() []string {
	if scope := a.TokenScope(); scope != acl.ResourceMCP.String() {
		return []string{acl.ResourceMCP.String(), scope}
	}

	return []string{acl.ResourceMCP.String()}
}

// staticAccess is the declaration of the reference tools.
var staticAccess = ToolAccess{Resource: acl.ResourceMCP, Action: acl.ActionView, Static: true}

var (
	toolAccess   = make(map[string]ToolAccess)
	toolAccessMu sync.RWMutex
)

// AddTool registers a tool on the server together with its declaration, which every tool must have.
func AddTool[In, Out any](server *sdkmcp.Server, access ToolAccess, tool *sdkmcp.Tool, handler sdkmcp.ToolHandlerFor[In, Out]) {
	declareTool(tool.Name, access)
	sdkmcp.AddTool(server, tool, handler)
}

// declareTool records the declaration of a tool; a name declared twice must keep its declaration.
func declareTool(name string, access ToolAccess) {
	toolAccessMu.Lock()
	defer toolAccessMu.Unlock()

	if found, ok := toolAccess[name]; ok && found != access {
		panic(fmt.Sprintf("mcp: tool %s is already declared with other access", name))
	}

	toolAccess[name] = access
}

// ToolDeclaration returns the declaration of a registered tool.
func ToolDeclaration(name string) (ToolAccess, bool) {
	toolAccessMu.RLock()
	defer toolAccessMu.RUnlock()

	access, ok := toolAccess[name]

	return access, ok
}

// ToolDecision is the result of the per-tool authorization check.
type ToolDecision int

// Per-tool authorization results.
const (
	ToolAllowed ToolDecision = iota
	ToolRefused
	ToolMissingScope
)

// AuthorizeTool decides whether the session may call the tool. It returns ToolRefused if the tool has no
// declaration, or the role or account may not use it, and ToolMissingScope with the scopes the token needs
// if only the scope is missing. It writes no audit entry, so listing tools does not log every decision.
func AuthorizeTool(s *entity.Session, name string) (ToolDecision, []string) {
	access, ok := ToolDeclaration(name)

	if !ok || s == nil {
		return ToolRefused, nil
	}

	// The role must allow MCP and the declared action, for a client session also the client role.
	if !s.Grants(acl.ResourceMCP, acl.ActionView) || !s.Grants(access.Resource, access.Action) {
		return ToolRefused, nil
	}

	// A session that acts as a user also needs the account's MCP setting.
	if s.HasUser() && !s.GetUser().CanUseMCP() {
		return ToolRefused, nil
	}

	// The scope of a client token, and of a user session that has one, must permit mcp and the tool's scope.
	if s.IsClient() || s.HasScope() {
		perms := acl.Permissions{access.Action}

		if !s.ValidateScope(acl.ResourceMCP, perms) || !s.ValidateScope(acl.Resource(access.TokenScope()), perms) {
			return ToolMissingScope, access.TokenScopes()
		}
	}

	return ToolAllowed, nil
}

// RequestSession returns the authenticated session of the current request, or nil if there is none.
func RequestSession(req sdkmcp.Request) *entity.Session {
	if req == nil {
		return nil
	}

	extra := req.GetExtra()

	if extra == nil || extra.TokenInfo == nil || extra.TokenInfo.Extra == nil {
		return nil
	}

	s, _ := extra.TokenInfo.Extra[SessionKey].(*entity.Session)

	return s
}

// RequestClientIP returns the client IP address of the current request, or an empty string if it is unknown.
func RequestClientIP(req sdkmcp.Request) string {
	if req == nil {
		return ""
	}

	extra := req.GetExtra()

	if extra == nil || extra.TokenInfo == nil || extra.TokenInfo.Extra == nil {
		return ""
	}

	ip, _ := extra.TokenInfo.Extra[ClientIPKey].(string)

	return clean.IP(ip, "")
}

// toolNotAvailable returns the error result of a refused tool call, which does not say why it was refused.
func toolNotAvailable(name string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		IsError: true,
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: fmt.Sprintf("tool %s is not available", clean.LogQuote(name))}},
	}
}

// authorizeTools is the receiving middleware that checks each tools/call against the tool's declaration
// and removes the tools the caller's role or account cannot use from tools/list.
func authorizeTools(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
	return func(ctx context.Context, method string, req sdkmcp.Request) (sdkmcp.Result, error) {
		switch method {
		case "tools/call":
			s := RequestSession(req)
			name := ""

			if call, ok := req.(*sdkmcp.CallToolRequest); ok && call.Params != nil {
				name = call.Params.Name
			}

			if decision, _ := AuthorizeTool(s, name); decision != ToolAllowed {
				// Refused calls of declared tools are audited; calls of undeclared names get the same
				// result without an audit entry.
				if _, declared := ToolDeclaration(name); declared && s != nil {
					clientIp := RequestClientIP(req)
					event.AuditWarn([]string{clientIp, "session %s", "mcp", "call %s", status.Denied}, s.RefID, clean.Log(name))
				}

				return toolNotAvailable(name), nil
			}
		case "tools/list":
			result, err := next(ctx, method, req)

			if list, ok := result.(*sdkmcp.ListToolsResult); ok && err == nil {
				s := RequestSession(req)
				tools := make([]*sdkmcp.Tool, 0, len(list.Tools))

				for _, tool := range list.Tools {
					if decision, _ := AuthorizeTool(s, tool.Name); decision != ToolRefused {
						tools = append(tools, tool)
					}
				}

				list.Tools = tools
			}

			return result, err
		}

		return next(ctx, method, req)
	}
}
