package server

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

var MethodsGetHead = []string{http.MethodGet, http.MethodHead}

// registerWebAppRoutes adds routes for the web user interface.
func registerWebAppRoutes(router *gin.Engine, conf *config.Config) {
	// Return if the web user interface is disabled.
	if conf.DisableFrontend() {
		return
	}

	// Serve user interface bootstrap template on all routes starting with "/library".
	ui := func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			api.AbortNotFound(c)
			return
		}

		// Get client configuration.
		clientConfig := conf.ClientPublic()

		// Set bootstrap template values.
		values := gin.H{
			"signUp":    config.SignUp,
			"config":    clientConfig,
			"splashCss": clientConfig.ClientAssets.SplashCssFileContents(),
		}

		// Render bootstrap template.
		c.HTML(http.StatusOK, conf.TemplateName(), values)
	}

	// HTML bootstrap for the SPA (served from /library/**).
	router.Any(conf.LibraryUri("/*path"), ui)

	// Serve the user interface manifest file.
	manifest := func(c *gin.Context) {
		c.Header(header.CacheControl, header.CacheControlNoStore)
		c.Header(header.ContentType, header.ContentTypeJsonUtf8)
		c.IndentedJSON(200, conf.AppManifest())
	}

	// Web App Manifest (served at /manifest.json under the base URI).
	router.Any(conf.BaseUri("/"+fs.ManifestJsonFile), manifest)

	// Serve user interface service worker file.
	swWorker := func(c *gin.Context) {
		c.Header(header.CacheControl, header.CacheControlNoStore)

		// Return if only headers are requested.
		if c.Request.Method == http.MethodHead {
			c.Header(header.ContentType, header.ContentTypeJavaScript)
			return
		}

		// Serve the Workbox-generated service worker when the frontend build has
		// produced one (default for production builds).
		if swFile := conf.StaticBuildFile(fs.SwJsFile); fs.FileExistsNotEmpty(swFile) {
			c.File(swFile)
			return
		}

		// Fall back to the embedded no-op service worker so tests and dev builds
		// still receive a valid response.
		if len(fallbackServiceWorker) > 0 {
			c.Data(http.StatusOK, header.ContentTypeJavaScript, fallbackServiceWorker)
			return
		}

		api.Abort(c, http.StatusNotFound, i18n.ErrNotFound)
	}

	// Primary service worker endpoint (/sw.js relative to the site root).
	router.Match(MethodsGetHead, "/"+fs.SwJsFile, swWorker)

	// Expose hashed Workbox runtime helpers alongside sw.js so service worker imports succeed
	// regardless of whether the app is hosted at the root or under a base URI.
	workboxHandler := newWorkboxHandler(conf)

	// Handler for shared domain (service worker registered from /sw.js).
	router.Match(MethodsGetHead, "/workbox-:hash", workboxHandler)

	// Handle service worker requests on a shared domain.
	if conf.BaseUri("") != "" {
		router.Match(MethodsGetHead, conf.BaseUri("/"+fs.SwJsFile), swWorker)
		router.Match(MethodsGetHead, conf.BaseUri("/workbox-:hash"), workboxHandler)
	}
}

// newWorkboxHandler serves hashed workbox helpers (workbox-<hash>.js). The regex
// matches the raw filename (without the "workbox-" prefix) as seen by Gin, so
// the pattern must be `^[A-Za-z0-9_-]+\.js$`. Note the single backslash – the
// string is a raw literal, meaning the regex engine receives an escaped dot.
func newWorkboxHandler(conf *config.Config) gin.HandlerFunc {
	workboxPattern := regexp.MustCompile(`^[A-Za-z0-9_-]+\.js$`)

	return func(c *gin.Context) {
		raw := c.Param("hash")
		if !workboxPattern.MatchString(raw) {
			c.Status(http.StatusNotFound)
			return
		}

		filePath := conf.StaticBuildFile("workbox-" + raw)
		if !fs.FileExists(filePath) {
			c.Status(http.StatusNotFound)
			return
		}

		// Return if only headers are requested.
		if c.Request.Method == http.MethodHead {
			c.Header(header.ContentType, header.ContentTypeJavaScript)
			return
		}

		c.File(filePath)
	}
}
