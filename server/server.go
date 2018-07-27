package server

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
)

func Start(address string, port int) {
	router := gin.Default()

	router.LoadHTMLGlob("server/templates/*")

	router.StaticFile("/favicon.ico", "./server/assets/favicon.ico")
	router.StaticFile("/robots.txt", "./server/assets/robots.txt")

	router.Static("/assets", "./server/assets")

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "PhotoPrism",
		})
	})

	router.Run(fmt.Sprintf("%s:%d", address, port))
}
