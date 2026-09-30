package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// newProxyTestRouter creates a test router with trusted proxy settings applied.
func newProxyTestRouter(conf *config.Config) *gin.Engine {
	r := gin.New()
	configureTrustedProxySettings(r, conf)

	r.GET("/ip", func(c *gin.Context) {
		c.String(http.StatusOK, header.ClientIP(c))
	})

	return r
}

// requestClientIP performs a test request and returns the resolved client IP.
func requestClientIP(t *testing.T, router *gin.Engine, remoteAddr, forwardedFor string) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = remoteAddr

	if forwardedFor != "" {
		req.Header.Set(header.XForwardedFor, forwardedFor)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	return w.Body.String()
}

func TestConfigureTrustedProxySettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("UsesForwardedIPForTrustedProxy", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().TrustedProxies = []string{header.CidrDockerInternal}
		conf.Options().ProxyClientHeaders = []string{header.XForwardedFor}

		router := newProxyTestRouter(conf)
		ip := requestClientIP(t, router, "172.16.5.10:12345", "203.0.113.9")

		assert.Equal(t, "203.0.113.9", ip)
	})
	t.Run("DisablesProxyTrustWhenNoTrustedProxiesConfigured", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().TrustedProxies = nil
		conf.Options().ProxyClientHeaders = []string{header.XForwardedFor}

		router := newProxyTestRouter(conf)
		ip := requestClientIP(t, router, "198.51.100.10:12345", "10.0.0.123")

		assert.Equal(t, "198.51.100.10", ip)
	})
	t.Run("FallsBackToDirectIPWhenTrustedProxyIsInvalid", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().TrustedProxies = []string{"invalid"}
		conf.Options().ProxyClientHeaders = []string{header.XForwardedFor}

		router := newProxyTestRouter(conf)
		ip := requestClientIP(t, router, "198.51.100.11:12345", "10.0.0.124")

		assert.Equal(t, "198.51.100.11", ip)
	})
	t.Run("DefaultTrustsLoopbackProxy", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().ProxyClientHeaders = []string{header.XForwardedFor}

		for _, f := range config.Flags {
			if ssf, ok := f.Flag.(*cli.StringSliceFlag); ok && ssf.Name == "trusted-proxy" {
				conf.Options().TrustedProxies = ssf.Value.Value()
			}
		}

		require.NotEmpty(t, conf.Options().TrustedProxies)

		router := newProxyTestRouter(conf)
		assert.Equal(t, "203.0.113.9", requestClientIP(t, router, "127.0.0.1:12345", "203.0.113.9"))
		assert.Equal(t, "203.0.113.9", requestClientIP(t, router, "[::1]:12345", "203.0.113.9"))
		assert.Equal(t, "203.0.113.9", requestClientIP(t, router, "127.0.0.1:12345", "198.51.100.7, 203.0.113.9"))
		assert.Equal(t, "203.0.113.9", requestClientIP(t, router, "172.18.0.2:12345", "203.0.113.9"))
		assert.Equal(t, "198.51.100.10", requestClientIP(t, router, "198.51.100.10:12345", "203.0.113.9"))
	})
	t.Run("ResolvesForwardedForPlatformHeader", func(t *testing.T) {
		t.Cleanup(func() { header.SetTrustedPlatform("") })

		conf := config.NewConfig(config.CliTestContext())
		conf.Options().TrustedProxies = nil
		conf.Options().TrustedPlatform = header.XForwardedFor

		router := newProxyTestRouter(conf)
		assert.Empty(t, router.TrustedPlatform)

		req := httptest.NewRequest(http.MethodGet, "/ip", nil)
		req.RemoteAddr = "10.128.2.4:12345"
		req.Header.Add(header.XForwardedFor, "198.51.100.7")
		req.Header.Add(header.XForwardedFor, "203.0.113.5")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "203.0.113.5", w.Body.String())
	})
	t.Run("SetsOtherPlatformHeader", func(t *testing.T) {
		t.Cleanup(func() { header.SetTrustedPlatform("") })

		conf := config.NewConfig(config.CliTestContext())
		conf.Options().TrustedProxies = nil
		conf.Options().TrustedPlatform = gin.PlatformCloudflare

		router := newProxyTestRouter(conf)
		assert.Equal(t, gin.PlatformCloudflare, router.TrustedPlatform)

		req := httptest.NewRequest(http.MethodGet, "/ip", nil)
		req.RemoteAddr = "10.128.2.4:12345"
		req.Header.Set(gin.PlatformCloudflare, "203.0.113.9")
		req.Header.Set(header.XForwardedFor, "198.51.100.7")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "203.0.113.9", w.Body.String())
	})
}

