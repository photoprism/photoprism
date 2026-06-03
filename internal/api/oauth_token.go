package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// OAuthToken creates a new access token for clients using OAuth2 grant types.
//
//	@Summary	create an OAuth2 access token
//	@Id			OAuthToken
//	@Tags		Authentication
//	@Accept		json
//	@Produce	json
//	@Param		request		body		form.OAuthCreateToken	true	"token request (supports client_credentials, password, session, or authorization_code grant)"
//	@Success	200			{object}	gin.H
//	@Failure	400,401,429	{object}	i18n.Response
//	@Router		/api/v1/oauth/token [post]
func OAuthToken(router *gin.RouterGroup) {
	router.POST("/oauth/token", func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		// Get client IP address for logs and rate limiting checks.
		clientIp := ClientIP(c)
		actor := "unknown client"
		action := "create token"

		// Abort if running in public mode.
		if get.Config().Public() {
			event.AuditErr([]string{clientIp, "oauth2", actor, action, authn.ErrDisabledInPublicMode.Error()})
			AbortForbidden(c)
			return
		}

		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		// On Portal builds, a cluster authorization-code redemption is served by
		// the OIDC OP flow (EdDSA id_token). Every other grant — including
		// instance client_credentials and general authorization codes — falls
		// through to the opaque, DB-backed token path below.
		if OAuthClusterToken != nil && OAuthClusterToken(c) {
			return
		}

		// Token create request form.
		var frm form.OAuthCreateToken
		var sess *entity.Session
		var client *entity.Client
		var err error

		// Allow authentication with basic auth and form values.
		if clientId, clientSecret, _ := header.BasicAuth(c); clientId != "" && clientSecret != "" {
			frm.GrantType = authn.GrantClientCredentials
			frm.ClientID = clientId
			frm.ClientSecret = clientSecret
		} else if err = c.ShouldBind(&frm); err != nil {
			event.AuditWarn([]string{clientIp, "oauth2", actor, action, status.Error(err)})
			AbortBadRequest(c, err)
			return
		}

		// Check the credentials for completeness and the correct format.
		if err = frm.Validate(); err != nil {
			event.AuditWarn([]string{clientIp, "oauth2", actor, action, status.Error(err)})
			AbortInvalidCredentials(c)
			return
		}

		// Check request rate limit.
		r := limiter.Login.Request(clientIp)

		// Abort if request rate limit is exceeded.
		if r.Reject() || limiter.Auth.Reject(clientIp) {
			limiter.AbortJSON(c)
			return
		}

		switch {
		case frm.ClientID != "":
			actor = fmt.Sprintf("client %s", clean.Log(frm.ClientID))
		case frm.Username != "":
			actor = fmt.Sprintf("user %s", clean.Log(frm.Username))
		case frm.GrantType == authn.GrantPassword:
			actor = "unknown user"
		}

		// Create a new session (access token) based on the grant type specified in the request.
		switch frm.GrantType {
		case authn.GrantClientCredentials, authn.GrantUndefined:
			// Find client with the specified ID.
			client = entity.FindClientByUID(frm.ClientID)

			// Check if a client has been found, it is enabled, and the credentials are valid.
			if client == nil {
				event.AuditWarn([]string{clientIp, "oauth2", actor, action, authn.ErrInvalidClientID.Error()})
				AbortInvalidCredentials(c)
				return
			} else if !client.AuthEnabled {
				event.AuditWarn([]string{clientIp, "oauth2", actor, action, authn.ErrAuthenticationDisabled.Error()})
				AbortInvalidCredentials(c)
				return
			} else if method := client.Method(); !method.IsDefault() && method != authn.MethodOAuth2 {
				event.AuditWarn([]string{clientIp, "oauth2", actor, action, "method %s", status.Unsupported}, clean.LogQuote(method.String()))
				AbortInvalidCredentials(c)
				return
			} else if client.InvalidSecret(frm.ClientSecret) {
				event.AuditWarn([]string{clientIp, "oauth2", actor, action, authn.ErrInvalidClientSecret.Error()})
				AbortInvalidCredentials(c)
				return
			}

			// Update time of last activity.
			client.UpdateLastActive(true)

			// Cancel failure rate limit reservation.
			r.Success()

			// Create new client session.
			sess = client.NewSession(c, authn.GrantClientCredentials)
		case authn.GrantPassword, authn.GrantSession:
			// Generate an app password for a user account and check the password for confirmation.
			s := Session(clientIp, AuthToken(c))

			if s == nil {
				AbortInvalidCredentials(c)
				return
			} else if s.GetUserName() == "" || s.IsClient() || !s.IsRegistered() {
				event.AuditErr([]string{clientIp, "oauth2", actor, action, authn.ErrInvalidGrantType.Error()})
				AbortInvalidCredentials(c)
				return
			}

			actor = fmt.Sprintf("user %s", clean.Log(s.GetUserName()))

			if s.GetUser().Provider().SupportsPasswordAuthentication() {
				loginForm := form.Login{
					Username: s.GetUserName(),
					Password: frm.Password,
				}

				authUser, authProvider, authMethod, authErr := entity.Auth(loginForm, nil, c)

				switch {
				case authProvider.IsClient():
					event.AuditErr([]string{clientIp, "oauth2", actor, action, status.Denied})
					AbortInvalidCredentials(c)
					return
				case authMethod.Is(authn.Method2FA) && errors.Is(authErr, authn.ErrPasscodeRequired):
					// Ok.
				case authErr != nil:
					event.AuditErr([]string{clientIp, "oauth2", actor, action, status.Error(authErr)})
					AbortInvalidCredentials(c)
					return
				case !authUser.Equal(s.GetUser()):
					event.AuditErr([]string{clientIp, "oauth2", actor, action, authn.ErrUserDoesNotMatch.Error()})
					AbortInvalidCredentials(c)
					return
				}

				frm.GrantType = authn.GrantPassword
			} else {
				frm.GrantType = authn.GrantSession
			}

			sess = entity.NewClientSession(frm.ClientName, frm.ExpiresIn, frm.Scope, frm.GrantType, s.GetUser())

			// Return the reserved request rate limit tokens after successful authentication.
			r.Success()
		case authn.GrantAuthorizationCode:
			// Public clients authenticate with PKCE, not a secret; look up the
			// client only to confirm it exists and OAuth2 is enabled for it.
			codeClient := entity.FindClientByUID(frm.ClientID)
			if codeClient == nil {
				event.AuditWarn([]string{clientIp, "oauth2", actor, action, authn.ErrInvalidClientID.Error()})
				oauthAbortInvalidGrant(c)
				return
			} else if !codeClient.AuthEnabled {
				event.AuditWarn([]string{clientIp, "oauth2", actor, action, authn.ErrAuthenticationDisabled.Error()})
				oauthAbortInvalidGrant(c)
				return
			}

			// Look up the single-use authorization code.
			authCode, findErr := entity.FindOAuthCode(frm.Code)
			if findErr != nil {
				event.AuditErr([]string{clientIp, "oauth2", actor, action, status.Error(findErr)})
				AbortUnexpectedError(c)
				return
			} else if authCode == nil {
				event.AuditWarn([]string{clientIp, "oauth2", actor, action, "authorization code", status.NotFound})
				oauthAbortInvalidGrant(c)
				return
			}

			// Consume the code before validating so it can be redeemed at most
			// once, even on concurrent attempts and even if validation fails.
			if consumed, delErr := authCode.Consume(); delErr != nil {
				event.AuditErr([]string{clientIp, "oauth2", actor, action, status.Error(delErr)})
				AbortUnexpectedError(c)
				return
			} else if !consumed {
				event.AuditWarn([]string{clientIp, "oauth2", actor, action, "authorization code", status.Denied})
				oauthAbortInvalidGrant(c)
				return
			}

			// Validate the code against the request: it must be unexpired and
			// bound to this client, redirect_uri, and PKCE verifier. A mismatch
			// maps to invalid_grant without disclosing which check failed.
			if authCode.IsExpired() ||
				authCode.ClientUID != codeClient.GetUID() ||
				authCode.RedirectURI != frm.RedirectURI ||
				!authn.VerifyPKCE(frm.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
				event.AuditWarn([]string{clientIp, "oauth2", actor, action, "authorization code", status.Denied})
				oauthAbortInvalidGrant(c)
				return
			}

			// Resolve the account the code was issued to.
			user := entity.FindUserByUID(authCode.UserUID)
			if user == nil || user.IsUnknown() || user.IsDisabled() || !user.IsRegistered() {
				event.AuditWarn([]string{clientIp, "oauth2", actor, action, authn.ErrInvalidUser.Error()})
				oauthAbortInvalidGrant(c)
				return
			}

			actor = fmt.Sprintf("user %s", clean.Log(user.Username()))

			// Bound the token lifetime to the client's configured expiry. A
			// public client redeems the code, so it must not be able to request
			// a longer-lived token than the client is configured for; a shorter
			// requested expires_in is honored.
			expiresIn := codeClient.AuthExpires
			if frm.ExpiresIn > 0 && frm.ExpiresIn < expiresIn {
				expiresIn = frm.ExpiresIn
			}

			// Mint a general-purpose user session for the authorized app. It is
			// not linked to the client (no client-session token-limit pruning),
			// matching the password/session grants.
			sess = entity.NewClientSession(codeClient.Name(), expiresIn, authCode.Scope, authn.GrantAuthorizationCode, user)

			// Return the reserved request rate limit tokens after success.
			r.Success()
		default:
			event.AuditErr([]string{clientIp, "oauth2", actor, action, authn.ErrInvalidGrantType.Error()})
			AbortInvalidCredentials(c)
			return
		}

		// Save new session.
		if sess, err = get.Session().Save(sess); err != nil {
			event.AuditErr([]string{clientIp, "oauth2", actor, action, status.Error(err)})
			AbortInvalidCredentials(c)
			return
		} else if sess == nil {
			event.AuditErr([]string{clientIp, "oauth2", actor, action, StatusFailed.String()})
			AbortUnexpectedError(c)
			return
		} else {
			event.AuditInfo([]string{clientIp, "oauth2", actor, action, status.Created})
		}

		// Delete any existing client sessions above the configured limit.
		if client == nil {
			// Skip deletion if not created by a client.
		} else if deleted := client.EnforceAuthTokenLimit(); deleted > 0 {
			event.AuditInfo([]string{clientIp, "oauth2", actor, action, "deleted %s to enforce token limit"}, english.Plural(deleted, "session", "sessions"))
		}

		// Send response with access token, token type, and token lifetime.
		response := gin.H{
			"status":       StatusSuccess,
			"session_id":   sess.ID,
			"access_token": sess.AuthToken(),
			"token_type":   sess.AuthTokenType(),
			"expires_in":   sess.ExpiresIn(),
			"client_name":  sess.GetClientName(),
			"client_role":  sess.GetClientRole(),
			"scope":        sess.Scope(),
		}

		c.JSON(http.StatusOK, response)
	})
}

// oauthAbortInvalidGrant writes the OAuth2 invalid_grant error per RFC 6749
// §5.2 and stops the handler chain. Used by the authorization_code grant for an
// authorization code that is unknown, expired, already used, or mismatched, so
// the client cannot distinguish those cases.
func oauthAbortInvalidGrant(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"error":             "invalid_grant",
		"error_description": "authorization code is invalid, expired, or already used",
	})
}
