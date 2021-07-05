package server

import (
	"context"
	"fmt"
	"net/http"

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

	// Set http server mode.
	if conf.HttpMode() != "" {
		gin.SetMode(conf.HttpMode())
	} else if conf.Debug() == false {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router and add routing middleware.
	router := gin.New()
	router.Use(Logger(), Recovery())

	// Enable http compression (if any).
	switch conf.HttpCompression() {
	case "gzip":
		log.Infof("http: enabling gzip compression")
		router.Use(gzip.Gzip(
			gzip.DefaultCompression,
			gzip.WithExcludedPaths([]string{
				conf.BaseUri(config.ApiUri + "/t"),
				conf.BaseUri(config.ApiUri + "/folders/t"),
				conf.BaseUri(config.ApiUri + "/zip"),
				conf.BaseUri(config.ApiUri + "/albums"),
				conf.BaseUri(config.ApiUri + "/labels"),
			})))
	}

	// Set template directory
	router.LoadHTMLGlob(conf.TemplatesPath() + "/*")

	registerRoutes(router, conf)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort()),
		Handler: router,
	}

	go func() {
		log.Infof("http: starting web server at %s", server.Addr)

		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Info("http: web server shutdown complete")
			} else {
				log.Errorf("http: web server closed unexpect: %s", err)
			}
		}
	}()

	<-ctx.Done()
	log.Info("http: shutting down web server")
	err := server.Close()
	if err != nil {
		log.Errorf("http: web server shutdown failed: %v", err)
	}
}
