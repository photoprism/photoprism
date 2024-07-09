package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/oidc"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/internal/thumb/avatar"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
	"github.com/photoprism/photoprism/pkg/txt"
)

// OIDCRedirect creates a new API access token when a user has been successfully authenticated via OIDC,
// and then redirects the browser back to the app.
//
// GET /api/v1/oidc/redirect
func OIDCRedirect(router *gin.RouterGroup) {
	router.GET("/oidc/redirect", func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		// Disable caching of responses.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		// Get client IP address for logs and rate limiting checks.
		clientIp := ClientIP(c)
		userAgent := UserAgent(c)
		userName := "unknown user"

		// Get global config.
		conf := get.Config()

		// Abort in public mode and if OIDC is disabled.
		if get.Config().Public() {
			event.AuditErr([]string{clientIp, "create session", "oidc", authn.ErrDisabledInPublicMode.Error()})
			c.Redirect(http.StatusTemporaryRedirect, conf.LoginUri())
			return
		} else if !conf.OIDCEnabled() {
			event.AuditErr([]string{clientIp, "create session", "oidc", authn.ErrAuthenticationDisabled.Error()})
			c.Redirect(http.StatusTemporaryRedirect, conf.LoginUri())
			return
		}

		// Check request rate limit.
		var r *limiter.Request
		r = limiter.Login.Request(clientIp)

		// Abort if failure rate limit is exceeded.
		if r.Reject() || limiter.Auth.Reject(clientIp) {
			c.HTML(http.StatusTooManyRequests, "auth.gohtml", CreateSessionError(http.StatusTooManyRequests, i18n.Error(i18n.ErrTooManyRequests)))
			return
		}

		// Check if the required request parameters are present.
		if c.Query("state") == "" || c.Query("code") == "" {
			event.AuditErr([]string{clientIp, "create session", "oidc", authn.ErrAuthCodeRequired.Error()})
			c.Redirect(http.StatusTemporaryRedirect, conf.LoginUri())
			return
		}

		// Get OIDC provider.
		provider := get.OIDC()

		if provider == nil {
			event.AuditErr([]string{clientIp, "create session", "oidc", authn.ErrInvalidProviderConfiguration.Error()})
			c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
			return
		}

		// Check the auth request and, if successful, get user information and tokens.
		userInfo, tokens, claimErr := provider.CodeExchangeUserInfo(c)

		if claimErr != nil {
			event.AuditErr([]string{clientIp, "create session", "oidc", claimErr.Error()})
			return
		}

		// Step 1: Create user account if it does not exist yet.
		var user *entity.User
		var err error

		userEmail := clean.Email(userInfo.Email)

		// Optionally check if the email domain matches.
		if domain := conf.OIDCDomain(); domain == "" {
			// Do nothing.
		} else if _, emailDomain, _ := strings.Cut(userEmail, "@"); emailDomain == "" || !userInfo.EmailVerified {
			event.AuditErr([]string{clientIp, "create session", "oidc", authn.ErrVerifiedEmailRequired.Error()})
			event.LoginError(clientIp, "oidc", userEmail, userAgent, authn.ErrVerifiedEmailRequired.Error())
			c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrForbidden)))
			return
		} else if !strings.HasSuffix("."+emailDomain, "."+domain) {
			message := fmt.Sprintf("domain must match '%s'", domain)
			event.AuditErr([]string{clientIp, "create session", "oidc", userEmail, message})
			event.LoginError(clientIp, "oidc", userEmail, userAgent, message)
			c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrForbidden)))
			return
		}

		// Find existing user record and update it, if necessary.
		if oidcUser := entity.OidcUser(userInfo, oidc.Username(userInfo, conf.OIDCUsername())); authn.ProviderOIDC.NotEqual(oidcUser.AuthProvider) {
			event.AuditErr([]string{clientIp, "create session", "oidc", authn.ErrAuthProviderIsNotOIDC.Error()})
			event.LoginError(clientIp, "oidc", oidcUser.UserName, userAgent, authn.ErrAuthProviderIsNotOIDC.Error())
			c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
			return
		} else if oidcUser.UserName == "" {
			event.AuditErr([]string{clientIp, "create session", "oidc", authn.ErrUsernameRequiredToRegister.Error()})
			event.LoginError(clientIp, "oidc", oidcUser.UserName, userAgent, authn.ErrUsernameRequiredToRegister.Error())
			c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
			return
		} else if user = entity.FindUser(oidcUser); user != nil {
			// Ensure user has a username.
			if user.Username() == "" {
				event.AuditErr([]string{clientIp, "create session", "oidc", oidcUser.UserName, authn.ErrUsernameRequired.Error()})
				event.LoginError(clientIp, "oidc", oidcUser.UserName, userAgent, authn.ErrUsernameRequired.Error())
				c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
				return
			}

			userName = user.Username()
			event.AuditInfo([]string{clientIp, "create session", "oidc", "found user", userName})

			// Check if the account is enabled and the OIDC Subject ID matches.
			if !user.CanLogIn() {
				event.AuditErr([]string{clientIp, "create session", "oidc", userName, authn.ErrAccountDisabled.Error()})
				event.LoginError(clientIp, "oidc", userName, userAgent, authn.ErrAccountDisabled.Error())
				c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
				return
			} else if authn.ProviderOIDC.NotEqual(user.AuthProvider) {
				event.AuditErr([]string{clientIp, "create session", "oidc", userName, authn.ErrAuthProviderIsNotOIDC.Error()})
				event.LoginError(clientIp, "oidc", userName, userAgent, authn.ErrAuthProviderIsNotOIDC.Error())
				c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
				return
			} else if user.AuthID == "" || oidcUser.AuthID == "" || user.AuthID != oidcUser.AuthID {
				event.AuditErr([]string{clientIp, "create session", "oidc", userName, authn.ErrInvalidAuthID.Error()})
				event.LoginError(clientIp, "oidc", userName, userAgent, authn.ErrInvalidAuthID.Error())
				c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
				return
			}

			// Update user profile information.
			details := user.Details()

			// Update user display name.
			if entity.SrcPriority[details.NameSrc] <= entity.SrcPriority[entity.SrcOIDC] {
				user.SetDisplayName(userInfo.Name, entity.SrcOIDC)
				user.SetGivenName(userInfo.GivenName)
				user.SetFamilyName(userInfo.FamilyName)
				details.UserGender = clean.Name(string(userInfo.Gender))
			}

			// Update nickname.
			if name := clean.Name(userInfo.Nickname); name != "" {
				details.NickName = clean.Name(userInfo.Nickname)
			}

			// Update profile URL.
			if u := clean.Uri(userInfo.Profile); u != "" {
				details.ProfileURL = u
			}

			// Update website URL.
			if u := clean.Uri(userInfo.Website); u != "" {
				details.SiteURL = u
			}

			// Update UI locale.
			user.Settings().UILanguage = clean.Locale(userInfo.Locale.String(), user.Settings().UILanguage)

			// Update UI timezone.
			if tz := userInfo.Zoneinfo; tz != "" && tz != time.UTC.String() {
				user.Settings().UITimeZone = tz
			}

			// Update user location, if available.
			if addr := userInfo.GetAddress(); addr != nil {
				user.Details().UserLocation = clean.Name(addr.Locality)
				user.Details().UserCountry = clean.TypeLowerUnderscore(addr.Country)
			}

			// Update birthday, if available.
			if birthDate := txt.ParseTime(userInfo.Birthdate, userInfo.Zoneinfo); !birthDate.IsZero() {
				user.BornAt = &birthDate
				user.Details().BirthDay = birthDate.Day()
				user.Details().BirthMonth = int(birthDate.Month())
				user.Details().BirthYear = birthDate.Year()
			}

			// Update email, if verified.
			if userInfo.EmailVerified {
				user.UserEmail = clean.Email(userInfo.Email)
				user.VerifiedAt = entity.TimeStamp()
			}

			// Update existing user account.
			if err = user.Save(); err != nil {
				event.AuditErr([]string{clientIp, "create session", "oidc", userName, authn.ErrAccountUpdateFailed.Error(), err.Error()})
				event.LoginError(clientIp, "oidc", userName, userAgent, authn.ErrAccountUpdateFailed.Error()+" ("+err.Error()+")")
				c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
				return
			}

			// Set user avatar image?
			if avatarUrl := userInfo.Picture; avatarUrl == "" || user.HasAvatar() {
				// Do nothing.
			} else if err = avatar.SetUserImageURL(user, avatarUrl, entity.SrcOIDC, conf.ThumbCachePath()); err != nil {
				event.AuditWarn([]string{clientIp, "create session", "oidc", userName, "failed to set avatar image", err.Error()})
			}
		} else if conf.OIDCRegister() {
			// Create new user record.
			user = &oidcUser

			userName = oidcUser.Username()

			// Resolve potential naming conflict by adding a random number to the username.
			if found := entity.FindUserByName(userName); found != nil {
				userName = userName + rnd.Base10(6)
			}

			event.AuditInfo([]string{clientIp, "create session", "oidc", "create user", userName})

			user.UserName = userName

			// Set user profile information.
			user.SetDisplayName(userInfo.Name, entity.SrcOIDC)
			user.SetGivenName(userInfo.GivenName)
			user.SetFamilyName(userInfo.FamilyName)
			user.Details().UserGender = clean.Name(string(userInfo.Gender))
			user.Details().NickName = clean.Name(userInfo.Nickname)

			// Set user profile URL.
			user.Details().ProfileURL = clean.Uri(userInfo.Profile)

			// Set user site URL.
			user.Details().SiteURL = clean.Uri(userInfo.Website)

			// Set UI locale.
			user.Settings().UILanguage = clean.Locale(userInfo.Locale.String(), "")

			// Set UI timezone.
			user.Settings().UITimeZone = userInfo.Zoneinfo

			// Set user location, if available.
			if addr := userInfo.GetAddress(); addr != nil {
				user.Details().UserLocation = clean.Name(addr.Locality)
				user.Details().UserCountry = clean.TypeLowerUnderscore(addr.Country)
			}

			// Set birthday, if available.
			if birthDate := txt.ParseTime(userInfo.Birthdate, userInfo.Zoneinfo); !birthDate.IsZero() {
				user.BornAt = &birthDate
				user.Details().BirthDay = birthDate.Day()
				user.Details().BirthMonth = int(birthDate.Month())
				user.Details().BirthYear = birthDate.Year()
			}

			// Set email, if verified.
			if userInfo.EmailVerified {
				user.UserEmail = clean.Email(userInfo.Email)
				user.VerifiedAt = entity.TimeStamp()
			}

			// Set user role and permissions.
			user.SetRole(conf.OIDCRole().String())
			user.CanLogin = true
			user.WebDAV = conf.OIDCWebDAV()

			// Create new user account.
			if err = user.Create(); err != nil {
				event.AuditErr([]string{clientIp, "create session", "oidc", userName, authn.ErrAccountCreateFailed.Error(), err.Error()})
				event.LoginError(clientIp, "oidc", userName, userAgent, authn.ErrAccountCreateFailed.Error()+": "+err.Error())
				c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
				return
			}

			// Set user avatar image.
			if avatarUrl := userInfo.Picture; avatarUrl == "" {
				event.AuditDebug([]string{clientIp, "create session", "oidc", userName, "no avatar image provided"})
			} else if err = avatar.SetUserImageURL(user, avatarUrl, entity.SrcOIDC, conf.ThumbCachePath()); err != nil {
				event.AuditWarn([]string{clientIp, "create session", "oidc", userName, "failed to set avatar image", err.Error()})
			}
		} else {
			event.AuditErr([]string{clientIp, "create session", "oidc", userName, authn.ErrRegistrationDisabled.Error()})
			event.LoginError(clientIp, "oidc", userName, userAgent, authn.ErrRegistrationDisabled.Error())
			c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
			return
		}

		// Check if login is allowed.
		if !user.CanLogIn() {
			event.AuditErr([]string{clientIp, "create session", "oidc", userName, authn.ErrAccountDisabled.Error()})
			event.LoginError(clientIp, "oidc", userName, userAgent, authn.ErrAccountDisabled.Error())
			c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
			return
		}

		// Update Subject ID (auth_id).
		user.SetAuthID(userInfo.Subject)

		// Step 2: Create user session.
		sess := get.Session().New(c)
		sess.SetProvider(authn.ProviderOIDC)
		sess.SetMethod(authn.MethodDefault)
		sess.SetAuthID(user.AuthID)
		sess.SetUser(user)
		sess.SetGrantType(authn.GrantAuthorizationCode)
		sess.IdToken = tokens.IDToken

		// Set session expiration and timeout.
		sess.SetExpiresIn(unix.Day)
		sess.SetTimeout(-1)

		// Save session after successful authentication.
		if sess, err = get.Session().Save(sess); err != nil {
			event.AuditErr([]string{clientIp, "create session", "oidc", userName, "%s"}, err)
			c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrInvalidCredentials)))
			return
		} else if sess == nil {
			event.AuditErr([]string{clientIp, "create session", "oidc", userName, authn.Failed})
			c.HTML(http.StatusUnauthorized, "auth.gohtml", CreateSessionError(http.StatusUnauthorized, i18n.Error(i18n.ErrUnexpected)))
			return
		}

		// Return the reserved request rate limit token after successful authentication.
		r.Success()

		// Response includes user data, session data, and client config values.
		response := CreateSessionResponse(sess.AuthToken(), sess, conf.ClientSession(sess))

		// Log session created event.
		event.AuditInfo([]string{clientIp, "session %s", "oidc", userName, authn.Created}, sess.RefID)

		// Log session expiration time.
		if expires := sess.ExpiresAt(); !expires.IsZero() {
			event.AuditDebug([]string{clientIp, "session %s", "oidc", userName, "expires at %s"}, sess.RefID, txt.DateTime(&expires))
		}

		// Log successful login.
		event.LoginInfo(clientIp, "oidc", userName, userAgent)

		// Update login timestamp.
		user.UpdateLoginTime()

		// Step 3: Render HTML template to set the access token in localStorage.
		c.HTML(http.StatusOK, "auth.gohtml", response)
	})
}
