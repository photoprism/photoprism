package server

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"github.com/photoprism/photoprism"
)

func Start(address string, port int) {
	router := gin.Default()

	router.LoadHTMLGlob("server/templates/*")

	router.StaticFile("/favicon.ico", "./server/assets/favicon.ico")
	router.StaticFile("/robots.txt", "./server/assets/robots.txt")

	// JSON-REST API Version 1
	v1 := router.Group("/v1")
	{
		v1.GET("/photos", func(c *gin.Context) {

			photos := []*photoprism.Photo{ &photoprism.Photo{Title: "Bar"}, &photoprism.Photo{ Title: "Foo" } }

			c.JSON(http.StatusOK, photos)
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
