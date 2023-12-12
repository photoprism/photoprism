package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/pkg/authn"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/server/limiter"
)

// CreateOauthToken creates a new access token and returns it as JSON
// if the client's credentials have been successfully validated.
//
// POST /api/v1/oauth/token
func CreateOauthToken(router *gin.RouterGroup) {
	router.POST("/oauth/token", func(c *gin.Context) {
		// client_id, client_secret
		var err error
		var f form.ClientCredentials

		// Get client IP address for logs and rate limiting checks.
		clientIP := ClientIP(c)

		// Allow authentication with basic auth and form values.
		if clientId, clientSecret, _ := BasicAuth(c); clientId != "" && clientSecret != "" {
			f.ClientID = clientId
			f.ClientSecret = clientSecret
		} else if err = c.Bind(&f); err != nil {
			event.AuditWarn([]string{clientIP, "oauth", "%s"}, err)
			AbortBadRequest(c)
			return
		}

		// Check the credentials for completeness and the correct format.
		if err = f.Validate(); err != nil {
			event.AuditWarn([]string{clientIP, "oauth", "%s"}, err)
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
			event.AuditWarn([]string{clientIP, "client %s", "create access token", "invalid client id"}, f.ClientID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			limiter.Login.Reserve(clientIP)
			return
		} else if !client.AuthEnabled {
			event.AuditWarn([]string{clientIP, "client %s", "create access token", "authentication disabled"}, f.ClientID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if client.AuthMethod != authn.MethodOAuth2.String() {
			event.AuditWarn([]string{clientIP, "client %s", "create access token", "%s authentication not supported"}, f.ClientID, client.AuthMethod)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if client.WrongSecret(f.ClientSecret) {
			event.AuditWarn([]string{clientIP, "client %s", "create access token", "invalid client secret"}, f.ClientID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			limiter.Login.Reserve(clientIP)
			return
		}

		// Create new client session.
		sess := client.NewSession(c)

		// TODO: Enforce limit for maximum number of access tokens.

		// Try to log in and save session if successful.
		if sess, err = get.Session().Save(sess); err != nil {
			event.AuditErr([]string{clientIP, "client %s", "create access token", "%s"}, f.ClientID, err)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if sess == nil {
			event.AuditErr([]string{clientIP, "client %s", "create access token", "failed unexpectedly"}, f.ClientID)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrUnexpected)})
			return
		} else {
			event.AuditInfo([]string{clientIP, "client %s", "session %s", "access token created"}, f.ClientID, sess.RefID)
		}

		// Return access token.
		data := gin.H{
			"access_token": sess.ID,
			"token_type":   "Bearer",
			"expires_in":   sess.ExpiresIn(),
		}

		c.JSON(http.StatusOK, data)
	})
}
