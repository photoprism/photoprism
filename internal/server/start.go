package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/server/process"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
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
	} else if !conf.Debug() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create new router engine without standard middleware.
	router := gin.New()

	// Configure trusted proxy ranges and forwarded client IP headers.
	configureTrustedProxySettings(router, conf)

	// Set trusted platform client IP address header name?
	if trustedPlatform := conf.TrustedPlatform(); trustedPlatform != "" {
		router.TrustedPlatform = trustedPlatform

		// Enable support for HTTP/2 without TLS.
		router.UseH2C = true
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
		// Use a custom compression predicate for fast, targeted exclusions.
		router.Use(gzip.Gzip(
			gzip.DefaultCompression,
			gzip.WithCustomShouldCompressFn(NewGzipShouldCompressFn(conf)),
		))
		log.Infof("server: enabled gzip compression")
	}

	// Register security middleware.
	router.Use(Security(conf))

	// Create REST API router group.
	APIv1 = router.Group(conf.BaseUri(config.ApiUri), APIMiddleware(conf))

	// Initialize package extensions.
	Ext().Init(router, conf)

	// Find and load templates.
	router.LoadHTMLFiles(conf.TemplateFiles()...)

	// Register application routes.
	registerRoutes(router, conf)

	// Register standard health check endpoints to determine whether the server is running.
	isLive := func(c *gin.Context) {
		c.Header(header.CacheControl, header.CacheControlNoStore)
		c.Header(header.AccessControlAllowOrigin, header.Any)
		c.JSON(http.StatusOK, api.NewHealthResponse("ok"))
	}
	router.Any(conf.BaseUri("/livez"), isLive)
	router.Any(conf.BaseUri("/health"), isLive)
	router.Any(conf.BaseUri("/healthz"), isLive)

	// Register "/readyz" endpoint to check if the server has been successfully initialized.
	isReady := func(c *gin.Context) {
		c.Header(header.CacheControl, header.CacheControlNoStore)
		c.Header(header.AccessControlAllowOrigin, header.Any)
		if conf.IsReady() {
			c.JSON(http.StatusOK, api.NewHealthResponse("ok"))
		} else {
			c.JSON(http.StatusServiceUnavailable, api.NewHealthResponse("service unavailable"))
		}
	}
	router.Any(conf.BaseUri("/readyz"), isReady)

	var tlsErr error
	var tlsManager *autocert.Manager
	var server *http.Server

	// Listen on a Unix domain socket instead of a TCP port?
	if unixSocket := conf.HttpSocket(); unixSocket != nil {
		var listener net.Listener
		var unixAddr *net.UnixAddr
		var err error

		// Check if the Unix socket already exists and delete it if the force flag is set.
		if fs.SocketExists(unixSocket.Path) {
			if !txt.Bool(unixSocket.Query().Get("force")) {
				Fail("server: %s socket %s already exists", clean.Log(unixSocket.Scheme), clean.Log(unixSocket.Path))
				return
			} else if removeErr := os.Remove(unixSocket.Path); removeErr != nil { //nolint:gosec // unixSocket.Path is parsed/validated in config.HttpSocket().
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
			} else if modeErr := os.Chmod(unixSocket.Path, fs.ParseMode(mode, fs.ModeSocket)); modeErr != nil { //nolint:gosec // unixSocket.Path is parsed/validated in config.HttpSocket().
				log.Warnf(
					"server: failed to change permissions of %s socket %s (%s)",
					clean.Log(unixSocket.Scheme),
					clean.Log(unixSocket.Path),
					modeErr,
				)
			}

			// Listen on Unix socket, which should be automatically closed and removed after use:
			// https://pkg.go.dev/net#UnixListener.SetUnlinkOnClose.
			server = newHTTPServer(router, conf)
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

		server = newHTTPServer(router, conf)

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

		server = newHTTPServer(router, conf)

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
			server = newHTTPServer(router, conf)
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

// configureTrustedProxySettings configures trusted proxy ranges for client IP resolution.
func configureTrustedProxySettings(router *gin.Engine, conf *config.Config) {
	if router == nil || conf == nil {
		return
	}

	if trustedProxies := conf.TrustedProxies(); len(trustedProxies) > 0 {
		if err := router.SetTrustedProxies(trustedProxies); err != nil {
			log.Warnf("server: %s (trusted proxy), falling back to direct client IP", err)
			if fallbackErr := router.SetTrustedProxies(nil); fallbackErr != nil {
				log.Warnf("server: %s (trusted proxy fallback)", fallbackErr)
			}
		} else {
			router.RemoteIPHeaders = conf.ProxyClientHeaders()
		}
	} else if err := router.SetTrustedProxies(nil); err != nil {
		log.Warnf("server: %s", err)
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
		redirectSrv := newHTTPServer(m.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			redirect(w, req, conf)
		})), conf)
		redirectSrv.Addr = fmt.Sprintf("%s:%d", conf.HttpHost(), conf.HttpPort())

		return redirectSrv.ListenAndServe()
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

// redirect sends HTTP requests to the configured HTTPS site host.
func redirect(w http.ResponseWriter, req *http.Request, conf *config.Config) {
	target := canonicalRedirectTarget(req, conf)

	http.Redirect(w, req, target, httpsRedirect)
}

// canonicalRedirectTarget returns the HTTPS redirect target using the configured public site host.
func canonicalRedirectTarget(req *http.Request, conf *config.Config) string {
	targetHost := ""

	if req != nil {
		targetHost = req.Host
	}

	if conf != nil {
		if host := conf.SiteHost(); host != "" {
			targetHost = host
		}
	}

	return HTTPSRedirectTarget(req, targetHost)
}

// HTTPSRedirectTarget returns the HTTPS redirect target for the provided request and host.
func HTTPSRedirectTarget(req *http.Request, host string) string {
	target := &url.URL{
		Scheme:   "https",
		Host:     host,
		Path:     "/",
		RawPath:  "",
		RawQuery: "",
	}

	if req != nil && req.URL != nil {
		target.Path = req.URL.Path
		target.RawPath = req.URL.RawPath
		target.RawQuery = req.URL.RawQuery
	} else if req != nil && req.RequestURI != "" {
		if u, err := url.ParseRequestURI(req.RequestURI); err == nil {
			target.Path = u.Path
			target.RawPath = u.RawPath
			target.RawQuery = u.RawQuery
		}
	}

	return target.String()
}

// newHTTPServer creates an HTTP server with hardened header and idle settings.
func newHTTPServer(handler http.Handler, conf *config.Config) *http.Server {
	headerTimeout := config.DefaultHttpHeaderTimeout
	headerBytes := config.DefaultHttpHeaderBytes
	idleTimeout := config.DefaultHttpIdleTimeout

	if conf != nil {
		headerTimeout = conf.HttpHeaderTimeout()
		headerBytes = conf.HttpHeaderBytes()
		idleTimeout = conf.HttpIdleTimeout()
	}

	return &http.Server{
		ReadHeaderTimeout: headerTimeout,
		ReadTimeout:       0,
		WriteTimeout:      0,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    headerBytes,
		Handler:           handler,
	}
}
