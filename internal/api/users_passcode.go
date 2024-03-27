package api

import (
	"errors"
	"net/http"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// CreateUserPasscode sets up a new two-factor authentication passcode.
//
// POST /api/v1/users/:uid/passcode
func CreateUserPasscode(router *gin.RouterGroup) {
	router.POST("/users/:uid/passcode", func(c *gin.Context) {
		// Check authentication and authorization.
		s, user, frm, authErr := checkUserPasscodeAuth(c, acl.ActionCreate)

		if authErr != nil {
			return
		}

		// Check if the account password is correct.
		if user.WrongPassword(frm.Password) {
			limiter.Login.Reserve(ClientIP(c))
			Abort(c, http.StatusForbidden, i18n.ErrInvalidPassword)
			return
		}

		// Get config.
		conf := get.Config()

		// Generate and save new passcode key.
		var passcode *entity.Passcode
		if key, err := rnd.AuthKey(conf.AppName(), user.UserName); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "users", user.UserName, "failed to generate passcode", clean.Error(err)}, s.RefID)
			Abort(c, http.StatusInternalServerError, i18n.ErrUnexpected)
			return
		} else if passcode, err = entity.NewPasscode(user.UID(), key.String(), rnd.RecoveryCode()); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "users", user.UserName, "failed to create passcode", clean.Error(err)}, s.RefID)
			Abort(c, http.StatusInternalServerError, i18n.ErrUnexpected)
			return
		} else if err = passcode.Save(); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "users", user.UserName, "failed to save passcode", clean.Error(err)}, s.RefID)
			Abort(c, http.StatusConflict, i18n.ErrSaveFailed)
			return
		}

		event.AuditInfo([]string{ClientIP(c), "session %s", "users", user.UserName, "passcode", "created"}, s.RefID)

		c.JSON(http.StatusOK, passcode)
	})
}

// ConfirmUserPasscode checks a new passcode and flags it as verified so that it can be activated.
//
// POST /api/v1/users/:uid/passcode/confirm
func ConfirmUserPasscode(router *gin.RouterGroup) {
	router.POST("/users/:uid/passcode/confirm", func(c *gin.Context) {
		// Check authentication and authorization.
		s, user, frm, authErr := checkUserPasscodeAuth(c, acl.ActionUpdate)

		if authErr != nil {
			return
		}

		// Verify new passcode.
		valid, passcode, err := user.VerifyPasscode(frm.Passcode)

		if err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "users", user.UserName, "failed to verify passcode", clean.Error(err)}, s.RefID)
			Abort(c, http.StatusForbidden, i18n.ErrInvalidPasscode)
			return
		} else if !valid {
			event.AuditWarn([]string{ClientIP(c), "session %s", "users", user.UserName, "incorrect passcode"}, s.RefID)
			limiter.Login.ReserveN(ClientIP(c), 3)
			Abort(c, http.StatusForbidden, i18n.ErrInvalidPasscode)
			return
		}

		event.AuditInfo([]string{ClientIP(c), "session %s", "users", user.UserName, "passcode", "verified"}, s.RefID)

		// Clear session cache.
		s.ClearCache()

		c.JSON(http.StatusOK, passcode)
	})
}

// ActivateUserPasscode activates two-factor authentication if a passcode has been created and verified.
//
// POST /api/v1/users/:uid/passcode/activate
func ActivateUserPasscode(router *gin.RouterGroup) {
	router.POST("/users/:uid/passcode/activate", func(c *gin.Context) {
		// Check authentication and authorization.
		s, user, _, authErr := checkUserPasscodeAuth(c, acl.ActionUpdate)

		if authErr != nil {
			return
		}

		// Activate new passcode.
		passcode, err := user.ActivatePasscode()

		if err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "users", user.UserName, "failed to activate passcode", clean.Error(err)}, s.RefID)
			Abort(c, http.StatusForbidden, i18n.ErrSaveFailed)
			return
		}

		// Log event.
		event.AuditInfo([]string{ClientIP(c), "session %s", "users", user.UserName, "passcode", "activated"}, s.RefID)

		// Invalidate any other user sessions to protect the account:
		// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
		event.AuditInfo([]string{ClientIP(c), "session %s", "users", user.UserName, "invalidated %s"}, s.RefID,
			english.Plural(user.DeleteSessions([]string{s.ID}), "session", "sessions"))

		// Clear session cache.
		s.ClearCache()

		c.JSON(http.StatusOK, passcode)
	})
}

