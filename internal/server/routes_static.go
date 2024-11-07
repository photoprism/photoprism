package server

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/header"
)

// registerStaticRoutes adds routes for serving static content and templates.
func registerStaticRoutes(router *gin.Engine, conf *config.Config) {
	// Redirects to the login page.
	login := func(c *gin.Context) {
		if conf.OIDCEnabled() && conf.OIDCRedirect() {
			c.Redirect(http.StatusTemporaryRedirect, conf.OIDCLoginUri())
		} else {
			c.Redirect(http.StatusTemporaryRedirect, conf.LoginUri())
		}
	}

	// Control how crawlers index the site by serving a "robots.txt" file in addition
	// to the "X-Robots-Tag" response header set in the Security middleware:
	// https://developers.google.com/search/docs/crawling-indexing/robots/create-robots-txt
	router.Any(conf.BaseUri("/robots.txt"), func(c *gin.Context) {
		if robotsTxt, _ := conf.RobotsTxt(); len(robotsTxt) == 0 {
			// Return error 404 if file cannot be read or is empty.
			c.Data(http.StatusNotFound, header.ContentTypeText, []byte{})
		} else {
			// Allow clients to cache the response for one day.
			c.Header(header.CacheControl, header.CacheControlMaxAge(header.DurationDay, true))
			c.Data(http.StatusOK, header.ContentTypeText, robotsTxt)
		}
	})

	router.Any(conf.BaseUri("/"), login)

	// Shows "Page Not found" error if no other handler is registered.
	router.NoRoute(api.AbortNotFound)

	// Serves static favicon.
	router.StaticFile(conf.BaseUri("/favicon.ico"), filepath.Join(conf.ImgPath(), "favicon.ico"))

	// Serves static assets like js, css and font files.
	if dir := conf.StaticPath(); dir != "" {
		group := router.Group(conf.BaseUri(config.StaticUri), Static(conf))
		{
			group.Static("", dir)
		}
	}

	// Serves custom static assets if folder exists.
	if dir := conf.CustomStaticPath(); dir != "" {
		group := router.Group(conf.BaseUri(config.CustomStaticUri), Static(conf))
		{
			group.Static("", dir)
		}
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
