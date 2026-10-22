package api

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/oauth"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// OAuthConsentUri is the web app route where users approve or deny an authorization request.
const OAuthConsentUri = "/oauth/consent"

// OAuthAuthorize registers the OAuth2/OIDC authorization endpoint. The behavior
// is provided by OAuthAuthorizeHandler, which defaults to the app flow and is
// replaced by Portal builds with a dispatcher that also serves the cluster OP
// (see oauth_handlers.go and portal/internal/server/register.go).
//
//	@Summary	OAuth2 authorization endpoint for apps (authorization code flow with PKCE)
//	@Id			OAuthAuthorize
//	@Tags		Authentication
//	@Produce	json,html
//	@Param		response_type			query		string				true	"must be code"
//	@Param		client_id				query		string				true	"app client ID or client metadata document URL"
//	@Param		redirect_uri			query		string				true	"registered redirect URI"
//	@Param		code_challenge			query		string				true	"PKCE code challenge"
//	@Param		code_challenge_method	query		string				true	"must be S256"
//	@Param		state					query		string				false	"opaque value returned to the client"
//	@Param		scope					query		string				false	"space-separated scopes, mcp is always included"
//	@Param		resource				query		string				false	"MCP endpoint URL (RFC 8707)"
//	@Success	302						{string}	string				"redirect to the consent page, or to the client redirect_uri with an error"
//	@Failure	400,429					{object}	map[string]string	"OAuth error and error_description (HTML page for browsers)"
//	@Failure	403,404,405				{object}	i18n.Response
//	@Router		/api/v1/oauth/authorize [get]
func OAuthAuthorize(router *gin.RouterGroup) {
	router.GET("/oauth/authorize", func(c *gin.Context) {
		OAuthAuthorizeHandler(c)
	})
}

