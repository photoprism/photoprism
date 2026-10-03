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
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/server/process"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/dns"
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

	// Create the router engine with middleware and routes.
	router := newRouter(conf)

	var tlsErr error
	var tlsManager *autocert.Manager
	var server *http.Server

	// Listen on a Unix domain socket instead of a TCP port?
	if unixSocket := conf.HttpSocket(); unixSocket != nil {
		listener, err := listenUnixSocket(unixSocket)

		if err != nil {
			Fail("server: %s", err)
			return
		}

		// Listen on Unix socket, which should be automatically closed and removed after use:
		// https://pkg.go.dev/net#UnixListener.SetUnlinkOnClose.
		server = newHTTPServer(router, conf)
		server.Addr = listener.Addr().String()

		log.Infof("server: listening on %s [%s]", unixSocket.Path, time.Since(start))

		// Start Web server.
		go StartHttp(server, listener)
	} else if tlsManager, tlsErr = AutoTLS(conf); tlsErr == nil {
		log.Infof("server: starting in auto tls mode")

		server = newAutoTLSServer(router, conf, tlsManager)

		if listener, err := net.Listen("tcp", server.Addr); err != nil {
			Fail("server: %s", err)
			return
		} else {
			log.Infof("server: listening on %s [%s]", server.Addr, time.Since(start))

			// Start Web server.
			go StartTLS(server, listener)
		}
	} else if publicCert, privateKey := conf.TLS(); publicCert != "" && privateKey != "" {
		log.Infof("server: starting in tls mode")

		keyPair, err := tls.LoadX509KeyPair(publicCert, privateKey)

		if err != nil {
			Fail("server: %s", err)
			return
		}

		server = newTLSServer(router, conf, keyPair)

		if listener, listenErr := net.Listen("tcp", server.Addr); listenErr != nil {
			Fail("server: %s", listenErr)
			return
		} else {
			log.Infof("server: listening on %s [%s]", server.Addr, time.Since(start))

			// Start Web server.
			go StartTLS(server, listener)
		}
	} else {
		log.Infof("server: %s", tlsErr)

		tcpSocket := dns.JoinHostPort(conf.HttpHost(), conf.HttpPort())

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

// newRouter creates the router engine and registers its middleware, extensions, and routes.
func newRouter(conf *config.Config) *gin.Engine {
	// Create new router engine without standard middleware.
	router := gin.New()

	// Configure trusted proxy ranges and forwarded client IP headers.
	configureTrustedProxySettings(router, conf)

	// Enable support for HTTP/2 without TLS if a trusted platform header is set.
	if conf.TrustedPlatform() != "" {
		router.UseH2C = true
	}

	// Register panic recovery middleware.
	router.Use(Recovery())

	// Register logger middleware if debug mode is enabled.
	if conf.Debug() {
		router.Use(Logger())
	}

	// Warn once per unsupported compression token so operator typos are visible.
	for _, token := range conf.HttpCompressionUnknown() {
		log.Warnf("server: ignored unsupported http-compression value %q", token)
	}

	// Register compression middleware if enabled in the configuration.
	if prefs := conf.HttpCompressionPreferences(); len(prefs) > 0 {
		router.Use(NewCompressMiddleware(conf))
		log.Infof("server: enabled http compression (%s)", strings.Join(prefs, " > "))
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

	// Register health check endpoints.
	registerHealthRoutes(router, conf)

	return router
}

// registerHealthRoutes registers the endpoints that report whether the server is running and ready.
func registerHealthRoutes(router *gin.Engine, conf *config.Config) {
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
}

// listenUnixSocket listens on the Unix domain socket specified in the config and applies its
// "force" and "mode" query options, replacing an existing socket only when force is set.
func listenUnixSocket(unixSocket *url.URL) (net.Listener, error) {
	// Check if the Unix socket already exists and delete it if the force flag is set.
	if fs.SocketExists(unixSocket.Path) {
		if !txt.Bool(unixSocket.Query().Get("force")) {
			return nil, fmt.Errorf("%s socket %s already exists", clean.Log(unixSocket.Scheme), clean.Log(unixSocket.Path))
		} else if removeErr := os.Remove(unixSocket.Path); removeErr != nil { //nolint:gosec // unixSocket.Path is parsed/validated in config.HttpSocket().
			return nil, fmt.Errorf("%s socket %s already exists and cannot be deleted", clean.Log(unixSocket.Scheme), clean.Log(unixSocket.Path))
		}
	}

	// Create a Unix socket and listen on it.
	unixAddr, err := net.ResolveUnixAddr(unixSocket.Scheme, unixSocket.Path)

	if err != nil {
		return nil, fmt.Errorf("invalid %s socket (%s)", clean.Log(unixSocket.Scheme), err)
	}

	listener, err := net.ListenUnix(unixSocket.Scheme, unixAddr)

	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s socket (%s)", clean.Log(unixSocket.Scheme), err)
	}

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

	return listener, nil
}

// configureTrustedProxySettings configures trusted proxy ranges and the trusted platform header
// for client IP resolution.
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

	router.TrustedPlatform = header.SetTrustedPlatform(conf.TrustedPlatform())
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

// StartTLS starts the Web server in https mode with the certificates provided by s.TLSConfig.
// The listener is closed on return, including when the TLS setup fails before serving.
func StartTLS(s *http.Server, l net.Listener) {
	defer l.Close()

	if err := s.ServeTLS(l, "", ""); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("server: shutdown complete")
		} else {
			log.Errorf("server: %s", err)
		}
	}
}

// newTLSServer creates the HTTPS server for tls mode with the specified certificate.
func newTLSServer(handler http.Handler, conf *config.Config, keyPair tls.Certificate) *http.Server {
	server := newHTTPServer(handler, conf)
	server.Addr = dns.JoinHostPort(conf.HttpHost(), conf.HttpPort())
	server.TLSConfig = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{keyPair},
	}

	return server
}

// newAutoTLSServer creates the HTTPS server for auto tls mode.
// Certificates are obtained with the tls-alpn-01 challenge on the same port.
func newAutoTLSServer(handler http.Handler, conf *config.Config, m *autocert.Manager) *http.Server {
	tlsConfig := m.TLSConfig()
	tlsConfig.MinVersion = tls.VersionTLS12
	tlsConfig.GetCertificate = certificateWarner(tlsConfig.GetCertificate, conf.SiteDomain(), certificateWarnInterval)

	server := newHTTPServer(handler, conf)
	server.Addr = dns.JoinHostPort(conf.HttpHost(), conf.HttpPort())
	server.TLSConfig = tlsConfig

	return server
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
