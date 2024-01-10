package api

import (
	"net/http"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// CreateOAuthToken creates a new access token for clients that
// authenticate with valid OAuth2 client credentials.
//
// POST /api/v1/oauth/token
func CreateOAuthToken(router *gin.RouterGroup) {
	router.POST("/oauth/token", func(c *gin.Context) {
		// Get client IP address for logs and rate limiting checks.
		clientIP := ClientIP(c)

		// Abort if running in public mode.
		if get.Config().Public() {
			event.AuditErr([]string{clientIP, "create client session", "disabled in public mode"})
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
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
			event.AuditWarn([]string{clientIP, "create client session", "%s"}, err)
			AbortBadRequest(c)
			return
		}

		// Check the credentials for completeness and the correct format.
		if err = f.Validate(); err != nil {
			event.AuditWarn([]string{clientIP, "create client session", "%s"}, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		}

		// Check limit for failed auth requests (max. 10 per minute).
		if limiter.Login.Reject(clientIP) {
			limiter.AbortJSON(c)
			return
		}

		// Find the client that has the ID specified in the authentication request.
		client := entity.FindClient(f.ClientID)

		// Abort if the client ID or secret are invalid.
		if client == nil {
			event.AuditWarn([]string{clientIP, "client %s", "create session", "invalid client_id"}, f.ClientID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			limiter.Login.Reserve(clientIP)
			return
		} else if !client.AuthEnabled {
			event.AuditWarn([]string{clientIP, "client %s", "create session", "authentication disabled"}, f.ClientID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if method := client.Method(); !method.IsDefault() && method != authn.MethodOAuth2 {
			event.AuditWarn([]string{clientIP, "client %s", "create session", "method %s not supported"}, f.ClientID, clean.LogQuote(method.String()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if client.WrongSecret(f.ClientSecret) {
			event.AuditWarn([]string{clientIP, "client %s", "create session", "invalid client_secret"}, f.ClientID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			limiter.Login.Reserve(clientIP)
			return
		}

		// Create new client session.
		sess := client.NewSession(c)

		// Try to log in and save session if successful.
		if sess, err = get.Session().Save(sess); err != nil {
			event.AuditErr([]string{clientIP, "client %s", "create session", "%s"}, f.ClientID, err)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if sess == nil {
			event.AuditErr([]string{clientIP, "client %s", "create session", StatusFailed.String()}, f.ClientID)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrUnexpected)})
			return
		} else {
			event.AuditInfo([]string{clientIP, "client %s", "session %s", "created"}, f.ClientID, sess.RefID)
		}

		// Deletes old client sessions above the configured limit.
		if deleted := client.EnforceAuthTokenLimit(); deleted > 0 {
			event.AuditInfo([]string{clientIP, "client %s", "%s deleted"}, f.ClientID, english.Plural(deleted, "old session", "old sessions"))
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
		// Get client IP address for logs and rate limiting checks.
		clientIP := ClientIP(c)

		// Abort if running in public mode.
		if get.Config().Public() {
			event.AuditErr([]string{clientIP, "delete client session", "disabled in public mode"})
			Abort(c, http.StatusForbidden, i18n.ErrForbidden)
			return
		}

		var err error

		// Token revocation request data.
		var f form.ClientToken

		authToken := AuthToken(c)

		// Get the auth token to be revoked from the submitted form values or the request header.
		if err = c.ShouldBind(&f); err != nil && authToken == "" {
			event.AuditWarn([]string{clientIP, "delete client session", "%s"}, err)
			AbortBadRequest(c)
			return
		} else if f.Empty() {
			f.AuthToken = authToken
			f.TypeHint = form.ClientAccessToken
		}

		// Check the token form values.
		if err = f.Validate(); err != nil {
			event.AuditWarn([]string{clientIP, "delete client session", "%s"}, err)
			AbortBadRequest(c)
			return
		}

		// Find session based on auth token.
		sess, err := entity.FindSession(rnd.SessionID(f.AuthToken))

		if err != nil {
			event.AuditErr([]string{clientIP, "client %s", "session %s", "delete session as %s", "%s"}, clean.Log(sess.AuthID), clean.Log(sess.RefID), acl.RoleClient.String(), err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, i18n.NewResponse(http.StatusUnauthorized, i18n.ErrUnauthorized))
			return
		} else if sess == nil {
			event.AuditErr([]string{clientIP, "client %s", "session %s", "delete session as %s", "denied"}, clean.Log(sess.AuthID), clean.Log(sess.RefID), acl.RoleClient.String())
			c.AbortWithStatusJSON(http.StatusUnauthorized, i18n.NewResponse(http.StatusUnauthorized, i18n.ErrUnauthorized))
			return
		} else if sess.Abort(c) {
			event.AuditErr([]string{clientIP, "client %s", "session %s", "delete session as %s", "denied"}, clean.Log(sess.AuthID), clean.Log(sess.RefID), acl.RoleClient.String())
			return
		} else if !sess.IsClient() {
			event.AuditErr([]string{clientIP, "client %s", "session %s", "delete session as %s", "denied"}, clean.Log(sess.AuthID), clean.Log(sess.RefID), acl.RoleClient.String())
			c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
			return
		} else {
			event.AuditInfo([]string{clientIP, "client %s", "session %s", "delete session as %s", "granted"}, clean.Log(sess.AuthID), clean.Log(sess.RefID), acl.RoleClient.String())
		}

		// Delete session cache and database record.
		if err = sess.Delete(); err != nil {
			// Log error.
			event.AuditErr([]string{clientIP, "client %s", "session %s", "delete session as %s", "%s"}, clean.Log(sess.AuthID), clean.Log(sess.RefID), acl.RoleClient.String(), err)

			// Return JSON error.
			c.AbortWithStatusJSON(http.StatusNotFound, i18n.NewResponse(http.StatusNotFound, i18n.ErrNotFound))
			return
		}

		// Log event.
		event.AuditInfo([]string{clientIP, "client %s", "session %s", "deleted"}, clean.Log(sess.AuthID), clean.Log(sess.RefID))

		// Return JSON response for confirmation.
		c.JSON(http.StatusOK, DeleteSessionResponse(sess.ID))
	})
}
