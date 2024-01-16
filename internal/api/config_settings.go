package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/i18n"
)

// GetSettings returns the user app settings as JSON.
//
// GET /api/v1/settings
func GetSettings(router *gin.RouterGroup) {
	router.GET("/settings", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourceSettings, acl.Permissions{acl.AccessAll, acl.AccessOwn})

		// Abort if permission was not granted.
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

// SaveSettings saved the user app settings.
//
// POST /api/v1/settings
func SaveSettings(router *gin.RouterGroup) {
	router.POST("/settings", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourceSettings, acl.Permissions{acl.ActionView, acl.ActionUpdate, acl.ActionManage})

		// Abort if permission was not granted.
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

		// Only super admins can change global config defaults.
		if s.User().IsSuperAdmin() {
			settings = conf.Settings()

			if err := c.BindJSON(settings); err != nil {
				AbortBadRequest(c)
				return
			}

			if err := settings.Save(conf.SettingsYaml()); err != nil {
				log.Debugf("config: %s (save app settings)", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, err)
				return
			}

			// Flush session cache and update client config.
			entity.FlushSessionCache()
			UpdateClientConfig()
		} else {
			// Apply to user preferences and keep current values if unspecified.
			user := s.User()

			if user == nil {
				AbortUnexpectedError(c)
				return
			}

			settings = &customize.Settings{}

			if err := c.BindJSON(settings); err != nil {
				AbortBadRequest(c)
				return
			}

			if acl.Resources.DenyAll(acl.ResourceSettings, s.User().AclRole(), acl.Permissions{acl.ActionUpdate, acl.ActionManage}) {
				c.JSON(http.StatusOK, user.Settings().Apply(settings).ApplyTo(conf.Settings().ApplyACL(acl.Resources, user.AclRole())))
				return
			} else if err := user.Settings().Apply(settings).Save(); err != nil {
				log.Debugf("config: %s (save user settings)", err)
				AbortSaveFailed(c)
				return
			}
		}

		// Return updated user settings.
		c.JSON(http.StatusOK, get.Config().SessionSettings(s))
	})
}
