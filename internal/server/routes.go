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

		api.GetPreview(v1, conf)
		api.GetThumbnail(v1, conf)
		api.GetDownload(v1, conf)
		api.CreateZip(v1, conf)
		api.DownloadZip(v1, conf)

		api.GetGeo(v1, conf)
		api.GetPhoto(v1, conf)
		api.UpdatePhoto(v1, conf)
		api.GetPhotos(v1, conf)
		api.GetPhotoDownload(v1, conf)
		api.LikePhoto(v1, conf)
		api.DislikePhoto(v1, conf)
		api.AddPhotoLabel(v1, conf)
		api.RemovePhotoLabel(v1, conf)
		api.GetMomentsTime(v1, conf)
		api.GetFile(v1, conf)

		api.GetLabels(v1, conf)
		api.UpdateLabel(v1, conf)
		api.LikeLabel(v1, conf)
		api.DislikeLabel(v1, conf)
		api.LabelThumbnail(v1, conf)

		api.Upload(v1, conf)
		api.StartImport(v1, conf)
		api.CancelImport(v1, conf)
		api.StartIndexing(v1, conf)
		api.CancelIndexing(v1, conf)

		api.BatchPhotosArchive(v1, conf)
		api.BatchPhotosRestore(v1, conf)
		api.BatchPhotosPrivate(v1, conf)
		api.BatchPhotosStory(v1, conf)
		api.BatchAlbumsDelete(v1, conf)
		api.BatchLabelsDelete(v1, conf)

		api.GetAlbum(v1, conf)
		api.CreateAlbum(v1, conf)
		api.UpdateAlbum(v1, conf)
		api.DeleteAlbum(v1, conf)
		api.DownloadAlbum(v1, conf)
		api.GetAlbums(v1, conf)
		api.LikeAlbum(v1, conf)
		api.DislikeAlbum(v1, conf)
		api.AlbumThumbnail(v1, conf)
		api.AddPhotosToAlbum(v1, conf)
		api.RemovePhotosFromAlbum(v1, conf)

		api.GetSettings(v1, conf)
		api.SaveSettings(v1, conf)

		api.GetSvg(v1)

		api.Websocket(v1, conf)
	}

	// WebDAV server for file management / sharing
	if conf.WebDAVPassword() != "" {
		log.Info("webdav: enabled, username: photoprism")

		WebDAV(conf.OriginalsPath(), router.Group("/originals", gin.BasicAuth(gin.Accounts{
			"photoprism": conf.WebDAVPassword(),
		})), conf)

		log.Info("webdav: /originals available")

		WebDAV(conf.ExportPath(), router.Group("/export", gin.BasicAuth(gin.Accounts{
			"photoprism": conf.WebDAVPassword(),
		})), conf)

		log.Info("webdav: /export available")

		if conf.ReadOnly() {
			log.Info("webdav: /import not available in read-only mode")
		} else {
			WebDAV(conf.ImportPath(), router.Group("/import", gin.BasicAuth(gin.Accounts{
				"photoprism": conf.WebDAVPassword(),
			})), conf)

			log.Info("webdav: /import available")
		}
	} else {
		log.Info("webdav: disabled (no password set)")
	}

	// Default HTML page (client-side routing implemented via Vue.js)
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{"clientConfig": conf.PublicClientConfig()})
	})
}
