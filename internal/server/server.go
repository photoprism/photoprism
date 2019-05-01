package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
	log "github.com/sirupsen/logrus"
)

// Start the REST API server using the configuration provided
func Start(conf photoprism.Config) {
	if conf.HttpServerMode() != "" {
		gin.SetMode(conf.HttpServerMode())
	} else if conf.Debug() == false {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.Default()

	// Set template directory
	app.LoadHTMLGlob(conf.HttpTemplatesPath() + "/*")

	registerRoutes(app, conf)

	if err := app.Run(fmt.Sprintf("%s:%d", conf.HttpServerHost(), conf.HttpServerPort())); err != nil {
		log.Error(err)
	}
}