func TestNewHTTPServer(t *testing.T) {
	t.Run("UsesConfiguredValues", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().HttpHeaderTimeout = 15 * time.Second
		conf.Options().HttpHeaderBytes = 2048
		conf.Options().HttpIdleTimeout = 2 * time.Minute

		server := newHTTPServer(http.NewServeMux(), conf)

		assert.Equal(t, 15*time.Second, server.ReadHeaderTimeout)
		assert.Equal(t, 0*time.Second, server.ReadTimeout)
		assert.Equal(t, 0*time.Second, server.WriteTimeout)
		assert.Equal(t, 2*time.Minute, server.IdleTimeout)
		assert.Equal(t, 2048, server.MaxHeaderBytes)
	})
	t.Run("UsesDefaultsWhenConfigIsNil", func(t *testing.T) {
		server := newHTTPServer(http.NewServeMux(), nil)

		assert.Equal(t, config.DefaultHttpHeaderTimeout, server.ReadHeaderTimeout)
		assert.Equal(t, 0*time.Second, server.ReadTimeout)
		assert.Equal(t, 0*time.Second, server.WriteTimeout)
		assert.Equal(t, config.DefaultHttpIdleTimeout, server.IdleTimeout)
		assert.Equal(t, config.DefaultHttpHeaderBytes, server.MaxHeaderBytes)
	})
}

func TestHTTPSRedirectTarget(t *testing.T) {
	t.Run("UsesHostAndKeepsPathAndQuery", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost:2342/library/login?next=%2Flibrary", nil)

		assert.Equal(t, "https://photos.example.com:7443/library/login?next=%2Flibrary", HTTPSRedirectTarget(req, "photos.example.com:7443"))
	})
	t.Run("KeepsRawPath", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost:2342/library/a%2Fb", nil)

		assert.Equal(t, "https://photos.example.com/library/a%2Fb", HTTPSRedirectTarget(req, "photos.example.com"))
	})
	t.Run("UsesRequestURIWithoutURL", func(t *testing.T) {
		req := &http.Request{RequestURI: "/library/albums?q=a%20b"}

		assert.Equal(t, "https://photos.example.com/library/albums?q=a%20b", HTTPSRedirectTarget(req, "photos.example.com"))
	})
	t.Run("InvalidRequestURI", func(t *testing.T) {
		req := &http.Request{RequestURI: "library"}

		assert.Equal(t, "https://photos.example.com/", HTTPSRedirectTarget(req, "photos.example.com"))
	})
	t.Run("NilRequest", func(t *testing.T) {
		assert.Equal(t, "https://photos.example.com/", HTTPSRedirectTarget(nil, "photos.example.com"))
	})
}

// newTestKeyPair returns a self-signed PEM certificate and key for host.
func newTestKeyPair(t *testing.T, host string) (certPEM, keyPEM []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: host},
		DNSNames:     []string{host},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
}

// startTestTLS serves s on a loopback listener and returns its address; the server is closed on cleanup.
func startTestTLS(t *testing.T, s *http.Server) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	done := make(chan struct{})

	go func() {
		StartTLS(s, l)
		close(done)
	}()

	t.Cleanup(func() {
		_ = s.Close()
		<-done
	})

	return l.Addr().String()
}

func TestNewAutoTLSServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().SiteUrl = "https://photos.example.com/"
		conf.Options().HttpHost = "2001:db8::1"
		conf.Options().HttpPort = 443
		conf.Options().HttpHeaderTimeout = 7 * time.Second
		conf.Options().HttpHeaderBytes = 64 * 1024
		conf.Options().HttpIdleTimeout = 90 * time.Second

		handler := http.NewServeMux()
		server := newAutoTLSServer(handler, conf, &autocert.Manager{HostPolicy: autocert.HostWhitelist(conf.SiteDomain())})

		assert.Equal(t, "[2001:db8::1]:443", server.Addr)
		assert.Same(t, handler, server.Handler)
		require.NotNil(t, server.TLSConfig)
		assert.Equal(t, uint16(tls.VersionTLS12), server.TLSConfig.MinVersion)
		assert.Contains(t, server.TLSConfig.NextProtos, acme.ALPNProto)
		assert.NotNil(t, server.TLSConfig.GetCertificate)
		assert.Equal(t, 7*time.Second, server.ReadHeaderTimeout)
		assert.Equal(t, 64*1024, server.MaxHeaderBytes)
		assert.Equal(t, 90*time.Second, server.IdleTimeout)
	})
	t.Run("ReportsFailedCertificate", func(t *testing.T) {
		hook := captureSystemLog(t)
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().SiteUrl = "https://photos.example.com/"

		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(conf.SiteDomain()),
			Cache:      autocert.DirCache(t.TempDir()),
			Client:     &acme.Client{DirectoryURL: "http://127.0.0.1:1/directory"},
		}

		server := newAutoTLSServer(http.NewServeMux(), conf, m)
		_, err := server.TLSConfig.GetCertificate(&tls.ClientHelloInfo{ServerName: "photos.example.com", SupportedProtos: []string{"h2", "http/1.1"}})
		assert.Error(t, err)
		assert.Len(t, hook.AllEntries(), 1)
	})
}

func TestNewTLSServer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().HttpHost = "2001:db8::1"
		conf.Options().HttpPort = 8443
		conf.Options().HttpHeaderTimeout = 7 * time.Second
		conf.Options().HttpHeaderBytes = 64 * 1024
		conf.Options().HttpIdleTimeout = 90 * time.Second

		certPEM, keyPEM := newTestKeyPair(t, "photos.example.com")
		keyPair, err := tls.X509KeyPair(certPEM, keyPEM)
		require.NoError(t, err)

		handler := http.NewServeMux()
		server := newTLSServer(handler, conf, keyPair)

		assert.Equal(t, "[2001:db8::1]:8443", server.Addr)
		assert.Same(t, handler, server.Handler)
		require.NotNil(t, server.TLSConfig)
		assert.Equal(t, uint16(tls.VersionTLS12), server.TLSConfig.MinVersion)
		require.Len(t, server.TLSConfig.Certificates, 1)
		assert.Equal(t, keyPair.Certificate, server.TLSConfig.Certificates[0].Certificate)
		assert.Equal(t, 7*time.Second, server.ReadHeaderTimeout)
		assert.Equal(t, 64*1024, server.MaxHeaderBytes)
		assert.Equal(t, 90*time.Second, server.IdleTimeout)
	})
}

