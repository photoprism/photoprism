package api

import (
	"errors"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/oauth"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// OAuthConsentScope represents a requested scope and whether the account can use it.
type OAuthConsentScope struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// OAuthConsentClient represents the app that requests access.
type OAuthConsentClient struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Origin string `json:"origin"`
}

// OAuthConsentAccount represents the signed-in account that is asked for consent.
type OAuthConsentAccount struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// OAuthConsentResponse describes a pending authorization request on the consent page.
type OAuthConsentResponse struct {
	Client    OAuthConsentClient  `json:"client"`
	Account   OAuthConsentAccount `json:"account"`
	SiteUrl   string              `json:"siteUrl"`
	Scopes    []OAuthConsentScope `json:"scopes"`
	Available bool                `json:"available"`
	ExpiresAt string              `json:"expiresAt"`
}

// OAuthConsentRedirect contains the URL the browser follows to return to the app.
type OAuthConsentRedirect struct {
	RedirectUri string `json:"redirectUri"`
}

// OAuthConsent returns the pending authorization request shown on the consent page.
//
//	@Summary	returns a pending OAuth authorization request for the consent page
//	@Id			OAuthConsent
//	@Tags		Authentication
//	@Accept		json
//	@Produce	json
//	@Param		request						body		form.OAuthConsent	true	"pending request ID"
//	@Success	200							{object}	OAuthConsentResponse
//	@Failure	400,401,403,404,413,429,500	{object}	i18n.Response
//	@Router		/api/v1/oauth/consent [post]
func OAuthConsent(router *gin.RouterGroup) {
	router.POST("/oauth/consent", func(c *gin.Context) {
		s, user, m, _, ok := oauthConsentRequest(c, "view consent")

		if !ok {
			return
		}

		requested := oauth.RequestScopes(m)
		usable := oauthConsentUsableScopes(s, user, requested)
		scopes := make([]OAuthConsentScope, 0, len(requested))

		for _, name := range requested {
			scopes = append(scopes, OAuthConsentScope{Name: name, Available: slices.Contains(usable, name)})
		}

		event.AuditInfo([]string{ClientIP(c), "session %s", "oauth2", "client %s", "view consent", status.Succeeded}, s.RefID, clean.Log(m.ClientID))

		c.JSON(http.StatusOK, OAuthConsentResponse{
			Client: OAuthConsentClient{
				ID:     m.ClientID,
				Name:   m.ClientName,
				Origin: oauth.ClientOrigin(m.ClientID, m.RedirectURI),
			},
			Account: OAuthConsentAccount{
				Name:        user.Username(),
				DisplayName: user.FullName(),
			},
			SiteUrl:   get.Config().SiteUrl(),
			Scopes:    scopes,
			Available: len(usable) > 0,
			ExpiresAt: m.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	})
}

// OAuthConsentApprove approves a pending authorization request for the approved scopes and returns the
// URL that passes the authorization code to the app.
//
//	@Summary	approves a pending OAuth authorization request
//	@Id			OAuthConsentApprove
//	@Tags		Authentication
//	@Accept		json
//	@Produce	json
//	@Param		request						body		form.OAuthConsent	true	"pending request ID and approved scopes"
//	@Success	200							{object}	OAuthConsentRedirect
//	@Failure	400,401,403,404,413,429,500	{object}	i18n.Response
//	@Router		/api/v1/oauth/consent/approve [post]
func OAuthConsentApprove(router *gin.RouterGroup) {
	router.POST("/oauth/consent/approve", func(c *gin.Context) {
		s, user, m, frm, ok := oauthConsentRequest(c, "approve")

		if !ok {
			return
		}

		clientIp := ClientIP(c)
		usable := oauthConsentUsableScopes(s, user, oauth.RequestScopes(m))

		// Approved scopes must include the MCP scope, have been requested, and be usable with the account.
		if !slices.Contains(frm.Scopes, oauth.DefaultScope) {
			event.AuditWarn([]string{clientIp, "session %s", "oauth2", "client %s", "approve", "scope %s missing", status.Failed}, s.RefID, clean.Log(m.ClientID), oauth.DefaultScope)
			AbortBadRequest(c)
			return
		}

		for _, name := range frm.Scopes {
			if !slices.Contains(usable, name) {
				event.AuditWarn([]string{clientIp, "session %s", "oauth2", "client %s", "approve", "scope %s", status.Denied}, s.RefID, clean.Log(m.ClientID), clean.Log(name))
				AbortForbidden(c)
				return
			}
		}

		code, err := oauth.CompleteRequest(m, user.UserUID, frm.Scopes)

		if errors.Is(err, oauth.ErrRequestNotFound) {
			event.AuditWarn([]string{clientIp, "session %s", "oauth2", "client %s", "approve", status.Error(err)}, s.RefID, clean.Log(m.ClientID))
			AbortEntityNotFound(c)
			return
		} else if err != nil {
			event.AuditErr([]string{clientIp, "session %s", "oauth2", "client %s", "approve", status.Error(err)}, s.RefID, clean.Log(m.ClientID))
			AbortUnexpectedError(c)
			return
		}

		dest, err := oauth.CodeRedirectURL(m.RedirectURI, code, m.State, get.Config().SiteUrl())

		if err != nil {
			event.AuditErr([]string{clientIp, "session %s", "oauth2", "client %s", "approve", status.Error(err)}, s.RefID, clean.Log(m.ClientID))
			AbortUnexpectedError(c)
			return
		}

		event.AuditInfo([]string{clientIp, "session %s", "oauth2", "client %s", "user %s", "approve %s", status.Granted}, s.RefID, clean.Log(m.ClientID), clean.Log(user.Username()), clean.Log(oauth.JoinScopes(frm.Scopes)))
		c.JSON(http.StatusOK, OAuthConsentRedirect{RedirectUri: dest})
	})
}

// OAuthConsentDeny denies a pending authorization request and returns the URL that passes the
// access_denied error to the app.
//
//	@Summary	denies a pending OAuth authorization request
//	@Id			OAuthConsentDeny
//	@Tags		Authentication
//	@Accept		json
//	@Produce	json
//	@Param		request						body		form.OAuthConsent	true	"pending request ID"
//	@Success	200							{object}	OAuthConsentRedirect
//	@Failure	400,401,403,404,413,429,500	{object}	i18n.Response
//	@Router		/api/v1/oauth/consent/deny [post]
func OAuthConsentDeny(router *gin.RouterGroup) {
	router.POST("/oauth/consent/deny", func(c *gin.Context) {
		s, user, m, _, ok := oauthConsentRequest(c, "deny")

		if !ok {
			return
		}

		clientIp := ClientIP(c)

		if err := oauth.DenyRequest(m); errors.Is(err, oauth.ErrRequestNotFound) {
			event.AuditWarn([]string{clientIp, "session %s", "oauth2", "client %s", "deny", status.Error(err)}, s.RefID, clean.Log(m.ClientID))
			AbortEntityNotFound(c)
			return
		} else if err != nil {
			event.AuditErr([]string{clientIp, "session %s", "oauth2", "client %s", "deny", status.Error(err)}, s.RefID, clean.Log(m.ClientID))
			AbortUnexpectedError(c)
			return
		}

		dest, err := oauth.ErrorRedirectURL(m.RedirectURI, oauth.ErrCodeAccessDenied, "access denied", m.State, get.Config().SiteUrl())

		if err != nil {
			event.AuditErr([]string{clientIp, "session %s", "oauth2", "client %s", "deny", status.Error(err)}, s.RefID, clean.Log(m.ClientID))
			AbortUnexpectedError(c)
			return
		}

		event.AuditInfo([]string{clientIp, "session %s", "oauth2", "client %s", "user %s", "deny", status.Denied}, s.RefID, clean.Log(m.ClientID), clean.Log(user.Username()))
		c.JSON(http.StatusOK, OAuthConsentRedirect{RedirectUri: dest})
	})
}

// oauthConsentUsableScopes returns the requested scopes that the account's role, and the scope of the session
// if it has one, allow to use; none if the MCP scope is not usable or the account may not use MCP.
func oauthConsentUsableScopes(s *entity.Session, user *entity.User, requested []string) []string {
	result := []string{}

	if s == nil || user == nil || !user.CanUseMCP() {
		return result
	}

	for _, name := range oauth.UsableScopes(requested, user.AclRole()) {
		if s.NoScope() || s.ValidateScope(acl.Resource(name), acl.Permissions{acl.ActionView}) {
			result = append(result, name)
		}
	}

	if !slices.Contains(result, oauth.DefaultScope) {
		return []string{}
	}

	return result
}

// oauthConsentRequest checks the session and form of a consent request and returns the signed-in user and
// the pending authorization request, or aborts with an error. Only a full web session of an active user may
// decide on a request, not a client or an app password.
func oauthConsentRequest(c *gin.Context, action string) (s *entity.Session, user *entity.User, m *entity.AuthCode, frm form.OAuthConsent, ok bool) {
	// Prevent CDNs from caching this endpoint.
	if header.IsCdn(c.Request) {
		AbortNotFound(c)
		return s, user, m, frm, false
	}

	c.Header(header.CacheControl, header.CacheControlNoStore)

	conf := get.Config()
	clientIp := ClientIP(c)

	switch {
	case conf.Public():
		event.AuditErr([]string{clientIp, "oauth2", "%s", authn.ErrDisabledInPublicMode.Error()}, action)
		AbortForbidden(c)
		return s, user, m, frm, false
	case !conf.OAuthAppsEnabled():
		AbortFeatureDisabled(c)
		return s, user, m, frm, false
	case limiter.Auth.Reject(clientIp):
		limiter.AbortJSON(c)
		return s, user, m, frm, false
	}

	var err error

	if s, err = LookupSession(clientIp, AuthToken(c)); errors.Is(err, authn.ErrRateLimitExceeded) {
		limiter.AbortJSON(c)
		return s, user, m, frm, false
	} else if s == nil || s.VerifyStored() != nil {
		AbortUnauthorized(c)
		return s, user, m, frm, false
	}

	user = s.GetUser()

	if s.IsClient() || !s.IsRegistered() || user.IsDisabled() || user.DenyLogIn() {
		event.AuditWarn([]string{clientIp, "session %s", "oauth2", "%s", "session not eligible", status.Denied}, s.RefID, action)
		AbortForbidden(c)
		return s, user, m, frm, false
	}

	LimitRequestBodyBytes(c, MaxOAuthRequestBytes)

	if err = c.ShouldBindJSON(&frm); err != nil {
		if IsRequestBodyTooLarge(err) {
			AbortRequestTooLarge(c, i18n.ErrBadRequest)
		} else {
			AbortBadRequest(c, err)
		}

		return s, user, m, frm, false
	}

	if m, err = oauth.FindRequest(frm.Request); err != nil {
		limiter.Auth.Reserve(clientIp)
		event.AuditWarn([]string{clientIp, "session %s", "oauth2", "%s", status.Error(err)}, s.RefID, action)
		AbortEntityNotFound(c)
		return s, user, m, frm, false
	}

	// The app client or its callback URL may have been disabled or removed since the request was made.
	if client, clientErr := oauth.FindClient(conf, m.ClientID, oauth.ClientLookup{AllowFetch: oauthMetadataFetchAllowed(clientIp)}); clientErr != nil {
		event.AuditWarn([]string{clientIp, "session %s", "oauth2", "client %s", "%s", status.Error(clientErr)}, s.RefID, clean.Log(m.ClientID), action)
		AbortEntityNotFound(c)
		return s, user, m, frm, false
	} else if !oauth.MatchRedirectURI(client.RedirectURIs, m.RedirectURI) {
		event.AuditWarn([]string{clientIp, "session %s", "oauth2", "client %s", "%s", "redirect uri not registered", status.Denied}, s.RefID, clean.Log(m.ClientID), action)
		AbortEntityNotFound(c)
		return s, user, m, frm, false
	}

	frm.Scopes = slices.Compact(slices.Sorted(slices.Values(frm.Scopes)))

	return s, user, m, frm, true
}
