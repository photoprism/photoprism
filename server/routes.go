package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism"
	"github.com/photoprism/photoprism/forms"
	"net/http"
	"strconv"
)

func ConfigureRoutes(app *gin.Engine, conf *photoprism.Config) {
	serverAssetsPath := conf.ServerAssetsPath
	app.LoadHTMLGlob(serverAssetsPath + "/templates/*")

	app.StaticFile("/favicon.ico", serverAssetsPath + "/favicons/favicon.ico")
	app.StaticFile("/favicon.png", serverAssetsPath + "/favicons/favicon.png")

	app.Static("/assets", serverAssetsPath + "/public")

	// JSON-REST API Version 1
	v1 := app.Group("/api/v1")
	{
		v1.GET("/photos", func(c *gin.Context) {
			var form forms.PhotoSearchForm

			search := photoprism.NewSearch(conf.OriginalsPath, conf.GetDb())

			c.MustBindWith(&form, binding.Form)

			if photos, err := search.Photos(form); err == nil {
				c.Header("x-result-total", strconv.Itoa(len(photos)))
				c.Header("x-result-count", strconv.Itoa(form.Count))
				c.Header("x-result-offset", strconv.Itoa(form.Offset))

				c.JSON(http.StatusOK, photos)
			} else {
				c.AbortWithError(400, err)
			}
		})

		// v1.OPTIONS()

		v1.GET("/files", func(c *gin.Context) {
			search := photoprism.NewSearch(conf.OriginalsPath, conf.GetDb())

			files := search.FindFiles(70, 0)

			c.JSON(http.StatusOK, files)
		})

		v1.GET("/files/:id/thumbnail", func(c *gin.Context) {
			id := c.Param("id")
			size, _ := strconv.Atoi(c.Query("size"))

			search := photoprism.NewSearch(conf.OriginalsPath, conf.GetDb())

			file := search.FindFile(id)

			if mediaFile, err := photoprism.NewMediaFile(file.FileName); err == nil {
				thumbnail, _ := mediaFile.GetThumbnail(conf.ThumbnailsPath, size)

				c.File(thumbnail.GetFilename())
			} else {
				c.Data(404, "image/svg+xml", notFoundSvg)
			}
		})

		v1.GET("/files/:id/square_thumbnail", func(c *gin.Context) {
			id := c.Param("id")
			size, _ := strconv.Atoi(c.Query("size"))

			search := photoprism.NewSearch(conf.OriginalsPath, conf.GetDb())

			file := search.FindFile(id)

			if mediaFile, err := photoprism.NewMediaFile(file.FileName); err == nil {
				thumbnail, _ := mediaFile.GetSquareThumbnail(conf.ThumbnailsPath, size)

				c.File(thumbnail.GetFilename())
			} else {
				c.Data(404, "image/svg+xml", notFoundSvg)
			}
		})

		v1.GET("/albums", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		v1.GET("/tags", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	app.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", conf.GetClientConfig())
	})
}
