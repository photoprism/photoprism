package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

// Start the REST API server using the configuration provided
func Start(ctx context.Context, conf *config.Config) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	start := time.Now()

	// Set HTTP server mode.
	if conf.HttpMode() != "" {
		gin.SetMode(conf.HttpMode())
	} else if conf.Debug() == false {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create new HTTP router engine without standard middleware.
	router := gin.New()

	// Register common middleware.
	router.Use(Logger(), Recovery(), Security(conf))

	// Initialize package extensions.
	Ext().Init(router, conf)

	// Enable HTTP compression?
	switch conf.HttpCompression() {
	case "gzip":
		log.Infof("server: enabling gzip compression")
		router.Use(gzip.Gzip(
			gzip.DefaultCompression,
			gzip.WithExcludedPaths([]string{
				conf.BaseUri(config.ApiUri + "/t"),
				conf.BaseUri(config.ApiUri + "/folders/t"),
				conf.BaseUri(config.ApiUri + "/zip"),
				conf.BaseUri(config.ApiUri + "/albums"),
				conf.BaseUri(config.ApiUri + "/labels"),
				conf.BaseUri(config.ApiUri + "/videos"),
			})))
	}

	// Find and load templates.
	router.LoadHTMLFiles(conf.TemplateFiles()...)

	// Register HTTP route handlers.
	registerRoutes(router, conf)

	// Create new HTTP server instance.
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort()),
		Handler: router,
	}

	// Start HTTP server.
	go func() {
		log.Infof("server: listening on %s [%s]", server.Addr, time.Since(start))

		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Info("server: shutdown complete")
			} else {
				log.Errorf("server: %s", err)
			}
		}
	}()

	// Graceful HTTP server shutdown.
	<-ctx.Done()
	log.Info("server: shutting down")
	err := server.Close()
	if err != nil {
		log.Errorf("server: shutdown failed (%s)", err)
	}
}
