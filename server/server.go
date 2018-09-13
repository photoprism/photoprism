package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism"
)

func Start(address string, port int, conf *photoprism.Config) {
	app := gin.Default()

	ConfigureRoutes(app, conf)

	app.Run(fmt.Sprintf("%s:%d", address, port))
}
