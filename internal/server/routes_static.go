package server

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	httpsec "github.com/photoprism/photoprism/pkg/http/security"
	"github.com/photoprism/photoprism/pkg/i18n"
)

const (
	// IndexHtml is the default frontend entrypoint file name.
	IndexHtml = "index.html"
)

// registerStaticRoutes adds routes for serving static content and templates.
func registerStaticRoutes(router *gin.Engine, conf *config.Config) {
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

	// Return if the web user interface is disabled.
	if conf.DisableFrontend() {
		log.Info("frontend: disabled")
		router.NoRoute(func(c *gin.Context) {
			api.Abort(c, http.StatusNotFound, i18n.ErrNotFound)
		})
		return
	}

	// Redirects to the login page.
	login := func(c *gin.Context) {
		if conf.OIDCEnabled() && conf.OIDCRedirect() {
			c.Redirect(http.StatusTemporaryRedirect, conf.OIDCLoginUri())
		} else {
			c.Redirect(http.StatusTemporaryRedirect, conf.LoginUri())
		}
	}

	webBase := clean.SlashPath(conf.BasePath())

	// Attach the web request handler to serve assets only if a web storage directory exists.
	if webDir := conf.WebStoragePath(); webDir != "." && webDir != "/" && fs.PathExists(webDir) {
		// Serves static web assets from the web storage path.
		web := func(c *gin.Context) {
			// Gin always provides a valid request context here; only handle GET/HEAD.
			if c.Request.Method != header.MethodGet && c.Request.Method != header.MethodHead {
				return
			}

			requestPath := c.Request.URL.Path
			escapedPath := c.Request.URL.EscapedPath()

			// Reject ambiguous path variants that could bypass expected access checks.
			if httpsec.OverlayHasAmbiguousPath(requestPath, escapedPath) {
				return
			}

			// Resolve request path to an overlay-relative path (if in scope).
			webPath, ok := httpsec.OverlayRelativePath(requestPath, webBase)
			if !ok {
				return
			}

			if webPath == "" {
				webPath = IndexHtml
			} else if path.Ext(webPath) == "" {
				webPath = path.Join(webPath, IndexHtml)
			}

			// Block hidden/special paths and sensitive file names from direct access.
			if httpsec.OverlayPathBlocked(webPath) {
				log.Tracef("web: blocked overlay path %s", clean.Log(webPath))
				return
			}

			webFile, hasFile := httpsec.OverlayResolveFile(webDir, webPath)

			// Resolve unmatched requests by serving an overlay file, redirecting root,
			// or falling through so the next NoRoute handler can return 404.
			switch {
			case hasFile:
				// Serve the matched overlay file and stop the NoRoute handler chain.
				log.Debugf("web: serving %s", clean.Log(webFile))
				c.Abort()
				c.File(webFile)
			case webPath == IndexHtml:
				// No root index in overlay: keep the default login/landing redirect behavior.
				c.Abort()
				login(c)
			default:
				// Intentionally do not abort so api.AbortNotFound handles the request next.
				log.Tracef("web: no asset found for request path %s", clean.Log(webPath))
			}
		}

		// Serve overlay assets when available and otherwise fall through to Not Found.
		router.NoRoute(web, api.AbortNotFound)
	} else {
		// Redirect to default login/landing page.
		router.Match(MethodsGetHead, conf.BaseUri("/"), login)

		// Render error 404 Not Found.
		router.NoRoute(api.AbortNotFound)
	}

	// Serves static favicon.
	router.StaticFile(conf.BaseUri("/favicon.ico"), conf.SiteFavicon())

	// Serves bundled assets like JS, CSS, and fonts.
	if dir := conf.StaticPath(); dir != "" {
		group := router.Group(conf.BaseUri(config.StaticUri), Static(conf))
		{
			group.Static("", dir)
		}
	}

	// Serves custom assets, e.g. bundled with extensions.
	if dir := conf.CustomStaticPath(); dir != "" {
		group := router.Group(conf.BaseUri(config.CustomStaticUri), Static(conf))
		{
			group.Static("", dir)
		}
	}

	// Serves rainbow test page.
	router.GET(conf.BaseUri("/_rainbow"), func(c *gin.Context) {
		clientConfig := conf.ClientPublic()
		c.HTML(http.StatusOK, "rainbow.gohtml", gin.H{"config": clientConfig})
	})

	// Serves splash screen test page.
	router.GET(conf.BaseUri("/_splash"), func(c *gin.Context) {
		clientConfig := conf.ClientPublic()
		c.HTML(http.StatusOK, "splash.gohtml", gin.H{"config": clientConfig})
	})
}
