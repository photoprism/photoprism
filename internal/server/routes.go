package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/context"
)

func registerRoutes(router *gin.Engine, ctx *context.Context) {
	// Favicon
	router.StaticFile("/favicon.ico", ctx.HttpFaviconsPath()+"/favicon.ico")

	// Static assets like js and css files
	router.Static("/assets", ctx.HttpPublicPath())

	// JSON-REST API Version 1
	v1 := router.Group("/api/v1")
	{
		api.GetPhotos(v1, ctx)
		api.GetThumbnail(v1, ctx)
		api.LikePhoto(v1, ctx)
		api.DislikePhoto(v1, ctx)
	}

	// Default HTML page (client-side routing implemented via Vue.js)
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", ctx.ClientConfig())
	})
}
