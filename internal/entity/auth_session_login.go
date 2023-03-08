package entity

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Auth checks if the credentials are valid and returns the user and authentication provider.
var Auth = func(f form.Login, m *Session, c *gin.Context) (user *User, provider string, err error) {
	name := f.Username()

	user = FindUserByName(name)
	err = AuthLocal(user, f, m)

	if err != nil {
		return user, authn.ProviderNone, err
	}

	// Update login timestamp.
	user.UpdateLoginTime()

	return user, authn.ProviderLocal, err
}

// AuthLocal authenticates against the local user database with the specified username and password.
func AuthLocal(user *User, f form.Login, m *Session) (err error) {
	name := f.Username()

	// User found?
	if user == nil {
		message := "account not found"
		if m != nil {
			limiter.Login.Reserve(m.IP())
			event.AuditWarn([]string{m.IP(), "session %s", "login as %s", message}, m.RefID, clean.LogQuote(name))
			event.LoginError(m.IP(), "api", name, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
		}
		return i18n.Error(i18n.ErrInvalidCredentials)
	}

	// Login allowed?
	if !user.CanLogIn() {
		message := "account disabled"
		if m != nil {
			event.AuditWarn([]string{m.IP(), "session %s", "login as %s", message}, m.RefID, clean.LogQuote(name))
			event.LoginError(m.IP(), "api", name, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
		}
		return i18n.Error(i18n.ErrInvalidCredentials)
	}

	// Password valid?
	if user.WrongPassword(f.Password) {
		message := "incorrect password"
		if m != nil {
			limiter.Login.Reserve(m.IP())
			event.AuditErr([]string{m.IP(), "session %s", "login as %s", message}, m.RefID, clean.LogQuote(name))
			event.LoginError(m.IP(), "api", name, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
		}
		return i18n.Error(i18n.ErrInvalidCredentials)
	} else if m != nil {
		event.AuditInfo([]string{m.IP(), "session %s", "login as %s", "succeeded"}, m.RefID, clean.LogQuote(name))
		event.LoginInfo(m.IP(), "api", name, m.UserAgent)
	}

	return err
}

// LogIn performs authentication checks against the specified login form.
func (m *Session) LogIn(f form.Login, c *gin.Context) (err error) {
	if c != nil {
		m.SetContext(c)
	}

	var user *User
	var provider string

	// Login credentials provided?
	if f.HasCredentials() {
		if m.IsRegistered() {
			m.RegenerateID()
		}

		user, provider, err = Auth(f, m, c)

		if err != nil {
			return err
		}

		m.SetUser(user)
		m.SetProvider(provider)
	}

	// Share token provided?
	if f.HasToken() {
		user = m.User()

		// Redeem token.
		if user.IsRegistered() {
			if shares := user.RedeemToken(f.AuthToken); shares == 0 {
				limiter.Login.Reserve(m.IP())
				event.AuditWarn([]string{m.IP(), "session %s", "share token %s is invalid"}, m.RefID, clean.LogQuote(f.AuthToken))
				m.Status = http.StatusNotFound
				return i18n.Error(i18n.ErrInvalidLink)
			} else {
				event.AuditInfo([]string{m.IP(), "session %s", "token redeemed for %d shares"}, m.RefID, user.RedeemToken(f.AuthToken))
			}
		} else if data := m.Data(); data == nil {
			m.Status = http.StatusInternalServerError
			return i18n.Error(i18n.ErrUnexpected)
		} else if shares := data.RedeemToken(f.AuthToken); shares == 0 {
			limiter.Login.Reserve(m.IP())
			event.AuditWarn([]string{m.IP(), "session %s", "share token %s is invalid"}, m.RefID, clean.LogQuote(f.AuthToken))
			event.LoginError(m.IP(), "api", "", m.UserAgent, "invalid share token")
			m.Status = http.StatusNotFound
			return i18n.Error(i18n.ErrInvalidLink)
		} else {
			m.SetData(data)
			event.AuditInfo([]string{m.IP(), "session %s", "token redeemed for %d shares"}, m.RefID, shares, data)
		}

		// Upgrade session to visitor.
		if user.IsUnknown() {
			user = &Visitor
			event.AuditDebug([]string{m.IP(), "session %s", "role upgraded to %s"}, m.RefID, user.AclRole().String())
			expires := UTC().Add(time.Hour * 24)
			m.Expires(expires)
			event.AuditDebug([]string{m.IP(), "session %s", "expires at %s"}, m.RefID, txt.TimeStamp(&expires))
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
