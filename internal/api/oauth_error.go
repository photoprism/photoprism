package api

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// oauthWantsHTML reports whether an OAuth/OIDC error should render as a branded
// HTML page rather than JSON: true for a top-level browser navigation
// (Sec-Fetch-Mode: navigate or Accept: text/html), false for API clients.
func oauthWantsHTML(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}
	if strings.Contains(strings.ToLower(c.GetHeader(header.FetchMode)), "navigate") {
		return true
	}
	return strings.Contains(strings.ToLower(c.GetHeader(header.Accept)), gin.MIMEHTML)
}

// OAuthWantsHTML reports whether the request is a top-level browser navigation
// that should receive an HTML response instead of JSON, so OP handlers in
// extension builds can apply the same content negotiation as the shared
// OAuth/OIDC error helpers.
func OAuthWantsHTML(c *gin.Context) bool {
	return oauthWantsHTML(c)
}

// RenderOAuthError responds to a non-redirectable OAuth/OIDC error with a branded
// HTML page for browsers or the standard JSON body for API clients. Use it only
// when there is no trusted redirect_uri (RFC 6749 §4.1.2.1 forbids redirecting to
// an unverified URI).
func RenderOAuthError(c *gin.Context, statusCode int, errCode, errDescription string) {
	if c == nil {
		return
	}

	c.Header(header.CacheControl, header.CacheControlNoStore)

	if oauthWantsHTML(c) {
		c.HTML(statusCode, "oauth-error.gohtml", gin.H{
			"config":            get.Config().ClientPublic(),
			"code":              statusCode,
			"error":             errCode,
			"error_description": errDescription,
		})
		c.Abort()
		return
	}

	c.AbortWithStatusJSON(statusCode, gin.H{
		"error":             errCode,
		"error_description": errDescription,
	})
}

// RedirectOAuthError sends the browser back to a validated redirect_uri with the
// standard error, error_description, and echoed state (RFC 6749 §4.1.2.1). Callers
// MUST have validated redirectURI first; an unparseable URI falls back to
// RenderOAuthError rather than being followed.
func RedirectOAuthError(c *gin.Context, redirectURI, state, errCode, errDescription string) {
	if c == nil {
		return
	}

	u, err := url.Parse(redirectURI)
	if err != nil || u.Scheme == "" || u.Host == "" {
		RenderOAuthError(c, http.StatusBadRequest, "invalid_request", "invalid redirect_uri")
		return
	}

	q := u.Query()
	q.Set("error", errCode)
	if errDescription != "" {
		q.Set("error_description", errDescription)
	}
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()

	c.Header(header.CacheControl, header.CacheControlNoStore)
	c.Redirect(http.StatusFound, u.String())
	c.Abort()
}
