package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// GET /s/:token/...
func Shares(router *gin.RouterGroup) {
	router.GET("/:token", func(c *gin.Context) {
		conf := service.Config()

		token := sanitize.Token(c.Param("token"))

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

		token := sanitize.Token(c.Param("token"))
		share := sanitize.Token(c.Param("share"))

		links := entity.FindValidLinks(token, share)

		if len(links) < 1 {
			log.Warn("share: invalid token or share")
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}

		uid := links[0].ShareUID
		clientConfig := conf.GuestConfig()

		if uid != share {
			c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf("%ss/%s/%s", clientConfig.SiteUrl, token, uid))
			return
		}

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
