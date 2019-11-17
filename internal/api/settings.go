package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/util"
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
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		s := conf.Settings()

		if err := c.BindJSON(s); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": util.UcFirst(err.Error())})
			return
		}

		if err := s.WriteValuesToFile(conf.SettingsFile()); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		event.Publish("config.updated", event.Data(conf.ClientConfig()))

		c.JSON(http.StatusOK, gin.H{"message": "saved"})
	})
}
