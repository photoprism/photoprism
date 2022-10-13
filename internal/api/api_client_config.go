package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service"
)

// UpdateClientConfig publishes updated client configuration values over the websocket connections.
func UpdateClientConfig() {
	event.Publish("config.updated", event.Data{"config": service.Config().ClientUser(false)})
}

// GetClientConfig returns the client configuration values as JSON.
//
// GET /api/v1/config
func GetClientConfig(router *gin.RouterGroup) {
	router.GET("/config", func(c *gin.Context) {
		s := Session(SessionID(c))
		conf := service.Config()

		if s == nil {
			c.JSON(http.StatusOK, conf.ClientPublic())
		} else {
			c.JSON(http.StatusOK, conf.ClientSession(s))
		}
	})
}
