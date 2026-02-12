package node

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newBootstrapTestConfig creates a minimal test config and closes its database on test cleanup.
func newBootstrapTestConfig(t *testing.T, name string) *config.Config {
	t.Helper()
	c := config.NewMinimalTestConfigWithDb(name, t.TempDir())
	t.Cleanup(func() {
		assert.NoError(t, c.CloseDb())
	})

	return c
}

func TestInitConfig_NoPortal_NoOp(t *testing.T) {
	c := newBootstrapTestConfig(t, "bootstrap")

	// Default NodeRole() resolves to app; no Portal configured.
	assert.Equal(t, cluster.RoleApp, c.NodeRole())
	assert.NoError(t, InitConfig(c))
}

func TestInitConfig_ServiceRole(t *testing.T) {
	c := newBootstrapTestConfig(t, "bootstrap-service")

	c.Options().NodeRole = cluster.RoleService

	assert.NoError(t, InitConfig(c))
}

func TestRegister_PersistSecretAndDB(t *testing.T) {
	// Fake Portal server.
	var jwksURL string
	expectedSite := "https://public.example.test/"
	var expectedAppName string
	var expectedAppVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cluster/nodes/register":
			var req cluster.RegisterRequest
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, expectedSite, req.SiteUrl)
			assert.Equal(t, expectedAppName, req.AppName)
			assert.Equal(t, expectedAppVersion, req.AppVersion)
			// Minimal successful registration with secrets + DSN.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			resp := cluster.RegisterResponse{
				Node:        cluster.Node{Name: "pp-node-01"},
				UUID:        rnd.UUID(),
				ClusterCIDR: "192.0.2.0/24",
				Secrets:     &cluster.RegisterSecrets{ClientSecret: cluster.ExampleClientSecret},
				JWKSUrl:     jwksURL,
				Database: cluster.RegisterDatabase{
					Driver:   dsn.DriverMySQL,
					Host:     "db.local",
					Port:     3306,
					Name:     "pp_db",
					User:     "pp_user",
					Password: "pp_pw",
					DSN:      "pp_user:pp_pw@tcp(db.local:3306)/pp_db?charset=utf8mb4&parseTime=true",
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/api/v1/cluster/theme":
			// No theme for this test.
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	jwksURL = srv.URL + "/.well-known/jwks.json"
	defer srv.Close()

	c := newBootstrapTestConfig(t, "bootstrap-reg")

	// Configure Portal.
	c.Options().PortalUrl = srv.URL
	c.Options().JoinToken = cluster.ExampleJoinToken
	c.Options().SiteUrl = expectedSite
	c.Options().AdvertiseUrl = expectedSite
	expectedAppName = c.About()
	expectedAppVersion = c.Version()
	// Gate rotate=true: driver mysql and no DSN/fields.
	c.Options().DatabaseDriver = dsn.DriverMySQL
	c.Options().DatabaseDSN = ""
	c.Options().DatabaseName = ""
	c.Options().DatabaseUser = ""
	c.Options().DatabasePassword = ""

	// Run bootstrap.
	assert.NoError(t, InitConfig(c))

	// Options should be reloaded; check values.
	assert.Equal(t, cluster.ExampleClientSecret, c.NodeClientSecret())
	// DSN branch should be preferred and persisted.
	assert.Contains(t, c.Options().DatabaseDSN, "@tcp(db.local:3306)/pp_db")
	assert.Equal(t, dsn.DriverMySQL, c.Options().DatabaseDriver)
	assert.Equal(t, srv.URL+"/.well-known/jwks.json", c.JWKSUrl())
	assert.Equal(t, "192.0.2.0/24", c.ClusterCIDR())
}

