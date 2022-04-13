package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/service"
)

// GET /api/v1/settings
func GetSettings(router *gin.RouterGroup) {
	router.GET("/settings", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceSettings, acl.ActionRead)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		if settings := service.Config().Settings(); settings != nil {
			c.JSON(http.StatusOK, settings)
		} else {
			Abort(c, http.StatusNotFound, i18n.ErrNotFound)
		}
	})
}

// POST /api/v1/settings
func SaveSettings(router *gin.RouterGroup) {
	router.POST("/settings", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceSettings, acl.ActionUpdate)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if conf.DisableSettings() {
			AbortUnauthorized(c)
			return
		}

		settings := conf.Settings()

		if err := c.BindJSON(settings); err != nil {
			AbortBadRequest(c)
			return
		}

		if err := settings.Save(conf.SettingsYaml()); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		UpdateClientConfig()

		event.InfoMsg(i18n.MsgSettingsSaved)

		c.JSON(http.StatusOK, settings)
	})
}
