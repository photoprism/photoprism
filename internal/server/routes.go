package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
)

func registerRoutes(router *gin.Engine, conf *config.Config) {
	// Static favicon file.
	router.StaticFile("/favicon.ico", conf.FaviconsPath()+"/favicon.ico")

	// Other static assets like JS and CSS files.
	router.Static("/static", conf.StaticPath())

	// Rainbow page.
	router.GET("/rainbow", func(c *gin.Context) {
		clientConfig := conf.PublicClientConfig()
		c.HTML(http.StatusOK, "rainbow.tmpl", gin.H{"config": clientConfig})
	})

	// JSON-REST API Version 1
	v1 := router.Group("/api/v1")
	{
		api.GetStatus(v1, conf)
		api.GetConfig(v1, conf)

		api.CreateSession(v1, conf)
		api.DeleteSession(v1, conf)

		api.GetPreview(v1, conf)
		api.GetThumbnail(v1, conf)
		api.GetDownload(v1, conf)
		api.GetVideo(v1, conf)
		api.CreateZip(v1, conf)
		api.DownloadZip(v1, conf)

		api.GetGeo(v1, conf)
		api.GetPhoto(v1, conf)
		api.GetPhotoYaml(v1, conf)
		api.UpdatePhoto(v1, conf)
		api.GetPhotos(v1, conf)
		api.GetPhotoDownload(v1, conf)
		api.GetPhotoLinks(v1, conf)
		api.CreatePhotoLink(v1, conf)
		api.UpdatePhotoLink(v1, conf)
		api.DeletePhotoLink(v1, conf)
		api.ApprovePhoto(v1, conf)
		api.LikePhoto(v1, conf)
		api.DislikePhoto(v1, conf)
		api.AddPhotoLabel(v1, conf)
		api.RemovePhotoLabel(v1, conf)
		api.UpdatePhotoLabel(v1, conf)
		api.GetMomentsTime(v1, conf)
		api.GetFile(v1, conf)
		api.SetPhotoPrimary(v1, conf)

		api.GetLabels(v1, conf)
		api.UpdateLabel(v1, conf)
		api.GetLabelLinks(v1, conf)
		api.CreateLabelLink(v1, conf)
		api.UpdateLabelLink(v1, conf)
		api.DeleteLabelLink(v1, conf)
		api.LikeLabel(v1, conf)
		api.DislikeLabel(v1, conf)
		api.LabelThumbnail(v1, conf)

		api.GetFoldersOriginals(v1, conf)
		api.GetFoldersImport(v1, conf)

		api.Upload(v1, conf)
		api.StartImport(v1, conf)
		api.CancelImport(v1, conf)
		api.StartIndexing(v1, conf)
		api.CancelIndexing(v1, conf)

		api.BatchPhotosArchive(v1, conf)
		api.BatchPhotosRestore(v1, conf)
		api.BatchPhotosPrivate(v1, conf)
		api.BatchAlbumsDelete(v1, conf)
		api.BatchLabelsDelete(v1, conf)

		api.GetAlbum(v1, conf)
		api.CreateAlbum(v1, conf)
		api.UpdateAlbum(v1, conf)
		api.DeleteAlbum(v1, conf)
		api.DownloadAlbum(v1, conf)
		api.GetAlbums(v1, conf)
		api.GetAlbumLinks(v1, conf)
		api.CreateAlbumLink(v1, conf)
		api.UpdateAlbumLink(v1, conf)
		api.DeleteAlbumLink(v1, conf)
		api.LikeAlbum(v1, conf)
		api.DislikeAlbum(v1, conf)
		api.AlbumThumbnail(v1, conf)
		api.CloneAlbums(v1, conf)
		api.AddPhotosToAlbum(v1, conf)
		api.RemovePhotosFromAlbum(v1, conf)

		api.GetAccounts(v1, conf)
		api.GetAccount(v1, conf)
		api.GetAccountDirs(v1, conf)
		api.ShareWithAccount(v1, conf)
		api.CreateAccount(v1, conf)
		api.DeleteAccount(v1, conf)
		api.UpdateAccount(v1, conf)

		api.GetSettings(v1, conf)
		api.SaveSettings(v1, conf)

		api.GetSvg(v1)

		api.Websocket(v1, conf)
	}

	// WebDAV server for file management, sync and sharing.
	if conf.WebDAVPassword() != "" {
		log.Info("webdav: enabled, username: photoprism")

		WebDAV(conf.OriginalsPath(), router.Group("/originals", gin.BasicAuth(gin.Accounts{
			"photoprism": conf.WebDAVPassword(),
		})), conf)

		log.Info("webdav: /originals/ available")

		if conf.ReadOnly() {
			log.Info("webdav: /import/ not available in read-only mode")
		} else {
			WebDAV(conf.ImportPath(), router.Group("/import", gin.BasicAuth(gin.Accounts{
				"photoprism": conf.WebDAVPassword(),
			})), conf)

			log.Info("webdav: /import/ available")
		}
	} else {
		log.Info("webdav: disabled (no password set)")
	}

	// Default HTML page for client-side rendering and routing via VueJS.
	router.NoRoute(func(c *gin.Context) {
		clientConfig := conf.PublicClientConfig()
		c.HTML(http.StatusOK, conf.DefaultTemplate(), gin.H{"config": clientConfig})
	})
}
