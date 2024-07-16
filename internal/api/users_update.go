package api

import (
	"net/http"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// UpdateUser updates the profile information of the currently authenticated user.
//
//	@Tags	Users
//	@Route /api/v1/users/{uid} [put]
func UpdateUser(router *gin.RouterGroup) {
	router.PUT("/users/:uid", func(c *gin.Context) {
		conf := get.Config()

		if conf.Public() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		// Check if the session user is allowed to manage all accounts or update his/her own account.
		s := AuthAny(c, acl.ResourceUsers, acl.Permissions{acl.ActionManage, acl.AccessOwn, acl.ActionUpdate})

		if s.Abort(c) {
			return
		}

		// UserUID.
		uid := clean.UID(c.Param("uid"))

		// Find user.
		m := entity.FindUserByUID(uid)

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrUserNotFound)
			return
		}

		// Init form with model values.
		f, err := m.Form()

		if err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// Assign and validate request form values.
		if err = c.BindJSON(&f); err != nil {
			log.Error(err)
			AbortBadRequest(c)
			return
		}

		// Check if the session user is has user management privileges.
		isAdmin := acl.Rules.AllowAll(acl.ResourceUsers, s.UserRole(), acl.Permissions{acl.AccessAll, acl.ActionManage})
		privilegeLevelChange := isAdmin && m.PrivilegeLevelChange(f)

		// Get user from session.
		u := s.User()

		// Save model with values from form.
		if err = m.SaveForm(f, u); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "users", m.UserName, "update", err.Error()}, s.RefID)
			AbortSaveFailed(c)
			return
		}

		// Log event.
		event.AuditInfo([]string{ClientIP(c), "session %s", "users", m.UserName, "updated"}, s.RefID)

		// Delete user sessions after a privilege level change.
		// see https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html#renew-the-session-id-after-any-privilege-level-change
		if privilegeLevelChange {
			// Prevent the current session from being deleted.
			deleted := m.DeleteSessions([]string{s.ID})
			// Delete active user sessions.
			event.AuditInfo([]string{ClientIP(c), "session %s", "users", m.UserName, "invalidated %s"}, s.RefID,
				english.Plural(deleted, "session", "sessions"))
		}

		// Flush session cache.
		if isAdmin {
			entity.FlushSessionCache()
		} else {
			s.ClearCache()
		}

		// Find and return the updated user record.
		m = entity.FindUserByUID(uid)

		if m == nil {
			AbortEntityNotFound(c)
			return
		}

		c.JSON(http.StatusOK, m)
	})
}
