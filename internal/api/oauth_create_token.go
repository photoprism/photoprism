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

		// Abort if running in public mode.
		if get.Config().Public() {
			event.AuditErr([]string{clientIp, "client", "create session", "oauth2", authn.ErrDisabledInPublicMode.Error()})
			AbortForbidden(c)
			return
		}

		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		// Token create request form.
		var f form.OAuthCreateToken
		var sess *entity.Session
		var client *entity.Client
		var err error

		// Allow authentication with basic auth and form values.
		if clientId, clientSecret, _ := header.BasicAuth(c); clientId != "" && clientSecret != "" {
			f.GrantType = authn.GrantClientCredentials
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

		// Check request rate limit.
		r := limiter.Login.Request(clientIp)

		// Abort if request rate limit is exceeded.
		if r.Reject() || limiter.Auth.Reject(clientIp) {
			limiter.AbortJSON(c)
			return
		}

		// Create a new session (access token) based on the grant type specified in the request.
		switch f.GrantType {
		case authn.GrantClientCredentials, authn.GrantUndefined:
			// Find client with the specified ID.
			client = entity.FindClientByUID(f.ClientID)

			// Check if a client has been found, it is enabled, and the credentials are valid.
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

			// Cancel failure rate limit reservation.
			r.Success()

			// Create new client session.
			sess = client.NewSession(c, authn.GrantClientCredentials)
		case authn.GrantPassword:
			// Generate an app password for a user account and accept the password for confirmation.
			event.AuditWarn([]string{clientIp, "client %s", "create session", "oauth2", "password grant type is not implemented yet"}, f.ClientID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		default:
			event.AuditErr([]string{clientIp, "client %s", "create session", "oauth2", authn.ErrInvalidGrantType.Error()}, f.ClientID)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		}

		// Save new session.
		if sess, err = get.Session().Save(sess); err != nil {
			event.AuditErr([]string{clientIp, "client %s", "create session", "oauth2", err.Error()}, f.ClientID)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			return
		} else if sess == nil {
			event.AuditErr([]string{clientIp, "client %s", "create session", "oauth2", StatusFailed.String()}, f.ClientID)
			c.AbortWithStatusJSON(sess.HttpStatus(), gin.H{"error": i18n.Msg(i18n.ErrUnexpected)})
			return
		} else {
			event.AuditInfo([]string{clientIp, "client %s", "session %s", "oauth2", "created"}, f.ClientID, sess.RefID)
		}

		// Delete any existing client sessions above the configured limit.
		if client == nil {
			// Skip deletion if not created by a client.
		} else if deleted := client.EnforceAuthTokenLimit(); deleted > 0 {
			event.AuditInfo([]string{clientIp, "client %s", "session %s", "oauth2", "deleted %s"}, f.ClientID, sess.RefID, english.Plural(deleted, "previously created client session", "previously created client sessions"))
		}

		// Send response with access token, token type, and token lifetime.
		response := gin.H{
			"access_token": sess.AuthToken(),
			"token_type":   sess.AuthTokenType(),
			"expires_in":   sess.ExpiresIn(),
		}

		c.JSON(http.StatusOK, response)
	})
}
