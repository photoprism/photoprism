package api

import (
	"github.com/gin-gonic/gin"
)

// Override hooks for the OAuth2/OIDC endpoints, following the server.WebDAVHandler
// pattern: CE owns the /api/v1/oauth/* routes, and Portal builds (which cannot
// re-register the same gin paths and which CE cannot import) swap in the OIDC OP
// behavior by reassigning these variables before the server starts serving.

// OAuthAuthorizeHandler handles GET /api/v1/oauth/authorize. Defaults to the app flow;
// Portal builds set it to a dispatcher that also serves the OIDC OP for cluster instances.
var OAuthAuthorizeHandler gin.HandlerFunc = OAuthAppAuthorize

// OAuthUserinfoHandler handles GET and POST /api/v1/oauth/userinfo. Defaults to
// a 405 placeholder; Portal builds set it to the OIDC OP userinfo handler.
var OAuthUserinfoHandler gin.HandlerFunc = defaultOAuthUserinfo

// OAuthLogoutHandler handles GET and POST /api/v1/oauth/logout (OIDC RP-initiated
// logout end-session endpoint). Defaults to a 405 placeholder; Portal builds set it
// to the OIDC OP end-session handler.
var OAuthLogoutHandler gin.HandlerFunc = defaultOAuthLogout

// OAuthAuthorizationCodeHandler handles the authorization_code grant on the shared
// POST /api/v1/oauth/token endpoint. Defaults to the app flow; Portal builds set it
// to a dispatcher that also serves the OIDC OP for cluster instances. OAuthToken
// handles the other grants itself.
var OAuthAuthorizationCodeHandler gin.HandlerFunc = OAuthAppToken
