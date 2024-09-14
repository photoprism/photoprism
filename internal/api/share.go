package api

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Shares handles link share
//
// GET /s/:token/...
func Shares(router *gin.RouterGroup) {
	router.GET("/:token", func(c *gin.Context) {
		conf := get.Config()

		token := clean.Token(c.Param("token"))
		links := entity.FindValidLinks(token, "")

		if len(links) == 0 {
			log.Debugf("share: invalid token")
			c.Redirect(http.StatusTemporaryRedirect, conf.BaseUri(""))
			return
		}

		clientConfig := conf.ClientShare()
		clientConfig.SiteUrl = clientConfig.SiteUrl + path.Join("s", token)

		uri := conf.BaseUri("/library/albums")
		c.HTML(http.StatusOK, "share.gohtml", gin.H{"shared": gin.H{"token": token, "uri": uri}, "config": clientConfig})
	})

	router.GET("/:token/:shared", func(c *gin.Context) {
		conf := get.Config()

		token := clean.Token(c.Param("token"))
		shared := clean.Token(c.Param("shared"))

		links := entity.FindValidLinks(token, shared)

		if len(links) < 1 {
			log.Debugf("share: invalid token or slug")
			c.Redirect(http.StatusTemporaryRedirect, conf.BaseUri(""))
			return
		}

		uid := links[0].ShareUID
		clientConfig := conf.ClientShare()
		clientConfig.SiteUrl = clientConfig.SiteUrl + path.Join("s", token, uid)
		clientConfig.SitePreview = clientConfig.SiteUrl + "/preview"

		if a, err := query.AlbumByUID(uid); err == nil {
			clientConfig.SiteCaption = a.AlbumTitle

			if a.AlbumDescription != "" {
				clientConfig.SiteDescription = a.AlbumDescription
			}
		}

		uri := conf.BaseUri(path.Join("/library/albums", uid, shared))

		c.HTML(http.StatusOK, "share.gohtml", gin.H{"shared": gin.H{"token": token, "uri": uri}, "config": clientConfig})
	})
}
