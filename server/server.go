package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism"
)

func Start(address string, port int, mode string, conf *photoprism.Config) {
	if mode != "" {
		gin.SetMode(mode)
	} else if conf.Debug == false{
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.Default()

	ConfigureRoutes(app, conf)

	app.Run(fmt.Sprintf("%s:%d", address, port))
}
