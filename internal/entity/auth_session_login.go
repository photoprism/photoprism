package entity

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Auth checks if the credentials are valid and returns the user and authentication provider.
var Auth = func(f form.Login, m *Session, c *gin.Context) (user *User, provider authn.ProviderType, method authn.MethodType, err error) {
	// Get username from login form.
	nameName := f.Username()

	// Find registered user account.
	user = FindUserByName(nameName)

	// Try local authentication.
	provider, method, err = AuthLocal(user, f, m, c)

	if err != nil {
		return user, provider, method, err
	}

	// Update login timestamp.
	user.UpdateLoginTime()

	return user, provider, method, err
}

// AuthSession returns the client session that belongs to the auth token provided, or returns nil if it was not found.
func AuthSession(f form.Login, c *gin.Context) (sess *Session, user *User, err error) {
	if f.Password == "" {
		// Abort authentication if no token was provided.
		return nil, nil, authn.ErrPasscodeRequired
	} else if !rnd.IsAppPassword(f.Password, true) {
		// Abort authentication if token doesn't match expected format.
		return nil, nil, authn.ErrInvalidPassword
	}

	// Get session ID for the auth token provided.
	sid := rnd.SessionID(f.Password)

	// Find the session based on the hashed token used as session ID and return it.
	sess, err = FindSession(sid)

	// Log error and return nil if no matching session was found.
	if sess == nil || err != nil {
		return nil, nil, authn.ErrInvalidPassword
	}

	// Update the client IP and the user agent from
	// the request context if they have changed.
	sess.UpdateContext(c)

	// Returns session and user if all checks have passed.
	return sess, sess.User(), nil
}

