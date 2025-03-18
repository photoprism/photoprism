package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/server/process"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/http/header"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Start the REST API server using the configuration provided
func Start(ctx context.Context, conf *config.Config) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	start := time.Now()

	// Log the server process ID for troubleshooting purposes.
	log.Infof("server: started as pid %d", process.ID)

	// Set web server mode.
	if conf.HttpMode() != "" {
		gin.SetMode(conf.HttpMode())
	} else if conf.Debug() == false {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create new router engine without standard middleware.
	router := gin.New()

	// Set proxy addresses from which headers related to the client and protocol can be trusted
	if err := router.SetTrustedProxies(conf.TrustedProxies()); err != nil {
		log.Warnf("server: %s", err)
	}

	// Register panic recovery middleware.
	router.Use(Recovery())

	// Register logger middleware if debug mode is enabled.
	if conf.Debug() {
		router.Use(Logger())
	}

	// Register compression middleware if enabled in the configuration.
	switch conf.HttpCompression() {
	case "br", "brotli":
		log.Infof("server: brotli compression is currently not supported")
	case "gzip":
		router.Use(gzip.Gzip(
			gzip.DefaultCompression,
			gzip.WithExcludedExtensions([]string{
				".png", ".gif", ".jpeg", ".jpg", ".webp", ".mp3", ".mp4", ".zip", ".gz",
			}),
			gzip.WithExcludedPaths([]string{
				conf.BaseUri("/health"),
				conf.BaseUri(config.ApiUri + "/t"),
				conf.BaseUri(config.ApiUri + "/folders/t"),
				conf.BaseUri(config.ApiUri + "/zip"),
				conf.BaseUri(config.ApiUri + "/albums"),
				conf.BaseUri(config.ApiUri + "/labels"),
				conf.BaseUri(config.ApiUri + "/videos"),
			}),
		))
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

	// Register application routes.
	registerRoutes(router, conf)

	// Register "GET /health" route so clients can perform health checks.
	router.GET(conf.BaseUri("/health"), func(c *gin.Context) {
		c.Header(header.CacheControl, header.CacheControlNoStore)
		c.Header(header.AccessControlAllowOrigin, header.Any)
		c.String(http.StatusOK, "OK")
	})

	// Create a new HTTP server instance with no read or write timeout, except for reading the headers:
	// https://pkg.go.dev/net/http#Server
	server := &http.Server{
		ReadHeaderTimeout: time.Minute,
		ReadTimeout:       -1,
		WriteTimeout:      -1,
		Handler:           router,
	}

	var tlsErr error
	var tlsManager *autocert.Manager

	// Listen on a Unix domain socket instead of a TCP port?
	if unixSocket := conf.HttpSocket(); unixSocket != nil {
		var listener net.Listener
		var unixAddr *net.UnixAddr
		var err error

		// Check if the Unix socket already exists and delete it if the force flag is set.
		if fs.SocketExists(unixSocket.Path) {
			if txt.Bool(unixSocket.Query().Get("force")) == false {
				Fail("server: %s socket %s already exists", clean.Log(unixSocket.Scheme), clean.Log(unixSocket.Path))
				return
			} else if removeErr := os.Remove(unixSocket.Path); removeErr != nil {
				Fail("server: %s socket %s already exists and cannot be deleted", clean.Log(unixSocket.Scheme), clean.Log(unixSocket.Path))
				return
			}
		}

		// Create a Unix socket and listen on it.
		if unixAddr, err = net.ResolveUnixAddr(unixSocket.Scheme, unixSocket.Path); err != nil {
			Fail("server: invalid %s socket (%s)", clean.Log(unixSocket.Scheme), err)
			return
		} else if listener, err = net.ListenUnix(unixSocket.Scheme, unixAddr); err != nil {
			Fail("server: failed to listen on %s socket (%s)", clean.Log(unixSocket.Scheme), err)
			return
		} else {
			// Update socket permissions?
			if mode := unixSocket.Query().Get("mode"); mode == "" {
				// Skip, no socket mode was specified.
			} else if modeErr := os.Chmod(unixSocket.Path, fs.ParseMode(mode, fs.ModeSocket)); modeErr != nil {
				log.Warnf(
					"server: failed to change permissions of %s socket %s (%s)",
					clean.Log(unixSocket.Scheme),
					clean.Log(unixSocket.Path),
					modeErr,
				)
			}

			// Listen on Unix socket, which should be automatically closed and removed after use:
			// https://pkg.go.dev/net#UnixListener.SetUnlinkOnClose.
			server.Addr = listener.Addr().String()

			log.Infof("server: listening on %s [%s]", unixSocket.Path, time.Since(start))

			// Start Web server.
			go StartHttp(server, listener)
		}
	} else if tlsManager, tlsErr = AutoTLS(conf); tlsErr == nil {
		log.Infof("server: starting in auto tls mode")

		tlsSocket := fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort())
		tlsConfig := tlsManager.TLSConfig()
		tlsConfig.MinVersion = tls.VersionTLS12

		// Listen on HTTPS socket.
		server.Addr = tlsSocket
		server.TLSConfig = tlsConfig

		log.Infof("server: listening on %s [%s]", server.Addr, time.Since(start))

		// Start Web server.
		go StartAutoTLS(server, tlsManager, conf)
	} else if publicCert, privateKey := conf.TLS(); publicCert != "" && privateKey != "" {
		log.Infof("server: starting in tls mode")

		tlsSocket := fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort())
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		// Listen on HTTPS socket.
		server.Addr = tlsSocket
		server.TLSConfig = tlsConfig

		log.Infof("server: listening on %s [%s]", server.Addr, time.Since(start))

		// Start Web server.
		go StartTLS(server, publicCert, privateKey)
	} else {
		log.Infof("server: %s", tlsErr)

		tcpSocket := fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort())

		if listener, err := net.Listen("tcp", tcpSocket); err != nil {
			Fail("server: %s", err)
			return
		} else {
			// Listen on HTTP socket.
			server.Addr = tcpSocket

			log.Infof("server: listening on %s [%s]", server.Addr, time.Since(start))

			// Start Web server.
			go StartHttp(server, listener)
		}
	}

	// Graceful web server shutdown.
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