func TestStartTLS(t *testing.T) {
	const host = "photos.example.com"

	t.Run("AutoTLSChallenge", func(t *testing.T) {
		certDir := t.TempDir()
		certPEM, keyPEM := newTestKeyPair(t, host)
		require.NoError(t, os.WriteFile(filepath.Join(certDir, host+"+token"), append(keyPEM, certPEM...), 0o600))

		conf := config.NewConfig(config.CliTestContext())
		conf.Options().SiteUrl = "https://" + host + "/"

		m := &autocert.Manager{
			HostPolicy: autocert.HostWhitelist(host),
			Cache:      autocert.DirCache(certDir),
			Client:     &acme.Client{DirectoryURL: "http://127.0.0.1:1/directory"},
		}
		addr := startTestTLS(t, newAutoTLSServer(http.NewServeMux(), conf, m))

		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, NextProtos: []string{acme.ALPNProto}, InsecureSkipVerify: true}) //nolint:gosec // G402: self-signed test certificate
		require.NoError(t, err)
		defer conn.Close()

		state := conn.ConnectionState()
		assert.Equal(t, acme.ALPNProto, state.NegotiatedProtocol)
		require.NotEmpty(t, state.PeerCertificates)
		assert.Equal(t, host, state.PeerCertificates[0].Subject.CommonName)

		resp, err := http.Get("http://" + addr + "/") //nolint:gosec // G107: loopback test server
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
	t.Run("Certificates", func(t *testing.T) {
		certPEM, keyPEM := newTestKeyPair(t, host)
		keyPair, err := tls.X509KeyPair(certPEM, keyPEM)
		require.NoError(t, err)

		conf := config.NewConfig(config.CliTestContext())
		addr := startTestTLS(t, newTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}), conf, keyPair))

		client := &http.Client{Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{ServerName: host, InsecureSkipVerify: true}, //nolint:gosec // G402: self-signed test certificate
			ForceAttemptHTTP2: true,
		}}
		resp, err := client.Get("https://" + addr + "/")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.Equal(t, 2, resp.ProtoMajor)

		_, err = tls.Dial("tcp", addr, &tls.Config{ServerName: host, InsecureSkipVerify: true, MaxVersion: tls.VersionTLS11}) //nolint:gosec // G402: self-signed test certificate
		assert.Error(t, err)
	})
	t.Run("ClosesListenerWithoutCertificate", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		done := make(chan struct{})

		go func() {
			StartTLS(newHTTPServer(http.NewServeMux(), nil), l)
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("StartTLS did not return")
		}

		_, err = net.DialTimeout("tcp", l.Addr().String(), time.Second)
		assert.Error(t, err)
	})
}

// routeSet returns the registered routes of a router as "METHOD path" strings.
func routeSet(router *gin.Engine) map[string]bool {
	result := make(map[string]bool)

	for _, r := range router.Routes() {
		result[r.Method+" "+r.Path] = true
	}

	return result
}

