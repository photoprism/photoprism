package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
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
		conf := get.Config()

		if s := AuthAny(c, acl.ResourceConfig, acl.Permissions{acl.ActionView}); s.Valid() {
			c.JSON(http.StatusOK, conf.ClientSession(s))
			return
		} else if conf.DisableFrontend() {
			AbortUnauthorized(c)
			return
		}

		// Return public client config for loading the web frontend
		c.JSON(http.StatusOK, conf.ClientPublic())
	})
}
