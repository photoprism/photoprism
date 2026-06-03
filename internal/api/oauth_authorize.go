package api

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// OAuth2/OIDC authorize error codes per RFC 6749 §4.1.2.1 and OIDC Core
// §3.1.2.6. Returned as the error query parameter on a trustworthy redirect,
// in the consent page for untrusted requests, or as JSON from the POST handler.
const (
	oauthErrInvalidRequest = "invalid_request"
	oauthErrInvalidClient  = "invalid_client"
	oauthErrAccessDenied   = "access_denied"
	oauthErrServerError    = "server_error"
	oauthErrLoginRequired  = "login_required"
)

// oauthAuthorizeParams holds the parsed OAuth2 authorize-request parameters.
type oauthAuthorizeParams struct {
	ResponseType string
	ClientID     string
	RedirectURI  string
	Scope        string
	State        string
	// Nonce is accepted for OIDC clients but currently ignored: the token
	// endpoint mints an opaque session token, not an id_token, so there is no
	// claim to bind it to. Wire it through once id_tokens are issued.
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
}

// parseOAuthAuthorizeParams extracts the authorize-request parameters from c.
// It reads via Request.FormValue so the same parser works for the GET query
// string and the POST form body (FormValue merges both).
func parseOAuthAuthorizeParams(c *gin.Context) *oauthAuthorizeParams {
	return &oauthAuthorizeParams{
		ResponseType:        strings.TrimSpace(c.Request.FormValue("response_type")),
		ClientID:            strings.TrimSpace(c.Request.FormValue("client_id")),
		RedirectURI:         strings.TrimSpace(c.Request.FormValue("redirect_uri")),
		Scope:               strings.TrimSpace(c.Request.FormValue("scope")),
		State:               c.Request.FormValue("state"),
		Nonce:               c.Request.FormValue("nonce"),
		CodeChallenge:       strings.TrimSpace(c.Request.FormValue("code_challenge")),
		CodeChallengeMethod: strings.TrimSpace(c.Request.FormValue("code_challenge_method")),
	}
}

// validateOAuthAuthorizeRequest enforces the per-parameter rules that may be
// reported back to the client via the (already validated) redirect_uri.
// Client and redirect_uri are validated separately because their failures MUST
// NOT redirect to an untrusted target per RFC 6749 §4.1.2.1.
func validateOAuthAuthorizeRequest(p *oauthAuthorizeParams) error {
	switch {
	case p.ResponseType == "":
		return errors.New("response_type required")
	case p.ResponseType != "code":
		return errors.New("response_type must be code")
	case p.Scope == "":
		return errors.New("scope required")
	case p.State == "":
		return errors.New("state required")
	case p.CodeChallenge == "":
		return errors.New("code_challenge required")
	case p.CodeChallengeMethod == "":
		return errors.New("code_challenge_method required")
	case p.CodeChallengeMethod != authn.PKCEMethodS256:
		return errors.New("code_challenge_method must be S256")
	}
	return nil
}

// oauthRedirectURIPermitted reports whether redirectURI exactly matches one of
// the client's registered redirect URIs. Exact match only (no trailing-slash
// normalization, no wildcards), so a client with no registered URIs can never
// pass and is implicitly excluded from the authorization-code flow.
func oauthRedirectURIPermitted(client *entity.Client, redirectURI string) bool {
	if client == nil || redirectURI == "" {
		return false
	}
	for _, allowed := range client.GetData().RedirectURIs {
		if allowed == redirectURI {
			return true
		}
	}
	return false
}

// buildOAuthRedirect appends a set of query parameters to redirectURI, keeping
// any query string already present in the registered URI.
func buildOAuthRedirect(redirectURI string, params map[string]string) (string, error) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// oauthScopeList splits a space-separated OAuth scope string into its entries.
func oauthScopeList(scope string) []string {
	return strings.Fields(scope)
}

// OAuthAuthorize registers the OAuth2/OIDC authorization endpoint that drives
// the Authorization Code flow with PKCE, see
// https://github.com/photoprism/photoprism/issues/4368.
//
// GET validates the request and renders a self-contained consent page (or, for
// an untrusted client/redirect_uri, an inline error page; for a trustworthy
// redirect_uri, a 302 carrying the error). The consent page reads the existing
// browser session token from namespaced localStorage and, on Allow, calls POST
// with that token; POST authenticates the session, re-validates the request,
// mints a single-use authorization code, and returns the redirect target.
//
//	@Summary	OAuth2 authorization endpoint (authorization code flow with PKCE)
//	@Id			OAuthAuthorize
//	@Tags		Authentication
//	@Produce	html
//	@Produce	json
//	@Param		response_type			query		string	true	"must be code"
//	@Param		client_id				query		string	true	"registered client UID"
//	@Param		redirect_uri			query		string	true	"exact registered redirect URI"
//	@Param		scope					query		string	true	"requested scope"
//	@Param		state					query		string	true	"opaque CSRF state"
//	@Param		code_challenge			query		string	true	"PKCE code challenge"
//	@Param		code_challenge_method	query		string	true	"must be S256"
//	@Success	200						{string}	string	"consent page"
//	@Failure	400,401,403,429			{object}	i18n.Response
//	@Router		/api/v1/oauth/authorize [get]
//	@Router		/api/v1/oauth/authorize [post]
func OAuthAuthorize(router *gin.RouterGroup) {
	router.GET("/oauth/authorize", oauthAuthorizeGet)
	router.POST("/oauth/authorize", oauthAuthorizePost)
}

