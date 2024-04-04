package api

import (
	"net/http"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// UpdateUserPassword changes the password of the currently authenticated user.
//
// PUT /api/v1/users/:uid/password
func UpdateUserPassword(router *gin.RouterGroup) {
	router.PUT("/users/:uid/password", func(c *gin.Context) {
		conf := get.Config()

		// You cannot change any passwords without authentication and settings enabled.
		if conf.Public() || conf.DisableSettings() {
			Abort(c, http.StatusForbidden, i18n.ErrPublic)
			return
		}

		// Get session.
		s := Auth(c, acl.ResourcePassword, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Get client IP address.
		clientIp := ClientIP(c)

		// Check request rate limit.
		r := limiter.Login.Request(clientIp)

		if r.Reject() {
			limiter.AbortJSON(c)
			return
		}

		// Check if the current user has management privileges.
		isAdmin := acl.Rules.AllowAll(acl.ResourceUsers, s.UserRole(), acl.Permissions{acl.AccessAll, acl.ActionManage})
		isSuperAdmin := isAdmin && s.User().IsSuperAdmin()
		uid := clean.UID(c.Param("uid"))

		var u *entity.User

		// Regular users may only change their own password.
		if !isAdmin && s.User().UserUID != uid {
			AbortForbidden(c)
			return
		} else if s.User().UserUID == uid {
			u = s.User()
			isAdmin = false
			isSuperAdmin = false
		} else if u = entity.FindUserByUID(uid); u == nil {
			Abort(c, http.StatusNotFound, i18n.ErrUserNotFound)
			return
		}

		f := form.ChangePassword{}

		if err := c.BindJSON(&f); err != nil {
			Error(c, http.StatusBadRequest, err, i18n.ErrInvalidPassword)
			return
		}

		// Check password and abort if invalid.
		if isSuperAdmin && f.OldPassword == "" {
			// Ignore if a super admin performs the change for another account.
		} else if u.InvalidPassword(f.OldPassword) {
			Abort(c, http.StatusBadRequest, i18n.ErrInvalidPassword)
			return
		}

		// Return the reserved request rate limit tokens after successful authentication.
		r.Success()

		// Set new password.
		if err := u.SetPassword(f.NewPassword); err != nil {
			Error(c, http.StatusBadRequest, err, i18n.ErrInvalidPassword)
			return
		}

		// Update tokens if user matches with session.
		if s.User().UserUID == u.UID() {
			s.SetPreviewToken(u.PreviewToken)
			s.SetDownloadToken(u.DownloadToken)
		}

		// Log event.
		event.AuditInfo([]string{ClientIP(c), "session %s", "users", u.UserName, "password", "changed"}, s.RefID)

		// Invalidate any other user sessions to protect the account:
		// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
		event.AuditInfo([]string{ClientIP(c), "session %s", "users", u.UserName, "invalidated %s"}, s.RefID,
			english.Plural(u.DeleteSessions([]string{s.ID}), "session", "sessions"))

		AddTokenHeaders(c, s)
		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgPasswordChanged))
	})
}
