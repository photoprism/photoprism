package api

import (
	"net/http"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ChangePassword changes the password of the currently authenticated user.
//
// PUT /api/v1/users/:uid/password
func ChangePassword(router *gin.RouterGroup) {
	router.PUT("/users/:uid/password", func(c *gin.Context) {
		conf := service.Config()

		// You cannot change any passwords without authentication and settings enabled.
		if conf.Public() || conf.DisableSettings() {
			Abort(c, http.StatusForbidden, i18n.ErrPublic)
			return
		}

		// Check limit for failed auth requests (max. 10 per minute).
		if limiter.Auth.Reject(ClientIP(c)) {
			limiter.AbortJSON(c)
			return
		}

		// Get session.
		s := Auth(c, acl.ResourcePassword, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))
		m := entity.FindUserByUID(uid)

		// Users may only change their own password.
		if s.User().UserUID != m.UserUID {
			AbortForbidden(c)
			return
		}

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrUserNotFound)
			return
		}

		f := form.ChangePassword{}

		if err := c.BindJSON(&f); err != nil {
			Error(c, http.StatusBadRequest, err, i18n.ErrInvalidPassword)
			return
		}

		// Verify that the old password is correct.
		if m.WrongPassword(f.OldPassword) {
			limiter.Auth.Reserve(ClientIP(c))
			Abort(c, http.StatusBadRequest, i18n.ErrInvalidPassword)
			return
		}

		// Change password.
		if err := m.SetPassword(f.NewPassword); err != nil {
			Error(c, http.StatusBadRequest, err, i18n.ErrInvalidPassword)
			return
		}

		// Invalidate all other user sessions to protect the account:
		// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
		event.AuditInfo([]string{ClientIP(c), "session %s", "password changed", "invalidated %s"}, s.RefID,
			english.Plural(m.DeleteSessions([]string{s.ID}), "session", "sessions"))

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgPasswordChanged))
	})
}
