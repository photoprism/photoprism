package api

import (
	"github.com/gin-gonic/gin"
)

// Override hooks for the OAuth2/OIDC endpoints, following the server.WebDAVHandler
// pattern: CE owns the /api/v1/oauth/* routes, and Portal builds (which cannot
// re-register the same gin paths and which CE cannot import) swap in the OIDC OP
// behavior by reassigning these variables before the server starts serving.

// OAuthAuthorizeHandler handles GET /api/v1/oauth/authorize. Defaults to a 405
// placeholder; Portal builds set it to the OIDC OP authorize handler.
var OAuthAuthorizeHandler gin.HandlerFunc = defaultOAuthAuthorize

// OAuthUserinfoHandler handles GET and POST /api/v1/oauth/userinfo. Defaults to
// a 405 placeholder; Portal builds set it to the OIDC OP userinfo handler.
var OAuthUserinfoHandler gin.HandlerFunc = defaultOAuthUserinfo

// OAuthAuthorizationCodeHandler handles the authorization_code grant on the
// shared POST /api/v1/oauth/token endpoint. Nil on CE/Pro builds (the grant is
// then reported as unsupported); Portal builds set it to the OIDC OP token
// handler. The client_credentials, password, and session grants are always
// handled by OAuthToken itself and stay DB-backed.
var OAuthAuthorizationCodeHandler gin.HandlerFunc