func TestRegister_AllowsHTTPPortalNonLoopback(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/cluster/nodes/register" {
			hits++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cluster.RegisterResponse{})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	origTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	baseTransport, ok := origTransport.(*http.Transport)
	if !ok {
		t.Fatalf("expected http.DefaultTransport to be *http.Transport")
	}
	transport := baseTransport.Clone()
	dialer := &net.Dialer{}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if strings.HasPrefix(addr, "portal.local:") {
			addr = srv.Listener.Addr().String()
		}
		return dialer.DialContext(ctx, network, addr)
	}
	http.DefaultTransport = transport

	prevJoin := cluster.BootstrapAutoJoinEnabled
	prevTheme := cluster.BootstrapAutoThemeEnabled
	cluster.BootstrapAutoJoinEnabled = true
	cluster.BootstrapAutoThemeEnabled = false
	t.Cleanup(func() {
		cluster.BootstrapAutoJoinEnabled = prevJoin
		cluster.BootstrapAutoThemeEnabled = prevTheme
	})

	c := newBootstrapTestConfig(t, "bootstrap-http")

	u, err := url.Parse(srv.URL)
	assert.NoError(t, err)
	c.Options().PortalUrl = "http://portal.local:" + u.Port()
	c.Options().JoinToken = cluster.ExampleJoinToken

	assert.NoError(t, InitConfig(c))
	assert.Equal(t, 1, hits)
}

func TestThemeInstall_Missing(t *testing.T) {
	// Build a tiny zip in-memory with app.js, version.txt, and style.css.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	appJS, _ := zw.Create(fs.AppJsFile)
	_, _ = appJS.Write([]byte("console.log('theme');\n"))
	versionTxt, _ := zw.Create(fs.VersionTxtFile)
	_, _ = versionTxt.Write([]byte(" theme-v1 \n"))
	styleCSS, _ := zw.Create("style.css")
	_, _ = styleCSS.Write([]byte("body{}\n"))
	_ = zw.Close()

	// Fake Portal server (register -> oauth token -> theme)
	clientSecret := cluster.ExampleClientSecret
	var jwksURL2 string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cluster/nodes/register":
			w.Header().Set("Content-Type", "application/json")
			// Return NodeClientID + NodeClientSecret so bootstrap can request OAuth token
			_ = json.NewEncoder(w).Encode(cluster.RegisterResponse{
				UUID:        rnd.UUID(),
				ClusterCIDR: "198.51.100.0/24",
				Node:        cluster.Node{ClientID: "cs5gfen1bgxz7s9i", Name: "pp-node-01", Theme: "theme-v1"},
				Secrets:     &cluster.RegisterSecrets{ClientSecret: clientSecret},
				JWKSUrl:     jwksURL2,
				Theme:       "theme-v1",
			})
		case "/api/v1/oauth/token":
			w.Header().Set("Content-Type", "application/json")
			type tokenResponse struct {
				AccessToken string `json:"access_token"`
				TokenType   string `json:"token_type"`
			}
			_ = json.NewEncoder(w).Encode(tokenResponse{AccessToken: "tok", TokenType: "Bearer"})
		case "/api/v1/cluster/theme":
			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(buf.Bytes())
		default:
			http.NotFound(w, r)
		}
	}))
	jwksURL2 = srv.URL + "/.well-known/jwks.json"
	defer srv.Close()

	c := newBootstrapTestConfig(t, "bootstrap-theme")

	// Point Portal.
	c.Options().PortalUrl = srv.URL
	c.Options().JoinToken = cluster.ExampleJoinToken

	nodeThemeDir := c.NodeThemePath()
	assert.NoError(t, os.RemoveAll(nodeThemeDir))

	// Run bootstrap.
	assert.NoError(t, InitConfig(c))

	// Expect theme artifacts to exist in node theme dir and version to match portal hint.
	assert.FileExists(t, filepath.Join(nodeThemeDir, fs.AppJsFile))
	assert.FileExists(t, filepath.Join(nodeThemeDir, fs.VersionTxtFile))
	assert.Equal(t, "theme-v1", c.NodeThemeVersion())
	assert.Equal(t, nodeThemeDir, c.ThemePath())
}