// DeactivateUserPasscode disables removes a passcode key to disable two-factor authentication.
//
// POST /api/v1/users/:uid/passcode/deactivate
func DeactivateUserPasscode(router *gin.RouterGroup) {
	router.POST("/users/:uid/passcode/deactivate", func(c *gin.Context) {
		// Check authentication and authorization.
		s, user, frm, authErr := checkUserPasscodeAuth(c, acl.ActionDelete)

		if authErr != nil {
			return
		}

		// Check if the account password is correct.
		if user.WrongPassword(frm.Password) {
			limiter.Login.Reserve(ClientIP(c))
			Abort(c, http.StatusForbidden, i18n.ErrInvalidPassword)
			return
		}

		// Delete passcode.
		if _, err := user.DeactivatePasscode(); err != nil {
			event.AuditErr([]string{ClientIP(c), "session %s", "users", user.UserName, "failed to deactivate passcode", clean.Error(err)}, s.RefID)
			Abort(c, http.StatusNotFound, i18n.ErrNotFound)
			return
		}

		event.AuditInfo([]string{ClientIP(c), "session %s", "users", user.UserName, "passcode", "deactivated"}, s.RefID)

		// Clear session cache.
		s.ClearCache()

		c.JSON(http.StatusOK, i18n.NewResponse(http.StatusOK, i18n.MsgSettingsSaved))
	})
}

// checkUserPasscodeAuth checks authentication and authorization for the passcode endpoints.
func checkUserPasscodeAuth(c *gin.Context, action acl.Permission) (*entity.Session, *entity.User, *form.UserPasscode, error) {
	conf := get.Config()

	// Prevent caching of API response.
	c.Header(header.CacheControl, header.CacheControlNoStore)

	// You cannot change any passwords without authentication and settings enabled.
	if conf.Public() || conf.DisableSettings() {
		Abort(c, http.StatusForbidden, i18n.ErrPublic)
		return nil, nil, nil, errors.New("unsupported")
	}

	// Check limit for failed auth requests (max. 10 per minute).
	if limiter.Login.Reject(ClientIP(c)) {
		limiter.AbortJSON(c)
		return nil, nil, nil, errors.New("rate limit exceeded")
	}

	// Get session.
	s := Auth(c, acl.ResourcePasscode, action)

	if s.Abort(c) {
		return s, nil, nil, errors.New("unauthorized")
	}

	// Check if the current user has management privileges.
	uid := clean.UID(c.Param("uid"))

	// Get user from session.
	user := s.User()

	// Regular users can only set up a passcode for their own account.
	if user.UserUID != uid {
		AbortForbidden(c)
		return s, nil, nil, errors.New("unauthorized")
	}

	// Check if the auth provider supports passcodes.
	if !user.Provider().Supports2FA() {
		Abort(c, http.StatusForbidden, i18n.ErrUnsupported)
		return s, nil, nil, errors.New("unsupported")
	}

	frm := &form.UserPasscode{}

	// Validate request parameters.
	if err := c.BindJSON(frm); err != nil {
		Error(c, http.StatusBadRequest, err, i18n.ErrInvalidPassword)
		return s, nil, nil, errors.New("invalid request")
	} else if authn.KeyTOTP.NotEqual(frm.Type) {
		Abort(c, http.StatusBadRequest, i18n.ErrUnsupportedType)
		return s, nil, nil, errors.New("unsupported")
	}

	return s, user, frm, nil
}
