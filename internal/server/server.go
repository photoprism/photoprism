package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

// Start the REST API server using the configuration provided
func Start(ctx context.Context, conf *config.Config) {
	if conf.HttpServerMode() != "" {
		gin.SetMode(conf.HttpServerMode())
	} else if conf.Debug() == false {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Set template directory
	router.LoadHTMLGlob(conf.HttpTemplatesPath() + "/*")

	registerRoutes(router, conf)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.HttpServerHost(), conf.HttpServerPort()),
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Info("web server shutdown complete")
			} else {
				log.Errorf("web server closed unexpect: %s", err)
			}
		}
	}()

	<-ctx.Done()
	log.Info("shutting down web server")
	err := server.Close()
	if err != nil {
		log.Errorf("web server shutdown failed: %v", err)
	}
}
