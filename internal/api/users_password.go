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
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/clean"
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

		// Check limit for failed auth requests (max. 10 per minute).
		if limiter.Login.Reject(ClientIP(c)) {
			limiter.AbortJSON(c)
			return
		}

		// Get session.
		s := Auth(c, acl.ResourcePassword, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Check if the session user is has user management privileges.
		isPrivileged := acl.Resources.AllowAll(acl.ResourceUsers, s.User().AclRole(), acl.Permissions{acl.AccessAll, acl.ActionManage})
		isSuperAdmin := isPrivileged && s.User().IsSuperAdmin()
		uid := clean.UID(c.Param("uid"))

		var u *entity.User

		// Users may only change their own password.
		if !isPrivileged && s.User().UserUID != uid {
			AbortForbidden(c)
			return
		} else if s.User().UserUID == uid {
			u = s.User()
			isPrivileged = false
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

		// Verify that the old password is correct.
		if isSuperAdmin && f.OldPassword == "" {
			// Do nothing.
		} else if u.WrongPassword(f.OldPassword) {
			limiter.Login.Reserve(ClientIP(c))
			Abort(c, http.StatusBadRequest, i18n.ErrInvalidPassword)
			return
		}

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

		// Invalidate all other user sessions to protect the account:
		// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
		event.AuditInfo([]string{ClientIP(c), "session %s", "password changed", "invalidated %s"}, s.RefID,
			english.Plural(u.DeleteSessions([]string{s.ID}), "session", "sessions"))

		AddTokenHeaders(c, s)
		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgPasswordChanged))
	})
}