func TestThemeInstall_VersionMismatch(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	appJS, _ := zw.Create(fs.AppJsFile)
	_, _ = appJS.Write([]byte("console.log('theme-v2');\n"))
	versionTxt, _ := zw.Create(fs.VersionTxtFile)
	_, _ = versionTxt.Write([]byte(" theme-v2 \n"))
	_ = zw.Close()

	clientSecret := cluster.ExampleClientSecret
	var jwksURL string
	var themeHits int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cluster/nodes/register":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cluster.RegisterResponse{
				UUID:        rnd.UUID(),
				ClusterCIDR: "198.51.100.0/24",
				Node:        cluster.Node{ClientID: "cs5gfen1bgxz7s9i", Name: "pp-node-01"},
				Secrets:     &cluster.RegisterSecrets{ClientSecret: clientSecret},
				JWKSUrl:     jwksURL,
				Theme:       "theme-v2",
			})
		case "/api/v1/oauth/token":
			w.Header().Set("Content-Type", "application/json")
			type tokenResponse struct {
				AccessToken string `json:"access_token"`
				TokenType   string `json:"token_type"`
			}
			_ = json.NewEncoder(w).Encode(tokenResponse{AccessToken: "tok", TokenType: "Bearer"})
		case "/api/v1/cluster/theme":
			themeHits++
			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(buf.Bytes())
		default:
			http.NotFound(w, r)
		}
	}))
	jwksURL = srv.URL + "/.well-known/jwks.json"
	defer srv.Close()

	c := newBootstrapTestConfig(t, "bootstrap-theme-mismatch")

	c.Options().PortalUrl = srv.URL
	c.Options().JoinToken = cluster.ExampleJoinToken

	nodeThemeDir := c.NodeThemePath()
	assert.NoError(t, os.MkdirAll(nodeThemeDir, fs.ModeDir))
	assert.NoError(t, os.WriteFile(filepath.Join(nodeThemeDir, fs.AppJsFile), []byte("console.log('theme-v1');\n"), fs.ModeFile))
	assert.NoError(t, os.WriteFile(filepath.Join(nodeThemeDir, fs.VersionTxtFile), []byte("theme-v1"), fs.ModeFile))
	assert.Equal(t, "theme-v1", c.NodeThemeVersion())

	assert.NoError(t, InitConfig(c))

	assert.Equal(t, "theme-v2", c.NodeThemeVersion())
	assert.Equal(t, 1, themeHits)
	assert.Equal(t, nodeThemeDir, c.ThemePath())
}

func TestRegister_SQLite_NoDBPersist(t *testing.T) {
	// Portal responds with DB DSN, but local driver is SQLite → must not persist DB.
	var jwksURL3 string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cluster/nodes/register":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			resp := cluster.RegisterResponse{
				Node:        cluster.Node{Name: "pp-node-01"},
				Secrets:     &cluster.RegisterSecrets{ClientSecret: cluster.ExampleClientSecret},
				ClusterCIDR: "203.0.113.0/24",
				JWKSUrl:     jwksURL3,
				Database:    cluster.RegisterDatabase{Host: "db.local", Port: 3306, Name: "pp_db", User: "pp_user", Password: "pp_pw", DSN: "pp_user:pp_pw@tcp(db.local:3306)/pp_db?charset=utf8mb4&parseTime=true"},
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	jwksURL3 = srv.URL + "/.well-known/jwks.json"
	defer srv.Close()

	c := newBootstrapTestConfig(t, "bootstrap-sqlite")

	// SQLite driver by default; set Portal.
	c.Options().PortalUrl = srv.URL
	c.Options().JoinToken = cluster.ExampleJoinToken
	// Remember original DSN so we can ensure it is not changed.
	origDSN := c.Options().DatabaseDSN
	t.Cleanup(func() { _ = os.Remove(origDSN) })

	// Run bootstrap.
	assert.NoError(t, InitConfig(c))

	// NodeClientSecret should persist, but DB should remain SQLite (no DSN update).
	assert.Equal(t, cluster.ExampleClientSecret, c.NodeClientSecret())
	assert.Equal(t, config.SQLite3, c.DatabaseDriver())
	assert.Equal(t, origDSN, c.Options().DatabaseDSN)
	assert.Equal(t, srv.URL+"/.well-known/jwks.json", c.JWKSUrl())
	assert.Equal(t, "203.0.113.0/24", c.ClusterCIDR())
}

