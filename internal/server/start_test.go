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
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
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
