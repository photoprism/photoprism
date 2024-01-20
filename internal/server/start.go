package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
)

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

	// Set proxy addresses from which headers related to the client and protocol can be trusted
	if err := router.SetTrustedProxies(conf.TrustedProxies()); err != nil {
		log.Warnf("server: %s", err)
	}

	// Register recovery and logger middleware.
	router.Use(Recovery(), Logger())

	// If enabled, register compression middleware.
	switch conf.HttpCompression() {
	case "br", "brotli":
		log.Infof("server: brotli compression is currently not supported")
	case "gzip":
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
		log.Infof("server: enabled gzip compression")
	}

	// Register security middleware.
	router.Use(Security(conf))

	// Create REST API router group.
	APIv1 = router.Group(conf.BaseUri(config.ApiUri), Api(conf))

	// Initialize package extensions.
	Ext().Init(router, conf)

	// Find and load templates.
	router.LoadHTMLFiles(conf.TemplateFiles()...)

	// Register HTTP route handlers.
	registerRoutes(router, conf)

	var tlsErr error
	var tlsManager *autocert.Manager
	var server *http.Server

	// Start HTTP server.
	if unixSocket := conf.HttpSocket(); unixSocket != "" {
		var listener net.Listener
		var unixAddr *net.UnixAddr
		var err error

		if unixAddr, err = net.ResolveUnixAddr("unix", unixSocket); err != nil {
			log.Errorf("server: resolve unix address failed (%s)", err)
			return
		} else if listener, err = net.ListenUnix("unix", unixAddr); err != nil {
			log.Errorf("server: listen unix address failed (%s)", err)
			return
		} else {
			server = &http.Server{
				Addr:    unixSocket,
				Handler: router,
			}

			log.Infof("server: listening on %s [%s]", unixSocket, time.Since(start))

			go StartHttp(server, listener)
		}
	} else if tlsManager, tlsErr = AutoTLS(conf); tlsErr == nil {
		server = &http.Server{
			Addr:      fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort()),
			TLSConfig: tlsManager.TLSConfig(),
			Handler:   router,
		}
		log.Infof("server: starting in auto tls mode on %s [%s]", server.Addr, time.Since(start))
		go StartAutoTLS(server, tlsManager, conf)
	} else if publicCert, privateKey := conf.TLS(); unixSocket == "" && publicCert != "" && privateKey != "" {
		log.Infof("server: starting in tls mode")
		server = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort()),
			Handler: router,
		}
		log.Infof("server: listening on %s [%s]", server.Addr, time.Since(start))
		go StartTLS(server, publicCert, privateKey)
	} else {
		log.Infof("server: %s", tlsErr)

		socket := fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort())

		if listener, err := net.Listen("tcp", socket); err != nil {
			log.Errorf("server: %s", err)
			return
		} else {
			server = &http.Server{
				Addr:    socket,
				Handler: router,
			}

			log.Infof("server: listening on %s [%s]", socket, time.Since(start))

			go StartHttp(server, listener)
		}
	}

	// Graceful HTTP server shutdown.
	<-ctx.Done()
	log.Info("server: shutting down")
	err := server.Close()
	if err != nil {
		log.Errorf("server: shutdown failed (%s)", err)
	}
}

// StartHttp starts the Web server in http mode.
func StartHttp(s *http.Server, l net.Listener) {
	if err := s.Serve(l); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("server: shutdown complete")
		} else {
			log.Errorf("server: %s", err)
		}
	}
}

// StartTLS starts the Web server in https mode.
func StartTLS(s *http.Server, httpsCert, privateKey string) {
	if err := s.ListenAndServeTLS(httpsCert, privateKey); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("server: shutdown complete")
		} else {
			log.Errorf("server: %s", err)
		}
	}
}

// StartAutoTLS starts the Web server with auto tls enabled.
func StartAutoTLS(s *http.Server, m *autocert.Manager, conf *config.Config) {
	var g errgroup.Group

	g.Go(func() error {
		return http.ListenAndServe(fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort()), m.HTTPHandler(http.HandlerFunc(redirect)))
	})

	g.Go(func() error {
		return s.ListenAndServeTLS("", "")
	})

	if err := g.Wait(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("server: shutdown complete")
		} else {
			log.Errorf("server: %s", err)
		}
	}
}

func redirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.RequestURI

	http.Redirect(w, req, target, httpsRedirect)
}
