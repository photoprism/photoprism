package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// GetConfigOptions returns backend config options.
//
//	@Summary	returns backend config options
//	@Id			GetConfigOptions
//	@Tags		Config, Settings
//	@Produce	json
//	@Success	200			{object}	config.Options
//	@Failure	401,403,429	{object}	i18n.Response
//	@Router		/api/v1/config/options [get]
func GetConfigOptions(router *gin.RouterGroup) {
	router.GET("/config/options", func(c *gin.Context) {
		s := Auth(c, acl.ResourceConfig, acl.AccessAll)
		conf := get.Config()

		// Abort if permission is not granted.
		if s.Invalid() || conf.Public() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		c.JSON(http.StatusOK, conf.Options())
	})
}

// SaveConfigOptions updates backend config options.
//
//	@Summary	updates backend config options
//	@Id			SaveConfigOptions
//	@Tags		Config, Settings
//	@Accept		json
//	@Produce	json
//	@Success	200					{object}	config.Options
//	@Failure	400,401,403,429,500	{object}	i18n.Response
//	@Param		options				body		config.Options	true	"properties to be updated (only submit values that should be changed)"
//	@Router		/api/v1/config/options [post]
func SaveConfigOptions(router *gin.RouterGroup) {
	router.POST("/config/options", func(c *gin.Context) {
		s := Auth(c, acl.ResourceConfig, acl.ActionManage)
		conf := get.Config()

		if s.Invalid() || conf.Public() || conf.DisableSettings() {
			AbortForbidden(c)
			return
		}

		v := make(entity.Values)

		LimitRequestBodyBytes(c, MaxSettingsRequestBytes)

		if err := c.BindJSON(&v); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			AbortBadRequest(c, err)
			return
		}

		if _, err := conf.SaveOptionsPatch(v); err != nil {
			log.Errorf("config: failed saving options patch (%s)", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Set restart flag.
		mutex.Restart.Store(true)

		// Update package defaults.
		conf.Propagate()

		// Flush session cache and update client config.
		entity.FlushSessionCache()
		UpdateClientConfig()

		// Return updated config options.
		c.JSON(http.StatusOK, conf.Options())
	})
}