// oauthAuthorizeGet renders the consent page or an authorize error. It performs
// no authentication: the page resolves the browser session client-side.
func oauthAuthorizeGet(c *gin.Context) {
	// Prevent CDNs from caching this endpoint.
	if header.IsCdn(c.Request) {
		AbortNotFound(c)
		return
	}

	// Disable caching of responses.
	c.Header(header.CacheControl, header.CacheControlNoStore)

	clientIp := ClientIP(c)
	action := "authorize"
	conf := get.Config()

	// Abort if running in public mode (sessions are meaningless then).
	if conf.Public() {
		event.AuditErr([]string{clientIp, "oauth2", "unknown client", action, authn.ErrDisabledInPublicMode.Error()})
		oauthAuthorizeErrorPage(c, http.StatusForbidden, oauthErrInvalidRequest, "disabled in public mode")
		return
	}

	// On Portal builds, a cluster-instance authorize request is served by the
	// OIDC OP flow (direct redirect, no consent page); other clients fall
	// through to the consent flow below.
	if OAuthClusterAuthorize != nil && OAuthClusterAuthorize(c) {
		return
	}

	// Throttle this unauthenticated endpoint to limit client_id/redirect_uri
	// probing. The generous Auth limiter (not the strict Login one) keeps
	// legitimate consent-page reloads working.
	if limiter.Auth.Request(clientIp).Reject() {
		oauthAuthorizeErrorPage(c, http.StatusTooManyRequests, oauthErrInvalidRequest, "too many requests")
		return
	}

	params := parseOAuthAuthorizeParams(c)

	// Validate the client and redirect_uri first: their failures MUST NOT
	// redirect to an untrusted target.
	client, errCode, errDesc := oauthResolveAuthorizeClient(params)
	if client == nil {
		event.AuditWarn([]string{clientIp, "oauth2", "unknown client", action, status.Denied})
		oauthAuthorizeErrorPage(c, http.StatusBadRequest, errCode, errDesc)
		return
	}

	// The redirect_uri is now trustworthy, so remaining validation errors are
	// reported back to the client via a 302 with error + state.
	if err := validateOAuthAuthorizeRequest(params); err != nil {
		dest, buildErr := buildOAuthRedirect(params.RedirectURI, map[string]string{
			"error":             oauthErrInvalidRequest,
			"error_description": err.Error(),
			"state":             params.State,
		})
		if buildErr != nil {
			oauthAuthorizeErrorPage(c, http.StatusBadRequest, oauthErrInvalidRequest, err.Error())
			return
		}
		event.AuditWarn([]string{clientIp, "oauth2", "client %s", action, status.Denied}, clean.Log(client.GetUID()))
		c.Redirect(http.StatusFound, dest)
		return
	}

	// Pre-build the deny redirect server-side so the consent page navigates to
	// a validated target instead of constructing a redirect URL in the browser.
	denyRedirect, err := buildOAuthRedirect(params.RedirectURI, map[string]string{
		"error": oauthErrAccessDenied,
		"state": params.State,
	})
	if err != nil {
		oauthAuthorizeErrorPage(c, http.StatusBadRequest, oauthErrInvalidRequest, "invalid redirect_uri")
		return
	}

	c.HTML(http.StatusOK, "oauth_authorize.gohtml", gin.H{
		"config":           conf.ClientPublic(),
		"clientName":       oauthClientDisplayName(client),
		"scopes":           oauthScopeList(clean.Scope(params.Scope)),
		"denyRedirect":     denyRedirect,
		"storageNamespace": conf.StorageNamespace(),
		"loginUri":         conf.LoginUri(),
	})
}

