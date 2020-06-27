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

		shareToken := c.Param("token")

		links := entity.FindValidLinks(shareToken, "")

		if len(links) == 0 {
			log.Warn("share: invalid token")
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}

		clientConfig := conf.GuestConfig()
		clientConfig.SiteUrl = fmt.Sprintf("%ss/%s", clientConfig.SiteUrl, shareToken)

		c.HTML(http.StatusOK, "share.tmpl", gin.H{"config": clientConfig})
	})

	router.GET("/:token/:uid", func(c *gin.Context) {
		conf := service.Config()

		shareToken := c.Param("token")
		share := c.Param("uid")

		links := entity.FindValidLinks(shareToken, share)

		if len(links) != 1 {
			log.Warn("share: invalid token or uid")
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}

		uid := links[0].ShareUID

		if uid != share {
			c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf("/s/%s/%s", shareToken, uid))
			return
		}

		clientConfig := conf.GuestConfig()
		clientConfig.SiteUrl = fmt.Sprintf("%ss/%s/%s", clientConfig.SiteUrl, shareToken, uid)
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
