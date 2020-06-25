package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/txt"
)

func shareHandler(c *gin.Context, conf *config.Config) {
	token := c.Param("token")
	links := entity.FindLinks(token, "")

	if len(links) == 0 {
		log.Warnf("sharing: invalid token %s", txt.Quote(token))
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	clientConfig := conf.GuestConfig()

	c.HTML(http.StatusOK, "share.tmpl", gin.H{"config": clientConfig})
}

func InitShare(router *gin.RouterGroup) {
	conf := service.Config()

	router.GET("/:token", func(c *gin.Context) {
		shareHandler(c, conf)
	})

	router.GET("/:token/*uid", func(c *gin.Context) {
		shareHandler(c, conf)
	})
}
