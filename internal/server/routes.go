package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/pjebs/restgate"
)

func newAuthHandler(key, secret string) gin.HandlerFunc {
	rg := restgate.New(
		"X-Auth-Key",    // key header name
		"X-Auth-Secret", // secret header name
		restgate.Static,
		restgate.Config{
			Key:                []string{key},
			Secret:             []string{secret},
			HTTPSProtectionOff: true,
		},
	)
	return func(c *gin.Context) {
		nextCalled := false
		nextAdapter := func(http.ResponseWriter, *http.Request) {
			nextCalled = true
			c.Next()
		}
		rg.ServeHTTP(c.Writer, c.Request, nextAdapter)
		if nextCalled == false {
			c.AbortWithStatus(401)
		}
	}
}

func registerRoutes(router *gin.Engine, conf *config.Config) {
	// Favicon
	router.StaticFile("/favicon.ico", conf.HttpFaviconsPath()+"/favicon.ico")

	// Static assets like js and css files
	router.Static("/static", conf.HttpStaticPath())

	// JSON-REST API Version 1
	v1 := router.Group("/api/v1")
	v1NoAuth := v1.Group("")
	{
		api.GetThumbnail(v1NoAuth, conf)
		api.GetDownload(v1NoAuth, conf)
	}
	v1Auth := v1.Group("")
	if pass := conf.HttpServerPassword(); pass != "" {
		v1Auth.Use(newAuthHandler("default", pass))
	}
	{
		api.GetPhotos(v1Auth, conf)
		api.LikePhoto(v1Auth, conf)
		api.DislikePhoto(v1Auth, conf)

		api.GetLabels(v1Auth, conf)
		api.LikeLabel(v1Auth, conf)
		api.DislikeLabel(v1Auth, conf)
		api.LabelThumbnail(v1Auth, conf)

		api.Upload(v1Auth, conf)
		api.Import(v1Auth, conf)
		api.Index(v1Auth, conf)

		api.BatchPhotosDelete(v1Auth, conf)
		api.BatchPhotosPrivate(v1Auth, conf)
		api.BatchPhotosStory(v1Auth, conf)

		api.GetAlbums(v1Auth, conf)
		api.LikeAlbum(v1Auth, conf)
		api.DislikeAlbum(v1Auth, conf)
		api.AlbumThumbnail(v1Auth, conf)
		api.CreateAlbum(v1Auth, conf)

		api.Ping(v1Auth, conf)
	}

	// Default HTML page (client-side routing implemented via Vue.js)
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", conf.ClientConfig())
	})
}
