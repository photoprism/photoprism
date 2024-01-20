package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
)

// UpdateClientConfig publishes updated client configuration values over the websocket connections.
func UpdateClientConfig() {
	event.Publish("config.updated", event.Data{"config": get.Config().ClientUser(false)})
}

// GetClientConfig returns the client configuration values as JSON.
//
// GET /api/v1/config
func GetClientConfig(router *gin.RouterGroup) {
	router.GET("/config", func(c *gin.Context) {
		sess := Session(ClientIP(c), AuthToken(c))
		conf := get.Config()

		if sess == nil {
			c.JSON(http.StatusOK, conf.ClientPublic())
		} else {
			c.JSON(http.StatusOK, conf.ClientSession(sess))
		}
	})
}