// AuthLocal authenticates against the local user database with the specified username and password.
func AuthLocal(user *User, f form.Login, m *Session, c *gin.Context) (provider authn.ProviderType, method authn.MethodType, err error) {
	// Set defaults.
	provider = authn.ProviderNone
	method = authn.MethodUndefined

	// Get client IP from request context.
	clientIp := header.ClientIP(c)

	// Get username from login form.
	userName := f.Username()

	// Check if user account exists.
	if user == nil {
		message := authn.ErrAccountNotFound.Error()
		limiter.Login.Reserve(clientIp)

		if m != nil {
			event.AuditWarn([]string{clientIp, "session %s", "login as %s", message}, m.RefID, clean.LogQuote(userName))
			event.LoginError(clientIp, "api", userName, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
		}

		return provider, method, i18n.Error(i18n.ErrInvalidCredentials)
	}

	// Login allowed?
	if !user.Provider().IsDefault() && !user.Provider().IsLocal() {
		message := fmt.Sprintf("%s authentication disabled", authn.ProviderLocal.String())

		if m != nil {
			event.AuditWarn([]string{clientIp, "session %s", "login as %s", message}, m.RefID, clean.LogQuote(userName))
			event.LoginError(clientIp, "api", userName, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
		}

		return provider, method, i18n.Error(i18n.ErrInvalidCredentials)
	} else if !user.CanLogIn() {
		message := authn.ErrAccountDisabled.Error()

		if m != nil {
			event.AuditWarn([]string{clientIp, "session %s", "login as %s", message}, m.RefID, clean.LogQuote(userName))
			event.LoginError(clientIp, "api", userName, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
		}

		return provider, method, i18n.Error(i18n.ErrInvalidCredentials)
	}

	// Authentication with personal access token if a valid secret has been provided as password.
	if authSess, authUser, authErr := AuthSession(f, c); authSess != nil && authUser != nil && authErr == nil {
		if !authUser.IsRegistered() || authUser.UserUID != user.UserUID {
			message := authn.ErrInvalidUsername.Error()
			limiter.Login.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "session %s", "login as %s with app password", message}, m.RefID, clean.LogQuote(userName))
			event.LoginError(clientIp, "api", userName, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
			return provider, method, i18n.Error(i18n.ErrInvalidCredentials)
		} else if insufficientScope := authSess.InsufficientScope(acl.ResourceSessions, acl.Permissions{acl.ActionCreate}); insufficientScope || !authSess.IsClient() {
			var message string
			if insufficientScope {
				message = authn.ErrInsufficientScope.Error()
			} else {
				message = authn.ErrUnauthorized.Error()
			}
			limiter.Login.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "session %s", "login as %s with app password", message}, m.RefID, clean.LogQuote(userName))
			event.LoginError(clientIp, "api", userName, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
			return provider, method, i18n.Error(i18n.ErrInvalidCredentials)
		} else {
			provider = authn.ProviderApplication
			method = authn.MethodSession
			m.ClientUID = authSess.ClientUID
			m.ClientName = authSess.ClientName
			m.SetScope(authSess.Scope())
			m.SetMethod(authn.MethodSession)
			event.AuditInfo([]string{clientIp, "session %s", "login as %s with app password", "succeeded"}, m.RefID, clean.LogQuote(userName))
			event.LoginInfo(clientIp, "api", userName, m.UserAgent)
			return provider, method, authErr
		}
	}

	// Otherwise, check account password.
	if user.WrongPassword(f.Password) {
		message := authn.ErrInvalidPassword.Error()
		limiter.Login.Reserve(clientIp)

		if m != nil {
			event.AuditErr([]string{clientIp, "session %s", "login as %s", message}, m.RefID, clean.LogQuote(userName))
			event.LoginError(clientIp, "api", userName, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
		}

		return provider, method, i18n.Error(i18n.ErrInvalidCredentials)
	}

	provider = authn.ProviderLocal

	// Perform two-factor authentication check, if required.
	if method = user.Method(); method.Is(authn.Method2FA) {
		if valid, _, passcodeErr := user.VerifyPasscode(f.Passcode); passcodeErr != nil {
			limiter.Login.Reserve(clientIp)
			return provider, method, passcodeErr
		} else if !valid {
			limiter.Login.ReserveN(clientIp, 3)
			return provider, method, authn.ErrInvalidPasscode
		}
	} else if method == authn.MethodUndefined {
		method = authn.MethodDefault
	}

	if m != nil {
		event.AuditInfo([]string{clientIp, "session %s", "login as %s", "succeeded"}, m.RefID, clean.LogQuote(userName))
		event.LoginInfo(clientIp, "api", userName, m.UserAgent)
	}

	return provider, method, nil
}

// LogIn performs authentication checks against the specified login form.
func (m *Session) LogIn(f form.Login, c *gin.Context) (err error) {
	if c != nil {
		m.SetContext(c)
	}

	var user *User
	var provider authn.ProviderType
	var method authn.MethodType

	// Log in with username and password?
	if f.HasCredentials() {
		if m.IsRegistered() {
			m.Regenerate()
		}

		user, provider, method, err = Auth(f, m, c)

		m.SetProvider(provider)
		m.SetMethod(method)

		if err != nil {
			return err
		}

		m.SetUser(user)
	}

	// Try to redeem link share token, if provided.
	if f.HasShareToken() {
		user = m.User()

		// Redeem token.
		if user.IsRegistered() {
			if shares := user.RedeemToken(f.ShareToken); shares == 0 {
				message := authn.ErrInvalidShareToken.Error()
				limiter.Login.Reserve(m.IP())
				event.AuditWarn([]string{m.IP(), "session %s", message}, m.RefID)
				m.Status = http.StatusNotFound
				return i18n.Error(i18n.ErrInvalidLink)
			} else {
				event.AuditInfo([]string{m.IP(), "session %s", "token redeemed for %d shares"}, m.RefID, user.RedeemToken(f.ShareToken))
			}
		} else if data := m.Data(); data == nil {
			m.Status = http.StatusInternalServerError
			return i18n.Error(i18n.ErrUnexpected)
		} else if shares := data.RedeemToken(f.ShareToken); shares == 0 {
			message := authn.ErrInvalidShareToken.Error()
			limiter.Login.Reserve(m.IP())
			event.AuditWarn([]string{m.IP(), "session %s", message}, m.RefID)
			event.LoginError(m.IP(), "api", "", m.UserAgent, message)
			m.Status = http.StatusNotFound
			return i18n.Error(i18n.ErrInvalidLink)
		} else {
			m.SetData(data)
			m.SetProvider(authn.ProviderLink)
			event.AuditInfo([]string{m.IP(), "session %s", "token redeemed for %d shares"}, m.RefID, shares, data)
		}

		// Upgrade the session user role to visitor if a valid share token has been provided.
		if user.IsUnknown() {
			user = &Visitor
			event.AuditDebug([]string{m.IP(), "session %s", "role upgraded to %s"}, m.RefID, user.AclRole().String())
			expires := UTC().Add(time.Hour * 24)
			m.Expires(expires)
			event.AuditDebug([]string{m.IP(), "session %s", "expires at %s"}, m.RefID, txt.DateTime(&expires))
		}

		m.SetUser(user)
	}

	// Unregistered visitors must use a valid share link to obtain a session.
	if m.User().NotRegistered() && m.Data().NoShares() {
		m.Status = http.StatusUnauthorized
		return i18n.Error(i18n.ErrInvalidCredentials)
	}

	m.Status = http.StatusOK

	return nil
}
