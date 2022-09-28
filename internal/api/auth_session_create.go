package api

import (
	"net/http"

	"github.com/photoprism/photoprism/internal/config"

	"github.com/photoprism/photoprism/internal/acl"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

// CreateSession creates a new client session and returns it as JSON if authentication was successful.
//
// POST /api/v1/session
func CreateSession(router *gin.RouterGroup) {
	router.POST("/session", func(c *gin.Context) {
		var err error
		var f form.Login

		if err = c.BindJSON(&f); err != nil {
			event.AuditWarn([]string{ClientIP(c), "invalid create session request"})
			AbortBadRequest(c)
			return
		}

		var user *entity.User
		var sess *entity.Session
		var data *entity.SessionData

		id := SessionID(c)

		// Search existing session.
		if s := Session(id); s != nil {
			sess = s
			data = s.Data()
			user = s.User()
		} else {
			data = entity.NewSessionData()
			user = &entity.User{}
			id = ""
		}

		conf := service.Config()

		// Share token provided?
		if f.HasToken() {
			if shares := data.RedeemToken(f.AuthToken); shares == 0 {
				event.AuditWarn([]string{ClientIP(c), "share token %s", "invalid"}, clean.LogQuote(f.AuthToken))
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": i18n.Msg(i18n.ErrInvalidLink)})
				event.LoginError(ClientIP(c), "", UserAgent(c), "invalid share token")
				return
			}

			event.AuditInfo([]string{ClientIP(c), "share token %s", "redeemed", "%#v"}, clean.LogQuote(f.AuthToken), data)

			// Upgrade from Unknown to Visitor. Don't downgrade.
			if user.IsUnknown() {
				user = &entity.Visitor
				event.AuditDebug([]string{ClientIP(c), "share token %s", "upgrading session to user role %s"}, clean.LogQuote(f.AuthToken), acl.RoleVisitor.String())
			}
		} else if f.HasCredentials() {
			// If not, authenticate with username and password.
			userName := f.Name()
			user = entity.FindUserByName(userName)

			// User found?
			if user == nil {
				message := "account not found"
				event.AuditWarn([]string{ClientIP(c), "login as %s", message}, clean.LogQuote(userName))
				c.AbortWithStatusJSON(400, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
				event.LoginError(ClientIP(c), f.Name(), UserAgent(c), message)
				return
			}

			// Login allowed?
			if !user.LoginAllowed() {
				message := "account disabled"
				event.AuditWarn([]string{ClientIP(c), "login as %s", message}, clean.LogQuote(userName))
				c.AbortWithStatusJSON(400, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
				event.LoginError(ClientIP(c), f.Name(), UserAgent(c), message)
				return
			}

			// Password valid?
			if user.InvalidPassword(f.Password) {
				message := "incorrect password"
				event.AuditErr([]string{ClientIP(c), "login as %s", message}, clean.LogQuote(userName))
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
				event.LoginError(ClientIP(c), f.Name(), UserAgent(c), message)
				return
			} else {
				event.AuditInfo([]string{ClientIP(c), "login as %s", "succeeded"}, clean.LogQuote(userName))
				event.LoginSuccess(ClientIP(c), f.Name(), UserAgent(c))
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": i18n.Msg(i18n.ErrInvalidCredentials)})
			event.LoginError(ClientIP(c), f.Name(), UserAgent(c), "invalid request")
			return
		}

		// Save session.
		if sess, err = service.Session().Save(id, user, c, data); err != nil {
			event.AuditWarn([]string{ClientIP(c), "%s"}, err)
		} else if sess == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": i18n.Msg(i18n.ErrUnexpected)})
			return
		}

		// Log event.
		event.AuditInfo([]string{ClientIP(c), "session %s", "created"}, sess.RefID)

		// Add session id to response headers.
		AddSessionHeader(c, sess.ID)

		var clientConfig config.ClientConfig

		if sess.User().IsVisitor() {
			clientConfig = conf.ClientShare()
		} else if sess.User().IsRegistered() {
			clientConfig = conf.ClientSession(sess)
		} else {
			clientConfig = conf.ClientPublic()
		}

		// Send JSON response with user information, session data, and client config values.
		c.JSON(http.StatusOK, gin.H{"status": "ok", "id": sess.ID, "user": sess.User(), "data": sess.Data(), "config": clientConfig})
	})
}