// oauthAuthorizePost authenticates the browser session, re-validates the
// request, mints a single-use authorization code, and returns the redirect.
func oauthAuthorizePost(c *gin.Context) {
	// Prevent CDNs from caching this endpoint.
	if header.IsCdn(c.Request) {
		AbortNotFound(c)
		return
	}

	// Disable caching of responses.
	c.Header(header.CacheControl, header.CacheControlNoStore)

	clientIp := ClientIP(c)
	action := "authorize"
	conf := get.Config()

	// Abort if running in public mode.
	if conf.Public() {
		event.AuditErr([]string{clientIp, "oauth2", "unknown client", action, authn.ErrDisabledInPublicMode.Error()})
		oauthAuthorizeJSON(c, http.StatusForbidden, oauthErrInvalidRequest, "disabled in public mode")
		return
	}

	// Check request rate limit.
	r := limiter.Login.Request(clientIp)
	if r.Reject() || limiter.Auth.Reject(clientIp) {
		limiter.AbortJSON(c)
		return
	}

	params := parseOAuthAuthorizeParams(c)

	// Resolve the client and redirect_uri before authenticating so a malformed
	// request fails the same way for everyone.
	client, errCode, errDesc := oauthResolveAuthorizeClient(params)
	if client == nil {
		event.AuditWarn([]string{clientIp, "oauth2", "unknown client", action, status.Denied})
		oauthAuthorizeJSON(c, http.StatusBadRequest, errCode, errDesc)
		return
	}
	if err := validateOAuthAuthorizeRequest(params); err != nil {
		event.AuditWarn([]string{clientIp, "oauth2", "client %s", action, status.Denied}, clean.Log(client.GetUID()))
		oauthAuthorizeJSON(c, http.StatusBadRequest, oauthErrInvalidRequest, err.Error())
		return
	}

	// Authenticate the browser session from the request token.
	sess := Session(clientIp, AuthToken(c))
	if sess == nil {
		event.AuditWarn([]string{clientIp, "oauth2", "client %s", action, status.Denied}, clean.Log(client.GetUID()))
		oauthAuthorizeJSON(c, http.StatusUnauthorized, oauthErrLoginRequired, "session required")
		return
	}

	user := sess.GetUser()
	if user == nil || user.IsUnknown() || user.IsDisabled() || !user.IsRegistered() {
		event.AuditWarn([]string{clientIp, "oauth2", "session %s", action, status.Denied}, sess.RefID)
		oauthAuthorizeJSON(c, http.StatusUnauthorized, oauthErrLoginRequired, "session lacks an active user")
		return
	}

	// Mint a single-use authorization code bound to the user, client,
	// redirect_uri, scope, and PKCE challenge.
	raw, _, err := entity.NewOAuthCode(entity.OAuthCodeSpec{
		ClientUID:           client.GetUID(),
		UserUID:             user.UserUID,
		RedirectURI:         params.RedirectURI,
		Scope:               clean.Scope(params.Scope),
		CodeChallenge:       params.CodeChallenge,
		CodeChallengeMethod: params.CodeChallengeMethod,
	})
	if err != nil {
		event.AuditErr([]string{clientIp, "oauth2", "session %s", action, status.Error(err)}, sess.RefID)
		oauthAuthorizeJSON(c, http.StatusInternalServerError, oauthErrServerError, "failed to issue authorization code")
		return
	}

	dest, err := buildOAuthRedirect(params.RedirectURI, map[string]string{
		"code":  raw,
		"state": params.State,
	})
	if err != nil {
		event.AuditErr([]string{clientIp, "oauth2", "session %s", action, status.Error(err)}, sess.RefID)
		oauthAuthorizeJSON(c, http.StatusInternalServerError, oauthErrServerError, "failed to build redirect")
		return
	}

	// Cancel the failure rate-limit reservation after a successful grant.
	r.Success()

	event.AuditInfo([]string{clientIp, "oauth2", "session %s", action, "client %s", status.Granted}, sess.RefID, clean.Log(client.GetUID()))

	c.JSON(http.StatusOK, gin.H{
		"status":   StatusSuccess,
		"redirect": dest,
	})
}

// oauthResolveAuthorizeClient resolves and gate-checks the client and
// redirect_uri shared by the GET and POST handlers. On failure it returns a nil
// client plus the OAuth error code and description to report.
func oauthResolveAuthorizeClient(p *oauthAuthorizeParams) (client *entity.Client, errCode, errDesc string) {
	if p.ClientID == "" {
		return nil, oauthErrInvalidRequest, "client_id required"
	}

	client = entity.FindClientByUID(p.ClientID)
	if client == nil {
		return nil, oauthErrInvalidClient, "unknown client"
	}
	if !client.AuthEnabled {
		return nil, oauthErrInvalidClient, "client authentication disabled"
	}
	if p.RedirectURI == "" {
		return nil, oauthErrInvalidRequest, "redirect_uri required"
	}
	if !oauthRedirectURIPermitted(client, p.RedirectURI) {
		return nil, oauthErrInvalidRequest, "redirect_uri not registered"
	}

	return client, "", ""
}

// oauthClientDisplayName returns a human-readable client name for the consent
// page, falling back to the client UID when no display name is set.
func oauthClientDisplayName(client *entity.Client) string {
	if name := client.Name(); name != "" {
		return name
	}
	return client.GetUID()
}

// oauthAuthorizeErrorPage renders the inline consent-page error variant with
// the given HTTP status. Used when there is no trustworthy redirect_uri.
func oauthAuthorizeErrorPage(c *gin.Context, statusCode int, errCode, errDesc string) {
	c.HTML(statusCode, "oauth_authorize.gohtml", gin.H{
		"config":           get.Config().ClientPublic(),
		"error":            errCode,
		"errorDescription": errDesc,
	})
}

// oauthAuthorizeJSON writes a standard OAuth2 error body and aborts the chain.
func oauthAuthorizeJSON(c *gin.Context, statusCode int, errCode, errDesc string) {
	c.AbortWithStatusJSON(statusCode, gin.H{
		"error":             errCode,
		"error_description": errDesc,
	})
}
