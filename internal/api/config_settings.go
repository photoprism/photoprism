package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/time/tz"
)

// GetSettings returns the user app settings as JSON.
//
//	@Summary	returns the user app settings as JSON
//	@Id			GetSettings
//	@Tags		Settings
//	@Produce	json
//	@Success	200			{object}	customize.Settings
//	@Failure	401,403,404	{object}	i18n.Response
//	@Router		/api/v1/settings [get]
func GetSettings(router *gin.RouterGroup) {
	router.GET("/settings", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourceSettings, acl.Permissions{acl.AccessAll, acl.AccessOwn})

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		settings := get.Config().SessionSettings(s)

		if settings == nil {
			Abort(c, http.StatusNotFound, i18n.ErrNotFound)
			return
		}

		c.JSON(http.StatusOK, settings)
	})
}

// SaveSettings saves the user app settings.
//
//	@Summary	saves the user app settings
//	@Id			SaveSettings
//	@Tags		Settings
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	customize.Settings
//	@Failure	400,401,403,404,413,500	{object}	i18n.Response
//	@Param		settings				body		customize.Settings	true	"user settings"
//	@Router		/api/v1/settings [post]
func SaveSettings(router *gin.RouterGroup) {
	router.POST("/settings", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourceSettings, acl.Permissions{acl.ActionView, acl.ActionUpdate, acl.ActionManage})

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		conf := get.Config()

		// Settings disabled?
		if conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		var settings *customize.Settings

		user := s.GetUser()

		// Update user preferences without changing global defaults.
		if user == nil {
			AbortUnexpectedError(c)
			return
		}

		// Only super admins without scope restrictions can change global config defaults.
		if user.IsSuperAdmin() && s.NoScope() {
			// Update global defaults and user preferences.
			settings = conf.Settings()
			timeZone := settings.UI.TimeZone

			// Set values from request.
			LimitRequestBodyBytes(c, MaxSettingsRequestBytes)

			if err := c.BindJSON(settings); err != nil {
				if IsRequestBodyTooLarge(err) {
					AbortRequestTooLarge(c, i18n.ErrBadRequest)
					return
				}

				AbortBadRequest(c, err)
				return
			}

			// Keep the previous default time zone if the new one is not a valid name.
			if zone := tz.Sanitize(settings.UI.TimeZone); zone != "" {
				settings.UI.TimeZone = zone
			} else {
				settings.UI.TimeZone = timeZone
			}

			// Update global defaults.
			if err := settings.Save(conf.SettingsYaml()); err != nil {
				settingsErr(s, "app", err)
				AbortSaveFailed(c)
				return
			}

			// Update user preferences.
			if err := user.Settings().Apply(settings).Save(); err != nil {
				settingsErr(s, "user", err)
				AbortSaveFailed(c)
				return
			}

			// Flush session cache and update client config
			// after global settings have been updated.
			entity.FlushSessionCache()
			UpdateClientConfig()
		} else {
			settings = &customize.Settings{}

			// Set values from request.
			LimitRequestBodyBytes(c, MaxSettingsRequestBytes)

			if err := c.BindJSON(settings); err != nil {
				if IsRequestBodyTooLarge(err) {
					AbortRequestTooLarge(c, i18n.ErrBadRequest)
					return
				}

				AbortBadRequest(c, err)
				return
			}

			// Update user preferences.
			if acl.Rules.DenyAll(acl.ResourceSettings, s.GetUserRole(), acl.Permissions{acl.ActionUpdate, acl.ActionManage}) {
				c.JSON(http.StatusOK, user.Settings().Apply(settings).ApplyTo(conf.Settings().ApplyACL(acl.Rules, user.AclRole())))
				return
			} else if err := user.Settings().Apply(settings).Save(); err != nil {
				settingsErr(s, "user", err)
				AbortSaveFailed(c)
				return
			}
		}

		// Return updated user settings.
		c.JSON(http.StatusOK, get.Config().SessionSettings(s))
	})
}

// settingsErr reports a failed save of the app or user settings on the console channel if err is not nil.
func settingsErr(s *entity.Session, kind string, err error) {
	if err != nil {
		event.SystemError([]string{"settings", "session %s", "failed to save %s settings", "%s"}, s.RefID, kind, clean.ErrorFull(err))
	}
}
