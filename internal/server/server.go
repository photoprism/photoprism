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

	// Register logger middleware.
	router.Use(Logger(), Recovery())

	// Register security middleware.
	router.Use(Security(SecurityOptions{
		IsDevelopment:         gin.Mode() != gin.ReleaseMode || conf.Test(),
		AllowedHosts:          []string{},
		SSLRedirect:           false,
		SSLHost:               "",
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:            0,
		STSIncludeSubdomains:  false,
		FrameDeny:             true,
		ContentTypeNosniff:    false,
		BrowserXssFilter:      false,
		ContentSecurityPolicy: "frame-ancestors 'none';",
	}))

	// Enable HTTP compression?
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
				conf.BaseUri(config.ApiUri + "/videos"),
			})))
	}

	// Set template directory
	router.LoadHTMLGlob(conf.TemplatesPath() + "/*")

	// Register HTTP route handlers.
	registerRoutes(router, conf)

	// Create new HTTP server instance.
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort()),
		Handler: router,
	}

	log.Debugf("http: successfully initialized [%s]", time.Since(start))

	// Start HTTP server.
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

	// Graceful HTTP server shutdown.
	<-ctx.Done()
	log.Info("http: shutting down web server")
	err := server.Close()
	if err != nil {
		log.Errorf("http: web server shutdown failed: %v", err)
	}
}
