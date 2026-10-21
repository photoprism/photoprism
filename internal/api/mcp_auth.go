package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/auth"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mcp"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// OAuth error codes used in the WWW-Authenticate challenge of the MCP endpoint (RFC 6750).
const (
	McpErrInvalidToken      = "invalid_token"
	McpErrInsufficientScope = "insufficient_scope"
)

// McpChallenge returns the WWW-Authenticate header value with an optional OAuth error code and the scopes a
// request needs; it points MCP clients to the protected resource metadata only while the app flow is enabled,
// and returns an empty string if there is nothing to say.
func McpChallenge(conf *config.Config, errCode string, scopes ...string) string {
	var params []string

	if conf.OAuthAppsEnabled() {
		params = append(params, fmt.Sprintf("resource_metadata=%q", conf.OAuthResourceMetadataUrl()))
	}

	if errCode != "" {
		params = append(params, fmt.Sprintf("error=%q", errCode))
	}

	if len(scopes) > 0 {
		params = append(params, fmt.Sprintf("scope=%q", strings.Join(scopes, " ")))
	}

	if len(params) == 0 {
		return ""
	}

	return header.AuthBearer + " " + strings.Join(params, ", ")
}

// mcpTokenInfo returns the token information the MCP SDK binds sessions to, or nil if the session has
// no stable owner. The expiry is left to Auth(), which also applies the leeway of cluster JWTs.
func mcpTokenInfo(s *entity.Session) *auth.TokenInfo {
	if s == nil {
		return nil
	}

	var userID string

	switch {
	case s.UserUID != "":
		userID = s.UserUID
	case s.ClientUID != "":
		userID = s.ClientUID
	case s.GrantType == authn.GrantJwtBearer.String() && s.AuthIssuer != "" && s.ClientName != "":
		userID = "jwt:" + s.AuthIssuer + " " + s.ClientName
	case s.RefID != "":
		userID = s.RefID
	default:
		return nil
	}

	// The session is passed in process only, so the tools and the middleware can authorize each call.
	return &auth.TokenInfo{
		Scopes: strings.Fields(s.Scope()),
		UserID: userID,
		Extra:  map[string]any{mcp.SessionKey: s},
	}
}

// mcpBearerRequest returns the request with the token that authenticated it as bearer token, so the
// SDK middleware also accepts requests authenticated through the X-Auth-Token or X-Session-ID header.
func mcpBearerRequest(r *http.Request, authToken string) *http.Request {
	if r == nil || authToken == "" {
		return r
	}

	if fields := strings.Fields(r.Header.Get(header.Auth)); len(fields) == 2 && strings.EqualFold(fields[0], header.AuthBearer) && fields[1] == authToken {
		return r
	}

	clone := r.Clone(r.Context())
	clone.Header.Set(header.Auth, header.AuthBearer+" "+authToken)

	return clone
}

// mcpSessionHandler returns a handler that passes the token information of an authenticated
// session to the SDK, which then refuses requests for MCP sessions of another user.
func mcpSessionHandler(handler http.Handler, info *auth.TokenInfo) http.Handler {
	if info == nil {
		return handler
	}

	verifier := func(_ context.Context, _ string, _ *http.Request) (*auth.TokenInfo, error) {
		return info, nil
	}

	return auth.RequireBearerToken(verifier, &auth.RequireBearerTokenOptions{AllowMissingExpiration: true})(handler)
}
