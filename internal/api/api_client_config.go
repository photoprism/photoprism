package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// UpdateClientConfig publishes updated client configuration values over the websocket connections.
func UpdateClientConfig() {
	if !entity.HasDbProvider() {
		return
	}

	conf := get.Config()
	if conf == nil {
		return
	}

	clientConfig := conf.ClientUser(false)

	// Do not broadcast session-specific tokens via the config.updated event,
	// as this is a global broadcast to all connected clients and may overwrite
	// individual session tokens with the wrong values. Session tokens are
	// updated via the X-Preview-Token and X-Download-Token response headers.
	clientConfig.PreviewToken = ""
	clientConfig.DownloadToken = ""

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Warnf("api: failed to publish updated client config (%v)", r)
			}
		}()

		event.Publish("config.updated", event.Data{"config": clientConfig})
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
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

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
