package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
)

func registerRoutes(router *gin.Engine, conf *config.Config) {
	// Favicon
	router.StaticFile("/favicon.ico", conf.HttpFaviconsPath()+"/favicon.ico")

	// Static assets like js and css files
	router.Static("/static", conf.HttpStaticPath())

	// JSON-REST API Version 1
	v1 := router.Group("/api/v1")
	{
		api.CreateSession(v1, conf)
		api.DeleteSession(v1, conf)

		api.GetThumbnail(v1, conf)
		api.GetDownload(v1, conf)

		api.GetPhotos(v1, conf)
		api.LikePhoto(v1, conf)
		api.DislikePhoto(v1, conf)

		api.GetLabels(v1, conf)
		api.LikeLabel(v1, conf)
		api.DislikeLabel(v1, conf)
		api.LabelThumbnail(v1, conf)

		api.Upload(v1, conf)
		api.Import(v1, conf)
		api.Index(v1, conf)

		api.BatchPhotosDelete(v1, conf)
		api.BatchPhotosPrivate(v1, conf)
		api.BatchPhotosStory(v1, conf)

		api.GetAlbums(v1, conf)
		api.LikeAlbum(v1, conf)
		api.DislikeAlbum(v1, conf)
		api.AlbumThumbnail(v1, conf)
		api.CreateAlbum(v1, conf)

		api.GetSettings(v1, conf)
		api.SaveSettings(v1, conf)

		api.Websocket(v1, conf)
	}

	// Default HTML page (client-side routing implemented via Vue.js)
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{"clientConfig": conf.ClientConfig()})
	})
}
