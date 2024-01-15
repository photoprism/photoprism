package server

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/i18n"
)

// registerStaticRoutes adds routes for serving static content and templates.
func registerStaticRoutes(router *gin.Engine, conf *config.Config) {
	// Redirects to the PWA for now, can be replaced by a template later.
	login := func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, conf.LoginUri())
	}
	router.Any(conf.BaseUri("/"), login)

	// Shows "Page Not found" error if no other handler is registered.
	router.NoRoute(func(c *gin.Context) {
		switch c.NegotiateFormat(gin.MIMEHTML, gin.MIMEJSON) {
		case gin.MIMEJSON:
			c.JSON(http.StatusNotFound, gin.H{"error": i18n.Msg(i18n.ErrNotFound)})
		default:
			values := gin.H{
				"signUp": gin.H{"message": config.MsgSponsor, "url": config.SignUpURL},
				"config": conf.ClientPublic(),
				"error":  i18n.Msg(i18n.ErrNotFound),
				"code":   http.StatusNotFound,
			}
			c.HTML(http.StatusNotFound, "404.gohtml", values)
		}
	})

	// Serves static favicon.
	router.StaticFile(conf.BaseUri("/favicon.ico"), filepath.Join(conf.ImgPath(), "favicon.ico"))

	// Serves static assets like js, css and font files.
	router.Static(conf.BaseUri(config.StaticUri), conf.StaticPath())

	// Serves custom static assets if folder exists.
	if dir := conf.CustomStaticPath(); dir != "" {
		router.Static(conf.BaseUri(config.CustomStaticUri), dir)
	}

	// Rainbow Page.
	router.GET(conf.BaseUri("/_rainbow"), func(c *gin.Context) {
		clientConfig := conf.ClientPublic()
		c.HTML(http.StatusOK, "rainbow.gohtml", gin.H{"config": clientConfig})
	})

	// Splash Screen.
	router.GET(conf.BaseUri("/_splash"), func(c *gin.Context) {
		clientConfig := conf.ClientPublic()
		c.HTML(http.StatusOK, "splash.gohtml", gin.H{"config": clientConfig})
	})
}
