package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// UpdateClientConfig publishes updated client configuration values over the websocket connections.
func UpdateClientConfig() {
	go func() {
		event.Publish("config.updated", event.Data{"config": get.Config().ClientUser(false)})
	}()
}

// GetClientConfig returns the client configuration values as JSON.
//
//	@Summary	get client configuration
//	@Id			GetClientConfig
//	@Tags		Config
//	@Produce	json
//	@Success	200	{object}	gin.H
//	@Failure	401	{object}	i18n.Response
//	@Router		/api/v1/config [get]
func GetClientConfig(router *gin.RouterGroup) {
	router.GET("/config", func(c *gin.Context) {
		sess := Session(ClientIP(c), AuthToken(c))
		conf := get.Config()

		// Check authentication.
		if sess != nil {
			// Return custom client config for authenticated user.
			c.JSON(http.StatusOK, conf.ClientSession(sess))
			return
		} else if conf.DisableFrontend() {
			// Abort if not authenticated, and the web frontend is disabled.
			AbortUnauthorized(c)
			return
		}

		// Return public client config for loading the web frontend
		c.JSON(http.StatusOK, conf.ClientPublic())
	})
}
