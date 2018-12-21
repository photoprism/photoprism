package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
)

// Start the REST API server using the configuration provided
func Start(conf photoprism.Config) {
	if conf.GetServerMode() != "" {
		gin.SetMode(conf.GetServerMode())
	} else if conf.Debug() == false {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.Default()

	// Set template directory
	app.LoadHTMLGlob(conf.GetTemplatesPath() + "/*")

	registerRoutes(app, conf)

	app.Run(fmt.Sprintf("%s:%d", conf.GetServerIP(), conf.GetServerPort()))
}
