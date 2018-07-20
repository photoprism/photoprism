package server

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
)

func Start(address string, port int) {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.Static("/assets", "./assets")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "PhotoPrism",
		})
	})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.Run(fmt.Sprintf("%s:%d", address, port))
}
