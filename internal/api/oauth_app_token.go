package api

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/oauth"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// OAuthAppTokenResponse represents the token response of the OAuth app flow.
type OAuthAppTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// OAuthAppToken redeems an authorization code of an app client for an access token and a refresh token.
// It serves the authorization_code grant of POST /api/v1/oauth/token, see OAuthAuthorizationCodeHandler.
func OAuthAppToken(c *gin.Context) {
	clientIp := ClientIP(c)
	action := "redeem code"

	client, ok := oauthAppTokenClient(c, action, false, func() error { return oauth.VerifyCode(c.PostForm("code")) })

	if !ok {
		return
	}

	actor := "client " + clean.Log(client.ID)

	code, err := oauth.RedeemCode(c.PostForm("code"), client.ID, c.PostForm("redirect_uri"), c.PostForm("code_verifier"))

	if err != nil {
		event.AuditWarn([]string{clientIp, "oauth2", "%s", "%s", status.Error(err)}, actor, action)
		oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "invalid authorization code")
		return
	}

	// The code is only valid if its redirect URI is still registered and its resource was requested.
	if !oauth.MatchRedirectURI(client.RedirectURIs, code.RedirectURI) {
		event.AuditWarn([]string{clientIp, "oauth2", "%s", "%s", "redirect uri not registered", status.Denied}, actor, action)
		oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "invalid authorization code")
		return
	} else if resource := c.PostForm("resource"); resource != "" && resource != code.Resource {
		event.AuditWarn([]string{clientIp, "oauth2", "%s", "%s", "invalid resource", status.Denied}, actor, action)
		oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidTarget, "resource does not match")
		return
	}

	user, ok := oauthAppTokenUser(c, code.UserUID, actor, action)

	if !ok {
		return
	}

	// Limit the grant to the scopes the client may still request.
	scope := oauth.ClientScopes(strings.Fields(code.Scope), client.Scope)

	if !slices.Contains(scope, oauth.DefaultScope) {
		event.AuditWarn([]string{clientIp, "oauth2", "%s", "%s", "client scope no longer includes %s", status.Denied}, actor, action, oauth.DefaultScope)
		oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "scope no longer allowed")
		return
	}

	code.Scope = oauth.JoinScopes(scope)

	grant, refreshToken, err := oauth.NewGrant(code, user, client.Name)

	if err != nil {
		event.AuditErr([]string{clientIp, "oauth2", "%s", "%s", status.Error(err)}, actor, action)
		oauthTokenError(c, http.StatusInternalServerError, oauth.ErrCodeServerError, "failed to create grant")
		return
	}

	oauthAppTokenResponse(c, grant, user, refreshToken, authn.GrantAuthorizationCode, actor, action)
}

