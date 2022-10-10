package entity

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/pkg/clean"
)

// SignIn checks user authentication based on the login form.
func (m *Session) SignIn(f form.Login, c *gin.Context) (err error) {
	if c != nil {
		m.SetContext(c)
	}

	// Username and password provided?
	if f.HasCredentials() {
		if m.IsRegistered() {
			m.RegenerateID()
		}

		name := f.Name()
		user := FindUserByName(name)

		// User found?
		if user == nil {
			message := "account not found"
			event.AuditWarn([]string{m.IP(), "session %s", "login as %s", message}, m.RefID, clean.LogQuote(name))
			event.LoginError(m.IP(), "api", name, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
			return i18n.Error(i18n.ErrInvalidCredentials)
		}

		// Login allowed?
		if !user.LoginAllowed() {
			message := "account disabled"
			event.AuditWarn([]string{m.IP(), "session %s", "login as %s", message}, m.RefID, clean.LogQuote(name))
			event.LoginError(m.IP(), "api", name, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
			return i18n.Error(i18n.ErrInvalidCredentials)
		}

		// Password valid?
		if user.InvalidPassword(f.Password) {
			message := "incorrect password"
			event.AuditErr([]string{m.IP(), "session %s", "login as %s", message}, m.RefID, clean.LogQuote(name))
			event.LoginError(m.IP(), "api", name, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
			return i18n.Error(i18n.ErrInvalidCredentials)
		} else {
			event.AuditInfo([]string{m.IP(), "session %s", "login as %s", "succeeded"}, m.RefID, clean.LogQuote(name))
			event.LoginInfo(m.IP(), "api", name, m.UserAgent)
		}

		m.SetUser(user)
	}

	// Share token provided?
	if f.HasToken() {
		user := m.User()

		// Redeem token.
		if user.IsRegistered() {
			if shares := user.RedeemToken(f.AuthToken); shares == 0 {
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
