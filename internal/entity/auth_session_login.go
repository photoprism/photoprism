package entity

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Auth checks if the credentials are valid and returns the user and authentication provider.
var Auth = func(f form.Login, m *Session, c *gin.Context) (user *User, provider authn.ProviderType, err error) {
	// Get username from login form.
	nameName := f.Username()

	// Find registered user account.
	user = FindUserByName(nameName)

	// Try local authentication.
	provider, err = AuthLocal(user, f, m, c)

	if err != nil {
		return user, authn.ProviderNone, err
	}

	// Update login timestamp.
	user.UpdateLoginTime()

	return user, provider, err
}

// AuthSession returns the client session that belongs to the auth token provided, or returns nil if it was not found.
func AuthSession(f form.Login, c *gin.Context) (sess *Session, user *User, err error) {
	if f.Password == "" {
		// Abort authentication if no token was provided.
		return nil, nil, fmt.Errorf("no auth secret provided")
	} else if !rnd.IsAuthSecret(f.Password, true) {
		// Abort authentication if token doesn't match expected format.
		return nil, nil, fmt.Errorf("auth secret does not match expected format")
	}

	// Get session ID for the auth token provided.
	sid := rnd.SessionID(f.Password)

	// Find the session based on the hashed token used as session ID and return it.
	sess, err = FindSession(sid)

	// Log error and return nil if no matching session was found.
	if sess == nil || err != nil {
		return nil, nil, fmt.Errorf("invalid auth secret")
	}

	// Update the client IP and the user agent from
	// the request context if they have changed.
	sess.UpdateContext(c)

	// Returns session and user if all checks have passed.
	return sess, sess.User(), nil
}

// AuthLocal authenticates against the local user database with the specified username and password.
func AuthLocal(user *User, f form.Login, m *Session, c *gin.Context) (authn.ProviderType, error) {
	// Get client IP from request context.
	clientIp := header.ClientIP(c)

	// Get username from login form.
	userName := f.Username()

	// Check if a session has been created.
	if m == nil {
		event.AuditErr([]string{clientIp, "login as %s", "invalid session"}, clean.LogQuote(userName))
		return authn.ProviderNone, i18n.Error(i18n.ErrInvalidCredentials)
	}

	// Check if user account exists.
	if user == nil {
		message := "account not found"
		limiter.Login.Reserve(clientIp)
		event.AuditWarn([]string{clientIp, "session %s", "login as %s", message}, m.RefID, clean.LogQuote(userName))
		event.LoginError(clientIp, "api", userName, m.UserAgent, message)
		m.Status = http.StatusUnauthorized
		return authn.ProviderNone, i18n.Error(i18n.ErrInvalidCredentials)
	}

	// Login allowed?
	if !user.Provider().IsDefault() && !user.Provider().IsLocal() {
		message := fmt.Sprintf("%s authentication disabled", authn.ProviderLocal.String())
		event.AuditWarn([]string{clientIp, "session %s", "login as %s", message}, m.RefID, clean.LogQuote(userName))
		event.LoginError(clientIp, "api", userName, m.UserAgent, message)
		m.Status = http.StatusUnauthorized
		return authn.ProviderNone, i18n.Error(i18n.ErrInvalidCredentials)
	} else if !user.CanLogIn() {
		message := "account disabled"
		event.AuditWarn([]string{clientIp, "session %s", "login as %s", message}, m.RefID, clean.LogQuote(userName))
		event.LoginError(clientIp, "api", userName, m.UserAgent, message)
		m.Status = http.StatusUnauthorized
		return authn.ProviderNone, i18n.Error(i18n.ErrInvalidCredentials)
	}

	// Authentication with personal access token if a valid secret has been provided as password.
	if authSess, authUser, err := AuthSession(f, c); err == nil {
		if !authUser.IsRegistered() || authUser.UserUID != user.UserUID {
			message := "incorrect user"
			limiter.Login.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "session %s", "login as %s with auth secret", message}, m.RefID, clean.LogQuote(userName))
			event.LoginError(clientIp, "api", userName, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
			return authn.ProviderNone, i18n.Error(i18n.ErrInvalidCredentials)
		} else if !authSess.IsClient() || !authSess.HasScope(acl.ResourceSessions.String()) {
			message := "unauthorized"
			limiter.Login.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "session %s", "login as %s with auth secret", message}, m.RefID, clean.LogQuote(userName))
			event.LoginError(clientIp, "api", userName, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
			return authn.ProviderNone, i18n.Error(i18n.ErrInvalidCredentials)
		} else {
			m.ClientUID = authSess.ClientUID
			m.ClientName = authSess.ClientName
			m.SetScope(authSess.Scope())
			m.SetMethod(authn.MethodSession)
			event.AuditInfo([]string{clientIp, "session %s", "login as %s with auth secret", "succeeded"}, m.RefID, clean.LogQuote(userName))
			event.LoginInfo(clientIp, "api", userName, m.UserAgent)
			return authn.ProviderClient, err
		}
	}

	// Otherwise, check account password.
	if user.WrongPassword(f.Password) {
		message := "incorrect password"
		limiter.Login.Reserve(clientIp)
		event.AuditErr([]string{clientIp, "session %s", "login as %s", message}, m.RefID, clean.LogQuote(userName))
		event.LoginError(clientIp, "api", userName, m.UserAgent, message)
		m.Status = http.StatusUnauthorized
		return authn.ProviderNone, i18n.Error(i18n.ErrInvalidCredentials)
	} else if m != nil {
		event.AuditInfo([]string{clientIp, "session %s", "login as %s", "succeeded"}, m.RefID, clean.LogQuote(userName))
		event.LoginInfo(clientIp, "api", userName, m.UserAgent)
	}

	return authn.ProviderLocal, nil
}

// LogIn performs authentication checks against the specified login form.
func (m *Session) LogIn(f form.Login, c *gin.Context) (err error) {
	if c != nil {
		m.SetContext(c)
	}

	var user *User
	var provider authn.ProviderType

	// Try to login with user credentials, if provided.
	if f.HasCredentials() {
		if m.IsRegistered() {
			m.Regenerate()
		}

		user, provider, err = Auth(f, m, c)

		if err != nil {
			return err
		}

		m.SetUser(user)
		m.SetProvider(provider)
	}

	// Try to redeem link share token, if provided.
	if f.HasShareToken() {
		user = m.User()

		// Redeem token.
		if user.IsRegistered() {
			if shares := user.RedeemToken(f.ShareToken); shares == 0 {
				limiter.Login.Reserve(m.IP())
				event.AuditWarn([]string{m.IP(), "session %s", "share token %s is invalid"}, m.RefID, clean.LogQuote(f.ShareToken))
				m.Status = http.StatusNotFound
				return i18n.Error(i18n.ErrInvalidLink)
			} else {
				event.AuditInfo([]string{m.IP(), "session %s", "token redeemed for %d shares"}, m.RefID, user.RedeemToken(f.ShareToken))
			}
		} else if data := m.Data(); data == nil {
			m.Status = http.StatusInternalServerError
			return i18n.Error(i18n.ErrUnexpected)
		} else if shares := data.RedeemToken(f.ShareToken); shares == 0 {
			limiter.Login.Reserve(m.IP())
			event.AuditWarn([]string{m.IP(), "session %s", "share token %s is invalid"}, m.RefID, clean.LogQuote(f.ShareToken))
			event.LoginError(m.IP(), "api", "", m.UserAgent, "invalid share token")
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