// OAuthAppAuthorize validates an authorization request of an app client, stores it as a pending request,
// and redirects the browser to the consent page. Errors are rendered as a page until the client and its
// redirect URI are verified, and redirected to the client afterwards.
func OAuthAppAuthorize(c *gin.Context) {
	// Prevent CDNs from caching this endpoint.
	if header.IsCdn(c.Request) {
		AbortNotFound(c)
		return
	}

	// Disable caching of responses.
	c.Header(header.CacheControl, header.CacheControlNoStore)

	conf := get.Config()
	clientIp := ClientIP(c)
	actor := "unknown client"
	action := "authorize"

	// Abort if running in public mode.
	if conf.Public() {
		event.AuditErr([]string{clientIp, "oauth2", actor, action, authn.ErrDisabledInPublicMode.Error()})
		Abort(c, http.StatusForbidden, i18n.ErrForbidden)
		return
	}

	// Abort if apps cannot be connected.
	if !conf.OAuthAppsEnabled() {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"status": StatusFailed})
		return
	}

	// Refuse clients that exceeded the authorization request rate limit.
	if limiter.Authorize.Reject(clientIp) {
		RenderOAuthError(c, http.StatusTooManyRequests, oauth.ErrCodeTemporarilyUnavailable, "rate limit exceeded")
		return
	}

	r := oauth.ParseAuthorizeRequest(c.Request.URL.Query())

	// Until the client and its redirect URI are verified, errors are rendered instead of redirected.
	if err := r.ValidateClient(); err != nil {
		limiter.Authorize.Reserve(clientIp)
		event.AuditWarn([]string{clientIp, "oauth2", actor, action, status.Error(err)})
		RenderOAuthError(c, http.StatusBadRequest, err.Code, err.Description)
		return
	}

	actor = "client " + clean.Log(r.ClientID)

	client, clientErr := oauth.FindClient(conf, r.ClientID, oauth.ClientLookup{AllowFetch: oauthMetadataFetchAllowed(clientIp)})

	switch {
	case errors.Is(clientErr, oauth.ErrMetadataRateLimited):
		event.AuditWarn([]string{clientIp, "oauth2", action, status.Error(clientErr)})
		RenderOAuthError(c, http.StatusTooManyRequests, oauth.ErrCodeTemporarilyUnavailable, "rate limit exceeded")
		return
	case clientErr != nil:
		limiter.Authorize.Reserve(clientIp)
		event.AuditWarn([]string{clientIp, "oauth2", actor, action, status.Error(clientErr)})
		RenderOAuthError(c, http.StatusBadRequest, oauth.ErrCodeInvalidClient, "unknown client")
		return
	case !oauth.MatchRedirectURI(client.RedirectURIs, r.RedirectURI):
		limiter.Authorize.Reserve(clientIp)
		event.AuditWarn([]string{clientIp, "oauth2", actor, action, "redirect uri not registered", status.Denied})
		RenderOAuthError(c, http.StatusBadRequest, oauth.ErrCodeInvalidRequest, "redirect_uri not registered")
		return
	}

	// From here, errors of registered clients are passed to the client with a redirect, while those of clients
	// with a metadata document are rendered, since the caller chooses its redirect URIs.
	fail := func(errCode, description string) {
		if client.Metadata {
			RenderOAuthError(c, http.StatusBadRequest, errCode, description)
		} else {
			oauthErrorRedirect(c, r.RedirectURI, errCode, description, r.State)
		}
	}

	if err := r.Validate(); err != nil {
		limiter.Authorize.Reserve(clientIp)
		event.AuditWarn([]string{clientIp, "oauth2", actor, action, status.Error(err)})
		fail(err.Code, err.Description)
		return
	}

	resource := r.Resource

	if resource == "" {
		resource = conf.OAuthResourceUrl()
	} else if resource != conf.OAuthResourceUrl() {
		limiter.Authorize.Reserve(clientIp)
		event.AuditWarn([]string{clientIp, "oauth2", actor, action, "invalid resource", status.Denied})
		fail(oauth.ErrCodeInvalidTarget, "resource not supported")
		return
	}

	scopes, scopeErr := oauth.ValidateScopes(r.Scopes(), client.Scope)

	if scopeErr != nil {
		limiter.Authorize.Reserve(clientIp)
		event.AuditWarn([]string{clientIp, "oauth2", actor, action, status.Error(scopeErr)})
		fail(scopeErr.Code, scopeErr.Description)
		return
	}

	// Charge the rate limit for each stored request, as the endpoint does not require authentication.
	limiter.Authorize.Reserve(clientIp)

	requestID, err := oauth.NewRequest(oauth.RequestSpec{
		ClientID:            client.ID,
		ClientName:          client.Name,
		RedirectURI:         r.RedirectURI,
		Scopes:              scopes,
		Resource:            resource,
		State:               r.State,
		CodeChallenge:       r.CodeChallenge,
		CodeChallengeMethod: r.CodeChallengeMethod,
	})

	if err != nil {
		event.AuditErr([]string{clientIp, "oauth2", actor, action, status.Error(err)})
		fail(oauth.ErrCodeServerError, "failed to store request")
		return
	}

	event.AuditInfo([]string{clientIp, "oauth2", actor, action, "redirecting to consent", status.Succeeded})
	c.Redirect(http.StatusFound, conf.FrontendUri(OAuthConsentUri)+"?"+url.Values{"request": {requestID}}.Encode())
}

// oauthMetadataFetchAllowed returns a function that charges the authorization rate limit before a client
// metadata document is fetched, and refuses the fetch once the limit is exceeded.
func oauthMetadataFetchAllowed(clientIp string) func() bool {
	return func() bool {
		if limiter.Authorize.Reject(clientIp) {
			return false
		}

		limiter.Authorize.Reserve(clientIp)

		return true
	}
}

// oauthErrorRedirect redirects the browser to a verified redirect URI with an OAuth error and the issuer,
// or renders the error if no redirect URL can be built.
func oauthErrorRedirect(c *gin.Context, redirectURI, errCode, description, state string) {
	dest, err := oauth.ErrorRedirectURL(redirectURI, errCode, description, state, get.Config().SiteUrl())

	if err != nil {
		RenderOAuthError(c, http.StatusBadRequest, errCode, description)
		return
	}

	c.Header(header.CacheControl, header.CacheControlNoStore)
	c.Redirect(http.StatusFound, dest)
	c.Abort()
}
