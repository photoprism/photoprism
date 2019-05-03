package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/context"
	log "github.com/sirupsen/logrus"
)

// Start the REST API server using the configuration provided
func Start(ctx *context.Context) {
	if ctx.HttpServerMode() != "" {
		gin.SetMode(ctx.HttpServerMode())
	} else if ctx.Debug() == false {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.New()
	app.Use(gin.Logger(), gin.Recovery())

	// Set template directory
	app.LoadHTMLGlob(ctx.HttpTemplatesPath() + "/*")

	registerRoutes(app, ctx)

	if err := app.Run(fmt.Sprintf("%s:%d", ctx.HttpServerHost(), ctx.HttpServerPort())); err != nil {
		log.Error(err)
	}
}