// OAuthAppRefresh issues a new access token and refresh token for the grant of an app client's refresh token.
// It serves the refresh_token grant of POST /api/v1/oauth/token.
func OAuthAppRefresh(c *gin.Context) {
	clientIp := ClientIP(c)
	action := "refresh token"

	client, ok := oauthAppTokenClient(c, action, true, func() error { return oauth.VerifyRefreshToken(c.PostForm("refresh_token")) })

	if !ok {
		return
	}

	actor := "client " + clean.Log(client.ID)
	raw := c.PostForm("refresh_token")
	current, err := oauth.FindRefreshGrant(raw)

	if err != nil && !errors.Is(err, oauth.ErrInvalidGrant) {
		event.AuditErr([]string{clientIp, "oauth2", "%s", "%s", status.Error(err)}, actor, action)
		oauthTokenError(c, http.StatusInternalServerError, oauth.ErrCodeServerError, "failed to refresh token")
		return
	} else if err != nil {
		oauthChargeMissingGrant(clientIp, err)
		event.AuditWarn([]string{clientIp, "oauth2", "%s", "%s", status.Error(err)}, actor, action)
		oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "invalid refresh token")
		return
	}

	// Check the user and the client scope first, so that a refusal does not consume the refresh token.
	var user *entity.User

	if current.ClientID == client.ID {
		if user, ok = oauthAppTokenUser(c, current.UserUID, actor, action); !ok {
			return
		} else if !slices.Contains(oauth.ClientScopes(strings.Fields(current.Scope), client.Scope), oauth.DefaultScope) {
			event.AuditWarn([]string{clientIp, "oauth2", "%s", "%s", "client scope no longer includes %s", status.Denied}, actor, action, oauth.DefaultScope)
			oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "scope no longer allowed")
			return
		}
	}

	// RefreshGrant succeeds only for the current grant of this client, whose user was checked above.
	grant, refreshToken, err := oauth.RefreshGrant(current, raw, client.ID)

	switch {
	case errors.Is(err, oauth.ErrRefreshTokenReused):
		event.AuditWarn([]string{clientIp, "oauth2", "%s", "user %s", "%s", "refresh token reused, grant %s revoked", status.Denied}, actor, clean.Log(grant.UserName), action, clean.Log(grant.GrantUID))
		oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "invalid refresh token")
		return
	case err != nil:
		event.AuditWarn([]string{clientIp, "oauth2", "%s", "%s", status.Error(err)}, actor, action)
		oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "invalid refresh token")
		return
	case user == nil:
		oauthTokenError(c, http.StatusInternalServerError, oauth.ErrCodeServerError, "failed to refresh token")
		return
	}

	// Keep only the latest earlier access token, which requests in flight may still use.
	if err = entity.DeleteGrantSessionsExceptLatest(grant.GrantUID); err != nil {
		event.AuditErr([]string{clientIp, "oauth2", "%s", "%s", status.Error(err)}, actor, action)
	}

	// Limit the access token to the scopes the client may still request.
	scope := oauth.ClientScopes(strings.Fields(grant.Scope), client.Scope)

	if !slices.Contains(scope, oauth.DefaultScope) {
		event.AuditWarn([]string{clientIp, "oauth2", "%s", "%s", "client scope no longer includes %s", status.Denied}, actor, action, oauth.DefaultScope)
		oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "scope no longer allowed")
		return
	}

	grant.Scope = oauth.JoinScopes(scope)

	oauthAppTokenResponse(c, grant, user, refreshToken, authn.GrantRefreshToken, actor, action)
}

// oauthAppTokenClient returns the app client of a token request, which authenticates with its client ID only,
// or responds with an error. The optional verify function checks the presented code or refresh token before the
// client is looked up. With allowStale, an expired client metadata document is used while it cannot be fetched.
func oauthAppTokenClient(c *gin.Context, action string, allowStale bool, verify func() error) (*oauth.Client, bool) {
	clientIp := ClientIP(c)
	conf := get.Config()

	if !conf.OAuthAppsEnabled() {
		if limiter.Authorize.Reject(clientIp) {
			oauthTokenError(c, http.StatusTooManyRequests, oauth.ErrCodeTemporarilyUnavailable, "rate limit exceeded")
		} else {
			limiter.Authorize.Reserve(clientIp)
			oauthTokenError(c, http.StatusBadRequest, "unsupported_grant_type", "grant_type is not supported")
		}

		return nil, false
	}

	clientID := c.PostForm("client_id")
	basicID, basicSecret, basic := c.Request.BasicAuth()

	// App clients are public and authenticate without a client secret.
	if basic && (basicSecret != "" || clientID != "" && clientID != basicID) || c.PostForm("client_secret") != "" {
		log.Debugf("oauth2: client %s sent a client secret (%s)", clean.Log(clientID), action)
		event.AuditWarn([]string{clientIp, "oauth2", "%s", "client secret not supported", status.Denied}, action)
		oauthTokenError(c, http.StatusUnauthorized, oauth.ErrCodeInvalidClient, "invalid client")
		return nil, false
	} else if basic {
		clientID = basicID
	}

	// Refuse an address over the authorization failure limit before looking up a registered client ID.
	registered := !strings.HasPrefix(clientID, "https://")

	if registered && limiter.Authorize.Reject(clientIp) {
		oauthTokenError(c, http.StatusTooManyRequests, oauth.ErrCodeTemporarilyUnavailable, "rate limit exceeded")
		return nil, false
	}

	if verify != nil {
		if err := verify(); err != nil {
			event.AuditWarn([]string{clientIp, "oauth2", "%s", status.Error(err)}, action)
			oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "invalid grant")
			return nil, false
		}
	}

	client, err := oauth.FindClient(conf, clientID, oauth.ClientLookup{AllowFetch: oauthMetadataFetchAllowed(clientIp), AllowStale: allowStale})

	switch {
	case errors.Is(err, oauth.ErrMetadataRateLimited):
		oauthTokenError(c, http.StatusTooManyRequests, oauth.ErrCodeTemporarilyUnavailable, "rate limit exceeded")
		return nil, false
	case err != nil:
		// A registered client ID that resolves to no app client counts towards limiter.Authorize, and only a
		// counted refusal logs the client ID.
		if registered {
			limiter.Authorize.Reserve(clientIp)
			event.AuditWarn([]string{clientIp, "oauth2", "client %s", "%s", status.Error(err)}, clean.Log(clientID), action)
		} else {
			log.Debugf("oauth2: client %s refused (%s)", clean.Log(clientID), action)
			event.AuditWarn([]string{clientIp, "oauth2", "%s", status.Error(err)}, action)
		}

		oauthTokenError(c, http.StatusUnauthorized, oauth.ErrCodeInvalidClient, "invalid client")
		return nil, false
	}

	return client, true
}

