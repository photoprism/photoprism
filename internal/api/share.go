package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
)

// GET /s/:token/...
func Shares(router *gin.RouterGroup) {
	router.GET("/:token", func(c *gin.Context) {
		conf := service.Config()

		token := c.Param("token")

		links := entity.FindValidLinks(token, "")

		if len(links) == 0 {
			log.Warn("share: invalid token")
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}

		clientConfig := conf.GuestConfig()
		clientConfig.SiteUrl = fmt.Sprintf("%ss/%s", clientConfig.SiteUrl, token)

		c.HTML(http.StatusOK, "share.tmpl", gin.H{"config": clientConfig})
	})

	router.GET("/:token/:share", func(c *gin.Context) {
		conf := service.Config()

		token := c.Param("token")
		share := c.Param("share")

		links := entity.FindValidLinks(token, share)

		if len(links) != 1 {
			log.Warn("share: invalid token or share")
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}

		uid := links[0].ShareUID

		if uid != share {
			c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf("/s/%s/%s", token, uid))
			return
		}

		clientConfig := conf.GuestConfig()
		clientConfig.SiteUrl = fmt.Sprintf("%ss/%s/%s", clientConfig.SiteUrl, token, uid)
		clientConfig.SitePreview = fmt.Sprintf("%s/preview", clientConfig.SiteUrl)

		if a, err := query.AlbumByUID(uid); err == nil {
			clientConfig.SiteCaption = a.AlbumTitle

			if a.AlbumDescription != "" {
				clientConfig.SiteDescription = a.AlbumDescription
			}
		}

		c.HTML(http.StatusOK, "share.tmpl", gin.H{"config": clientConfig})
	})
}
