package server

import (
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/photoprism"
	"net/http"
)

func registerRoutes(app *gin.Engine, conf *photoprism.Config) {
	// Favicon images
	app.StaticFile("/favicon.ico", conf.GetFaviconsPath()+"/favicon.ico")
	app.StaticFile("/favicon.png", conf.GetFaviconsPath()+"/favicon.png")

	// Static assets like js and css files
	app.Static("/assets", conf.GetPublicPath())

	// JSON-REST API Version 1
	v1 := app.Group("/api/v1")
	{
		api.GetPhotos(v1, conf)
		api.GetThumbnail(v1, conf)
		api.LikePhoto(v1, conf)
		api.DislikePhoto(v1, conf)
	}

	// Default HTML page (client-side routing implemented via Vue.js)
	app.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", conf.GetClientConfig())
	})
}