// oauthChargeMissingGrant counts a valid refresh token whose grant does not exist towards limiter.Authorize.
// It takes a token only while one is available, so the budget refills once requests stop.
func oauthChargeMissingGrant(clientIp string, err error) {
	if errors.Is(err, oauth.ErrGrantNotFound) {
		limiter.Authorize.Allow(clientIp)
	}
}

// oauthAppTokenUser returns the active user a code or grant belongs to, or responds with an error.
func oauthAppTokenUser(c *gin.Context, userUID, actor, action string) (*entity.User, bool) {
	user := entity.FindUserByUID(userUID)

	if user == nil || user.IsDeleted() || user.IsDisabled() || user.DenyLogIn() {
		event.AuditWarn([]string{ClientIP(c), "oauth2", "%s", "%s", "user %s", status.Denied}, actor, action, clean.Log(userUID))
		oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "user cannot sign in")
		return nil, false
	}

	return user, true
}

// oauthAppTokenResponse stores a new access token for the grant and sends it with the refresh token.
func oauthAppTokenResponse(c *gin.Context, grant *entity.Grant, user *entity.User, refreshToken string, grantType authn.GrantType, actor, action string) {
	clientIp := ClientIP(c)
	sess := oauth.NewAccessToken(grant, user, grantType).SetContext(c)

	if err := sess.Create(); err != nil {
		event.AuditErr([]string{clientIp, "oauth2", "%s", "%s", status.Error(err)}, actor, action)
		oauthTokenError(c, http.StatusInternalServerError, oauth.ErrCodeServerError, "failed to create access token")
		return
	}

	// The grant may have been revoked concurrently, e.g. because its refresh token was reused.
	if found, err := entity.LookupGrant(grant.GrantUID); err != nil {
		event.AuditErr([]string{clientIp, "oauth2", "%s", "%s", status.Error(err)}, actor, action)
		oauthTokenError(c, http.StatusInternalServerError, oauth.ErrCodeServerError, "failed to create access token")
		return
	} else if found == nil {
		_ = sess.Delete()
		event.AuditWarn([]string{clientIp, "oauth2", "%s", "%s", "grant %s revoked", status.Denied}, actor, action, clean.Log(grant.GrantUID))
		oauthTokenError(c, http.StatusBadRequest, oauth.ErrCodeInvalidGrant, "grant has been revoked")
		return
	}

	event.AuditInfo([]string{clientIp, "oauth2", "%s", "user %s", "%s", "session %s", status.Granted}, actor, clean.Log(user.Username()), action, sess.RefID)

	c.Header(header.CacheControl, header.CacheControlNoStore)
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusOK, OAuthAppTokenResponse{
		AccessToken:  sess.AuthToken(),
		TokenType:    header.AuthBearer,
		ExpiresIn:    int64(oauth.AccessTokenTTL.Seconds()),
		RefreshToken: refreshToken,
		Scope:        grant.Scope,
	})
}

// oauthTokenError responds with a standard OAuth error (RFC 6749, section 5.2).
func oauthTokenError(c *gin.Context, statusCode int, errCode, description string) {
	c.Header(header.CacheControl, header.CacheControlNoStore)
	c.AbortWithStatusJSON(statusCode, gin.H{"error": errCode, "error_description": description})
}
