package server

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/header"
)

// registerPWARoutes adds routes for bootstrapping and configuring the progressive web app user interface.
func registerPWARoutes(router *gin.Engine, conf *config.Config) {
	// Loads Progressive Web App (PWA) on all routes beginning with "library".
	pwa := func(c *gin.Context) {
		values := gin.H{
			"signUp": gin.H{"message": config.MsgSponsor, "url": config.SignUpURL},
			"config": conf.ClientPublic(),
		}
		c.HTML(http.StatusOK, conf.TemplateName(), values)
	}
	router.Any(conf.BaseUri("/library/*path"), pwa)

	// Progressive Web App (PWA) Manifest.
	manifest := func(c *gin.Context) {
		c.Header(header.CacheControl, header.CacheControlNoStore)
		c.Header(header.ContentType, header.ContentTypeJson)
		c.IndentedJSON(200, conf.AppManifest())
	}
	router.Any(conf.BaseUri("/manifest.json"), manifest)

	// Progressive Web App (PWA) Service Worker.
	swWorker := func(c *gin.Context) {
		c.Header(header.CacheControl, header.CacheControlNoStore)
		c.File(filepath.Join(conf.BuildPath(), "sw.js"))
	}
	router.Any("/sw.js", swWorker)

	if swUri := conf.BaseUri("/sw.js"); swUri != "/sw.js" {
		router.Any(swUri, swWorker)
	}
}
