package server

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/header"
)

// registerWebAppRoutes adds routes for the web user interface.
func registerWebAppRoutes(router *gin.Engine, conf *config.Config) {
	// Serve user interface bootstrap template on all routes starting with "/library".
	ui := func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			api.AbortNotFound(c)
			return
		}

		// Set values for UI bootstrap template.
		values := gin.H{
			"signUp": config.SignUp,
			"config": conf.ClientPublic(),
		}

		// Render UI bootstrap template.
		c.HTML(http.StatusOK, conf.TemplateName(), values)
	}
	router.Any(conf.LibraryUri("/*path"), ui)

	// Serve the user interface manifest file.
	manifest := func(c *gin.Context) {
		c.Header(header.CacheControl, header.CacheControlNoStore)
		c.Header(header.ContentType, header.ContentTypeJsonUtf8)
		c.IndentedJSON(200, conf.AppManifest())
	}
	router.Any(conf.BaseUri("/manifest.json"), manifest)

	// Serve user interface service worker file.
	swWorker := func(c *gin.Context) {
		c.Header(header.CacheControl, header.CacheControlNoStore)
		c.File(filepath.Join(conf.BuildPath(), "sw.js"))
	}
	router.Any("/sw.js", swWorker)

	if swUri := conf.BaseUri("/sw.js"); swUri != "/sw.js" {
		router.Any(swUri, swWorker)
	}
}
