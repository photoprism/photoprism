package server

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"github.com/photoprism/photoprism"
	"strconv"
	"log"
)

func Start(address string, port int, conf *photoprism.Config) {
	router := gin.Default()

	router.LoadHTMLGlob("server/templates/*")

	router.StaticFile("/favicon.ico", "./server/assets/favicon.ico")
	router.StaticFile("/robots.txt", "./server/assets/robots.txt")

	// JSON-REST API Version 1
	v1 := router.Group("/api/v1")
	{
		v1.GET("/photos", func(c *gin.Context) {
			search := photoprism.NewQuery(conf.OriginalsPath, conf.GetDb())

			count, _ := strconv.Atoi(c.DefaultQuery("count", "50"))
			offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

			query := c.DefaultQuery("q", "")

			photos := search.FindPhotos(query, count, offset)

			log.Printf("Query: %s, Count: %d", query, count)

			c.Header("x-result-total", strconv.Itoa(len(photos)))
			c.Header("x-result-count", strconv.Itoa(count))
			c.Header("x-result-offset", strconv.Itoa(offset))

			c.JSON(http.StatusOK, photos)
		})

		// v1.OPTIONS()

		v1.GET("/files", func(c *gin.Context) {
			search := photoprism.NewQuery(conf.OriginalsPath, conf.GetDb())

			files := search.FindFiles(70, 0)

			c.JSON(http.StatusOK, files)
		})

		v1.GET("/files/:id/thumbnail", func(c *gin.Context) {
			id := c.Param("id")
			size, _ := strconv.Atoi(c.Query("size"))

			search := photoprism.NewQuery(conf.OriginalsPath, conf.GetDb())

			file := search.FindFile(id)

			mediaFile := photoprism.NewMediaFile(file.Filename)

			thumbnail, _ := mediaFile.GetThumbnail(conf.ThumbnailsPath, size)

			c.File(thumbnail.GetFilename())
		})

		v1.GET("/files/:id/square_thumbnail", func(c *gin.Context) {
			id := c.Param("id")
			size, _ := strconv.Atoi(c.Query("size"))

			search := photoprism.NewQuery(conf.OriginalsPath, conf.GetDb())

			file := search.FindFile(id)

			mediaFile := photoprism.NewMediaFile(file.Filename)

			thumbnail, _ := mediaFile.GetSquareThumbnail(conf.ThumbnailsPath, size)

			c.File(thumbnail.GetFilename())
		})

		v1.GET("/albums", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		v1.GET("/tags", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	router.Static("/assets", "./server/assets")

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "PhotoPrism",
			"debug": true,
		})
	})

	router.Run(fmt.Sprintf("%s:%d", address, port))
}