func TestNewRouter(t *testing.T) {
	mode := gin.Mode()
	apiV1 := APIv1
	t.Cleanup(func() {
		gin.SetMode(mode)
		APIv1 = apiV1
	})
	gin.SetMode(gin.TestMode)

	conf := config.TestConfig()

	var router *gin.Engine

	require.NotPanics(t, func() { router = newRouter(conf) })
	require.NotNil(t, router)

	t.Run("RegistersRoutes", func(t *testing.T) {
		routes := routeSet(router)
		apiUri := conf.BaseUri(config.ApiUri)

		assert.Greater(t, len(routes), 200)
		assert.True(t, routes[http.MethodPost+" "+apiUri+"/session"])
		assert.True(t, routes[http.MethodGet+" "+apiUri+"/session"])
		assert.True(t, routes[http.MethodGet+" "+conf.BaseUri("/livez")])
		assert.True(t, routes[http.MethodGet+" "+conf.BaseUri("/readyz")])
		assert.True(t, routes[http.MethodGet+" "+conf.BaseUri("/.well-known/openid-configuration")])
		assert.True(t, routes[http.MethodGet+" "+conf.BaseUri("/.well-known/jwks.json")])
	})
	t.Run("SetsApiGroup", func(t *testing.T) {
		require.NotNil(t, APIv1)
		assert.Equal(t, conf.BaseUri(config.ApiUri), APIv1.BasePath())
	})
	t.Run("ServesHealthCheck", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, conf.BaseUri("/livez"), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"ok"`)
	})
	t.Run("UnknownApiRoute", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, conf.BaseUri(config.ApiUri+"/zz-unknown-route"), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestRegisterHealthRoutes(t *testing.T) {
	// request sends a GET request to the router and returns the recorded response.
	request := func(router *gin.Engine, path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w
	}

	t.Run("Ready", func(t *testing.T) {
		conf := config.TestConfig()
		require.True(t, conf.IsReady())

		router := gin.New()
		registerHealthRoutes(router, conf)

		for _, path := range []string{"/livez", "/health", "/healthz", "/readyz"} {
			w := request(router, conf.BaseUri(path))
			assert.Equal(t, http.StatusOK, w.Code, path)
			assert.Contains(t, w.Body.String(), `"status":"ok"`, path)
			assert.Equal(t, header.CacheControlNoStore, w.Header().Get(header.CacheControl), path)
			assert.Equal(t, header.Any, w.Header().Get(header.AccessControlAllowOrigin), path)
		}
	})
	t.Run("NotReady", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		require.False(t, conf.IsReady())

		router := gin.New()
		registerHealthRoutes(router, conf)

		w := request(router, conf.BaseUri("/readyz"))
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"service unavailable"`)
		assert.Equal(t, header.CacheControlNoStore, w.Header().Get(header.CacheControl))

		w = request(router, conf.BaseUri("/livez"))
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestListenUnixSocket(t *testing.T) {
	// socketUrl returns a Unix socket URL for a path with the specified query options.
	socketUrl := func(path, query string) *url.URL {
		return &url.URL{Scheme: "unix", Path: path, RawQuery: query}
	}

	// staleSocket creates a socket file that no process is listening on anymore.
	staleSocket := func(t *testing.T, path string) {
		t.Helper()

		l, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
		require.NoError(t, err)
		l.SetUnlinkOnClose(false)
		require.NoError(t, l.Close())
		require.True(t, fs.SocketExists(path))
	}

	t.Run("Success", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "server.sock")

		l, err := listenUnixSocket(socketUrl(path, ""))
		require.NoError(t, err)
		t.Cleanup(func() { _ = l.Close() })

		assert.True(t, fs.SocketExists(path))
		assert.Equal(t, path, l.Addr().String())
	})
	t.Run("Mode", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "server.sock")

		l, err := listenUnixSocket(socketUrl(path, "mode=0600"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = l.Close() })

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})
	t.Run("ExistsWithoutForce", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "server.sock")
		staleSocket(t, path)

		l, err := listenUnixSocket(socketUrl(path, ""))
		require.Error(t, err)
		assert.Nil(t, l)
		assert.Contains(t, err.Error(), "already exists")
		assert.True(t, fs.SocketExists(path))
	})
	t.Run("ExistsWithForce", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "server.sock")
		staleSocket(t, path)

		l, err := listenUnixSocket(socketUrl(path, "force=true"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = l.Close() })

		conn, err := net.Dial("unix", path)
		require.NoError(t, err)
		assert.NoError(t, conn.Close())
	})
	t.Run("ListenFailed", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing", "server.sock")

		l, err := listenUnixSocket(socketUrl(path, ""))
		require.Error(t, err)
		assert.Nil(t, l)
		assert.Contains(t, err.Error(), "failed to listen on unix socket")
	})
}

func TestStartHttp(t *testing.T) {
	// serve runs StartHttp in the background and returns a channel that is closed when it returns.
	serve := func(s *http.Server, l net.Listener) chan struct{} {
		done := make(chan struct{})

		go func() {
			StartHttp(s, l)
			close(done)
		}()

		return done
	}

	// wait fails the test if StartHttp does not return in time.
	wait := func(t *testing.T, done chan struct{}) {
		t.Helper()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("StartHttp did not return")
		}
	}

	t.Run("ServesUntilClosed", func(t *testing.T) {
		hook := captureRecoveryLog(t, logrus.InfoLevel)

		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		s := newHTTPServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}), nil)
		done := serve(s, l)

		resp, err := http.Get("http://" + l.Addr().String() + "/") //nolint:gosec // G107: loopback test server
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		require.NoError(t, s.Close())
		wait(t, done)

		require.NotEmpty(t, hook.AllEntries())
		assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
		assert.Equal(t, "server: shutdown complete", hook.LastEntry().Message)
	})
	t.Run("ListenerClosed", func(t *testing.T) {
		hook := captureRecoveryLog(t, logrus.InfoLevel)

		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		require.NoError(t, l.Close())

		wait(t, serve(newHTTPServer(http.NewServeMux(), nil), l))

		require.NotEmpty(t, hook.AllEntries())
		assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	})
}
