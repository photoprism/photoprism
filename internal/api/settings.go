package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GET /api/v1/settings
func GetSettings(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/settings", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		s := conf.Settings()

		c.JSON(http.StatusOK, s)
	})
}

// POST /api/v1/settings
func SaveSettings(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/settings", func(c *gin.Context) {
		if conf.DisableSettings() || Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		s := conf.Settings()

		if err := c.BindJSON(s); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if err := s.Save(conf.SettingsFile()); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		event.Publish("config.updated", event.Data(conf.ClientConfig()))
		log.Infof("settings saved")

		c.JSON(http.StatusOK, s)
	})
}
