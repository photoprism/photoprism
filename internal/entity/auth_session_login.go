package entity

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/jinzhu/gorm"
	"net/http"
	"os"
	"strings"
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
var Auth = func(f form.Login, m *Session, c *gin.Context) (user *User, provider authn.ProviderType, err error) {
	name := f.Username()
	username := clean.Username(name)
	if os.Getenv("PHOTOPRISM_LDAP_ENABLED") == "true" {
		isAdmin, err := AuthLdap(username, f.Password)
		if err != nil {
			return nil, authn.ProviderNone, err
		}
		user = FindUserByName(username)
		if user == nil {
			user, err = CreateUser(username, isAdmin)
			if err != nil {
				return nil, authn.ProviderNone, err
			}
		}
	} else {
		user = FindUserByName(username)
		err = AuthLocal(user, f, m)
		if err != nil {
			return user, authn.ProviderNone, err
		}
	}
	// Update login timestamp.
	user.UpdateLoginTime()

	return user, authn.ProviderLocal, err
}

func CreateUser(username string, isAdmin bool) (*User, error) {
	user := NewUser()
	user.UserName = username
	user.SuperAdmin = isAdmin
	user.CanLogin = true
	err := Db().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		log.Infof("successfully added user %s", clean.LogQuote(user.Username()))
		return nil
	})
	if err != nil {
		log.Errorln("user save error", err)
		return nil, err
	}
	return user, nil
}

func AuthLdap(username, password string) (bool, error) {
	conn, err := ldap.DialURL(os.Getenv("PHOTOPRISM_LDAP_URI"))
	if err != nil {
		log.Errorln("ldap dial error", err)
		return false, i18n.Error(i18n.ErrInvalidCredentials)
	}
	defer conn.Close()

	bindDn := strings.ReplaceAll(os.Getenv("PHOTOPRISM_LDAP_BIND_DN"), "{username}", username)
	err = conn.Bind(bindDn, password)
	if err != nil {
		log.Errorln("ldap bind error", err)
		return false, i18n.Error(i18n.ErrInvalidCredentials)
	}
	isAdmin, err := isLdapAdmin(conn, username)
	if err != nil {
		return false, i18n.Error(i18n.ErrInvalidCredentials)
	}
	return isAdmin, nil
}

func isLdapAdmin(conn *ldap.Conn, username string) (bool, error) {
	searchRequest := ldap.NewSearchRequest(
		os.Getenv("PHOTOPRISM_LDAP_ADMIN_GROUP_DN"),
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		strings.ReplaceAll(os.Getenv("PHOTOPRISM_LDAP_ADMIN_GROUP_FILTER"), "{username}", username),
		[]string{os.Getenv("PHOTOPRISM_LDAP_ADMIN_GROUP_ATTRIBUTE")},
		nil)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		log.Errorln("admin search error", err)
		return false, err
	}

	if len(sr.Entries) < 1 {
		return false, nil
	}
	return true, nil
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
	if !user.Provider().IsDefault() && !user.Provider().IsLocal() {
		message := fmt.Sprintf("%s authentication disabled", authn.ProviderLocal.String())
		if m != nil {
			event.AuditWarn([]string{m.IP(), "session %s", "login as %s", message}, m.RefID, clean.LogQuote(name))
			event.LoginError(m.IP(), "api", name, m.UserAgent, message)
			m.Status = http.StatusUnauthorized
		}
		return i18n.Error(i18n.ErrInvalidCredentials)
	} else if !user.CanLogIn() {
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
