package api

import (
	"net/http"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// UpdateUser updates profile information for the specified user.
//
//	@Summary	update user profile information
//	@Id			UpdateUser
//	@Tags		Users
//	@Accept		json
//	@Produce	json
//	@Param		uid						path		string		true	"user uid"
//	@Param		user					body		form.User	true	"properties to be updated"
//	@Success	200						{object}	entity.User
//	@Failure	400,401,403,404,409,429	{object}	i18n.Response
//	@Router		/api/v1/users/{uid} [put]
func UpdateUser(router *gin.RouterGroup) {
	router.PUT("/users/:uid", func(c *gin.Context) {
		conf := get.Config()

		if conf.Public() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		// Require user management or own-account update access.
		s := AuthAny(c, acl.ResourceUsers, acl.Permissions{acl.ActionManage, acl.AccessOwn, acl.ActionUpdate, acl.ActionUpdateOwn})

		if s.Abort(c) {
			return
		}

		// A verified Portal cluster JWT (GrantJwtBearer) with users-manage scope is
		// a trusted service principal — the Portal syncing cluster user state — with
		// no end-user identity, so authorize it for user management like an admin
		// instead of applying the per-user owner check below (which a user-less
		// service token can never satisfy).
		isClusterJWT := s.GrantType == authn.GrantJwtBearer.String() && s.ValidateScope(acl.ResourceUsers, acl.Permissions{acl.ActionManage})

		// Check whether the role can manage all user accounts.
		isAdmin := isClusterJWT || acl.Rules.AllowAll(acl.ResourceUsers, s.GetUserRole(), acl.Permissions{acl.AccessAll, acl.ActionManage})
		uid := clean.UID(c.Param("uid"))

		// Non-admin users may only update their own profile.
		if !isAdmin && s.GetUser().UserUID != uid {
			event.AuditErr([]string{ClientIP(c), "session %s", "users", clean.Log(uid), "update", status.Denied}, s.RefID)
			AbortForbidden(c)
			return
		}

		// Find user.
		m := entity.FindUserByUID(uid)

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrUserNotFound)
			return
		}

		// System accounts (Unknown id=-1, Visitor id=-2) must not be modified.
		if m.ID < 0 {
			event.AuditErr([]string{ClientIP(c), "session %s", "users", clean.Log(uid), "update", status.Denied}, s.RefID)
			AbortForbidden(c)
			return
		}

		// Initialize form with model values.
		f, err := m.Form()

		if err != nil {
			log.Error(err)
			AbortSaveFailed(c)
			return
		}

		// Assign and validate request form values.
		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err = c.BindJSON(&f); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			AbortBadRequest(c, err)
			return
		}

		privilegeLevelChange := isAdmin && m.PrivilegeLevelChange(f)

		// Check if the user account quota has been exceeded.
		if f.UserRole != "" && m.UserRole != f.UserRole && !conf.UsersQuotaReached(acl.ParseRole(m.UserRole)) && conf.UsersQuotaReached(acl.ParseRole(f.UserRole)) {
			event.AuditErr([]string{ClientIP(c), "session %s", "users", m.UserName, "update", authn.ErrUsersQuotaExceeded.Error()}, s.RefID)
			AbortQuotaExceeded(c)
			return
		}

		// Get user from session.
		u := s.GetUser()

		// Prevent users from changing their own account role, disabling their
		// own super admin status, or revoking their own web login — any of
		// which could lock an operator out of the admin UI (e.g. a super admin
		// demoting themselves). Other own-profile fields remain editable.
		if u != nil && u.UserUID == m.UserUID {
			switch {
			case f.UserRole != "" && clean.Role(f.UserRole) != clean.Role(m.UserRole):
				event.AuditErr([]string{ClientIP(c), "session %s", "users", m.UserName, "update own role", status.Denied}, s.RefID)
				AbortForbidden(c)
				return
			case m.SuperAdmin && !f.SuperAdmin:
				event.AuditErr([]string{ClientIP(c), "session %s", "users", m.UserName, "disable own super admin status", status.Denied}, s.RefID)
				AbortForbidden(c)
				return
			case m.CanLogin && !f.CanLogin:
				event.AuditErr([]string{ClientIP(c), "session %s", "users", m.UserName, "disable own web login", status.Denied}, s.RefID)
				AbortForbidden(c)
				return
			}
		}

		// Persist form values. SaveForm gates privilege-level fields on admin
		// authorization; the cluster JWT is a user-less service principal that
		// u.IsAdmin() rejects, so pass that decision explicitly or its login and
		// role sync would be silently dropped.
		if err = m.SaveForm(f, u, u.IsAdmin() || isClusterJWT); err != nil {
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
			if f.UserRole != "" {
				config.FlushUsageCache()
				UpdateClientConfig()
			}
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
