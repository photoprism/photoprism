package api

import (
	"net/http"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// CreateOAuthToken creates a new access token for clients that
// authenticate with valid OAuth2 client credentials.
//
// POST /api/v1/oauth/token
func CreateOAuthToken(router *gin.RouterGroup) {
	router.POST("/oauth/token", func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		// Get client IP address for logs and rate limiting checks.
		clientIp := ClientIP(c)

		if get.Config().Public() {
			// Abort if running in public mode.
			event.AuditErr([]string{clientIp, "client", "create session", "oauth2", authn.ErrDisabledInPublicMode.Error()})
			AbortForbidden(c)
			return
		}

		var err error

		// Client authentication request credentials.
		var f form.ClientCredentials

		// Allow authentication with basic auth and form values.
		if clientId, clientSecret, _ := header.BasicAuth(c); clientId != "" && clientSecret != "" {
			f.ClientID = clientId
			f.ClientSecret = clientSecret
		} else if err = c.ShouldBind(&f); err != nil {
			event.AuditWarn([]string{clientIp, "client", "create session", "oauth2", "%s"}, err)
			AbortBadRequest(c)
			return
		}

		// Check the credentials for completeness and the correct format.
		if err = f.Validate(); err != nil {
			event.AuditWarn([]string{clientIp, "client", "create session", "oauth2", "%s"}, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		}

		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		// Check request rate limit.
		r := limiter.Login.Request(clientIp)

		// Abort if request rate limit is exceeded.
		if r.Reject() || limiter.Auth.Reject(clientIp) {
			limiter.AbortJSON(c)
			return
		}

		// Find the client that has the ID specified in the authentication request.
		client := entity.FindClientByUID(f.ClientID)

		// Abort if the client ID or secret are invalid.
		if client == nil {
			event.AuditWarn([]string{clientIp, "client %s", "create session", "oauth2", authn.ErrInvalidClientID.Error()}, f.ClientID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if !client.AuthEnabled {
			event.AuditWarn([]string{clientIp, "client %s", "create session", "oauth2", authn.ErrAuthenticationDisabled.Error()}, f.ClientID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if method := client.Method(); !method.IsDefault() && method != authn.MethodOAuth2 {
			event.AuditWarn([]string{clientIp, "client %s", "create session", "oauth2", "method %s not supported"}, f.ClientID, clean.LogQuote(method.String()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if client.InvalidSecret(f.ClientSecret) {
			event.AuditWarn([]string{clientIp, "client %s", "create session", "oauth2", authn.ErrInvalidClientSecret.Error()}, f.ClientID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		}

		// Return the reserved request rate limit tokens after successful authentication.
		r.Success()

		// Create new client session.
		sess := client.NewSession(c)

		// Save new client session.
		if sess, err = get.Session().Save(sess); err != nil {
			event.AuditErr([]string{clientIp, "client %s", "create session", "oauth2", "%s"}, f.ClientID, err)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if sess == nil {
			event.AuditErr([]string{clientIp, "client %s", "create session", "oauth2", StatusFailed.String()}, f.ClientID)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrUnexpected)})
			return
		} else {
			event.AuditInfo([]string{clientIp, "client %s", "session %s", "oauth2", "created"}, f.ClientID, sess.RefID)
		}

		// Deletes old client sessions above the configured limit.
		if deleted := client.EnforceAuthTokenLimit(); deleted > 0 {
			event.AuditInfo([]string{clientIp, "client %s", "session %s", "oauth2", "deleted %s"}, f.ClientID, sess.RefID, english.Plural(deleted, "previously created client session", "previously created client sessions"))
		}

		// Response includes access token, token type, and token lifetime.
		data := gin.H{
			"access_token": sess.AuthToken(),
			"token_type":   sess.AuthTokenType(),
			"expires_in":   sess.ExpiresIn(),
		}

		// Return JSON response.
		c.JSON(http.StatusOK, data)
	})
}

// RevokeOAuthToken takes an access token and deletes it. A client may only delete its own tokens.
//
// POST /api/v1/oauth/revoke
func RevokeOAuthToken(router *gin.RouterGroup) {
	router.POST("/oauth/revoke", func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		// Get client IP address for logs and rate limiting checks.
		clientIp := ClientIP(c)

		// Abort if running in public mode.
		if get.Config().Public() {
			event.AuditErr([]string{clientIp, "client", "delete session", "oauth2", authn.ErrDisabledInPublicMode.Error()})
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		}

		var err error

		// Token revocation request data.
		var f form.ClientToken

		authToken := AuthToken(c)

		// Get the auth token to be revoked from the submitted form values or the request header.
		if err = c.ShouldBind(&f); err != nil && authToken == "" {
			event.AuditWarn([]string{clientIp, "client", "delete session", "oauth2", "%s"}, err)
			AbortBadRequest(c)
			return
		} else if f.Empty() {
			f.AuthToken = authToken
			f.TypeHint = form.ClientAccessToken
		}

		// Check the token form values.
		if err = f.Validate(); err != nil {
			event.AuditWarn([]string{clientIp, "client", "delete session", "oauth2", "%s"}, err)
			AbortBadRequest(c)
			return
		}

		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		// Find session based on auth token.
		sess, err := entity.FindSession(rnd.SessionID(f.AuthToken))

		if err != nil {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", "%s"}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String(), err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, i18n.NewResponse(http.StatusUnauthorized, i18n.ErrUnauthorized))
			return
		} else if sess == nil {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", authn.Denied}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String())
			c.AbortWithStatusJSON(http.StatusUnauthorized, i18n.NewResponse(http.StatusUnauthorized, i18n.ErrUnauthorized))
			return
		} else if sess.Abort(c) {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", authn.Denied}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String())
			return
		} else if !sess.IsClient() {
			event.AuditErr([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", authn.Denied}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String())
			c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
			return
		} else {
			event.AuditInfo([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", authn.Granted}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String())
		}

		// Delete session cache and database record.
		if err = sess.Delete(); err != nil {
			// Log error.
			event.AuditErr([]string{clientIp, "client %s", "session %s", "delete session as %s", "oauth2", "%s"}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID), sess.ClientRole().String(), err)

			// Return JSON error.
			c.AbortWithStatusJSON(http.StatusNotFound, i18n.NewResponse(http.StatusNotFound, i18n.ErrNotFound))
			return
		}

		// Log event.
		event.AuditInfo([]string{clientIp, "client %s", "session %s", "oauth2", "deleted"}, clean.Log(sess.ClientInfo()), clean.Log(sess.RefID))

		// Return JSON response for confirmation.
		c.JSON(http.StatusOK, DeleteSessionResponse(sess.ID))
	})
}