func TestDefaultClusterDomain(t *testing.T) {
	t.Run("explicit domain", func(t *testing.T) {
		c := newBootstrapTestConfig(t, "domain-explicit")

		c.Options().ClusterDomain = "photoprism.example"
		assert.Equal(t, "photoprism.example", defaultClusterDomain(c))
	})
	t.Run("portal host fallback", func(t *testing.T) {
		c := newBootstrapTestConfig(t, "domain-portal")

		c.Options().PortalUrl = "https://portal.photoprism.example"
		assert.Equal(t, "photoprism.example", defaultClusterDomain(c))
	})
	t.Run("no portal domain", func(t *testing.T) {
		c := newBootstrapTestConfig(t, "domain-none")

		c.Options().PortalUrl = "https://localhost:8443"
		assert.Equal(t, "", defaultClusterDomain(c))
	})
	t.Run("portal ip fallback empty", func(t *testing.T) {
		c := newBootstrapTestConfig(t, "domain-ip")

		c.Options().PortalUrl = "https://203.0.113.10"
		assert.Equal(t, "", defaultClusterDomain(c))
	})
	t.Run("invalid Portal URL", func(t *testing.T) {
		c := newBootstrapTestConfig(t, "domain-invalid")

		c.Options().PortalUrl = "://bad url"
		assert.Equal(t, "", defaultClusterDomain(c))
	})
}

func TestDefaultNodeURL(t *testing.T) {
	assert.Equal(t, "https://node1.photoprism.example", defaultNodeURL("Node1", "photoprism.example"))
	assert.Equal(t, "", defaultNodeURL("", "photoprism.example"))
	assert.Equal(t, "", defaultNodeURL("node1", ""))
	assert.Equal(t, "https://node-1.photoprism.example", defaultNodeURL("NODE_1", "photoprism.example"))
}

func TestRegister_404_NoRetry(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/cluster/nodes/register" {
			hits++
			http.NotFound(w, r)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newBootstrapTestConfig(t, "bootstrap")

	c.Options().PortalUrl = srv.URL
	c.Options().JoinToken = cluster.ExampleJoinToken

	// Run bootstrap; registration should attempt once and stop on 404.
	_ = InitConfig(c)

	assert.Equal(t, 1, hits)
}

func TestThemeInstall_SkipWhenAppJsExists(t *testing.T) {
	// Portal returns a valid zip, but theme dir already has app.js → skip.
	var served int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/cluster/theme" {
			served++
			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(http.StatusOK)
			zw := zip.NewWriter(w)
			_, _ = zw.Create("style.css")
			_ = zw.Close()
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newBootstrapTestConfig(t, "bootstrap")

	c.Options().PortalUrl = srv.URL
	c.Options().JoinToken = "t0k3n"

	// Prepare theme dir with app.js
	tempTheme, err := os.MkdirTemp("", "pp-theme-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempTheme) }()
	c.SetThemePath(tempTheme)
	assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, fs.AppJsFile), []byte("// app\n"), fs.ModeFile))

	assert.NoError(t, InitConfig(c))
	// Should have skipped request because app.js already exists.
	assert.Equal(t, 0, served)
	_, statErr := os.Stat(filepath.Join(tempTheme, "style.css"))
	assert.Error(t, statErr)
}
