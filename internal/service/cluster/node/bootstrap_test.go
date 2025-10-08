package node

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestInitConfig_NoPortal_NoOp(t *testing.T) {
	c := config.NewMinimalTestConfigWithDb("bootstrap", t.TempDir())
	defer c.CloseDb()

	// Default NodeRole() resolves to instance; no Portal configured.
	assert.Equal(t, cluster.RoleInstance, c.NodeRole())
	assert.NoError(t, InitConfig(c))
}

func TestInitConfig_ServiceRole(t *testing.T) {
	c := config.NewMinimalTestConfigWithDb("bootstrap-service", t.TempDir())
	defer c.CloseDb()

	c.Options().NodeRole = cluster.RoleService

	assert.NoError(t, InitConfig(c))
}

func TestRegister_PersistSecretAndDB(t *testing.T) {
	// Fake Portal server.
	var jwksURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cluster/nodes/register":
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
					Driver:   config.MySQL,
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

	c := config.NewMinimalTestConfigWithDb("bootstrap-reg", t.TempDir())
	defer c.CloseDb()

	// Configure Portal.
	c.Options().PortalUrl = srv.URL
	c.Options().JoinToken = cluster.ExampleJoinToken
	// Gate rotate=true: driver mysql and no DSN/fields.
	c.Options().DatabaseDriver = config.MySQL
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
	assert.Equal(t, config.MySQL, c.Options().DatabaseDriver)
	assert.Equal(t, srv.URL+"/.well-known/jwks.json", c.JWKSUrl())
	assert.Equal(t, "192.0.2.0/24", c.ClusterCIDR())
}

func TestThemeInstall_Missing(t *testing.T) {
	// Build a tiny zip in-memory with one file style.css
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, _ := zw.Create("style.css")
	_, _ = f.Write([]byte("body{}\n"))
	_ = zw.Close()

	// Fake Portal server (register -> oauth token -> theme)
	clientSecret := cluster.ExampleClientSecret
	var jwksURL2 string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cluster/nodes/register":
			w.Header().Set("Content-Type", "application/json")
			// Return NodeClientID + NodeClientSecret so bootstrap can request OAuth token
			_ = json.NewEncoder(w).Encode(cluster.RegisterResponse{UUID: rnd.UUID(), ClusterCIDR: "198.51.100.0/24", Node: cluster.Node{ClientID: "cs5gfen1bgxz7s9i", Name: "pp-node-01"}, Secrets: &cluster.RegisterSecrets{ClientSecret: clientSecret}, JWKSUrl: jwksURL2})
		case "/api/v1/oauth/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "token_type": "Bearer"})
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

	c := config.NewMinimalTestConfigWithDb("bootstrap-theme", t.TempDir())
	defer c.CloseDb()

	// Point Portal.
	c.Options().PortalUrl = srv.URL
	c.Options().JoinToken = cluster.ExampleJoinToken

	// Ensure theme dir is empty and unique.
	tempTheme, err := os.MkdirTemp("", "pp-theme-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempTheme) }()
	c.SetThemePath(tempTheme)
	// Remove style.css if any left from previous runs.
	_ = os.Remove(filepath.Join(tempTheme, "style.css"))

	// Run bootstrap.
	assert.NoError(t, InitConfig(c))

	// Expect style.css to exist in theme dir.
	_, statErr := os.Stat(filepath.Join(tempTheme, "style.css"))
	assert.NoError(t, statErr)
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

	c := config.NewMinimalTestConfigWithDb("bootstrap-sqlite", t.TempDir())
	defer c.CloseDb()

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

	c := config.NewMinimalTestConfigWithDb("bootstrap", t.TempDir())
	defer c.CloseDb()

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

	c := config.NewMinimalTestConfigWithDb("bootstrap", t.TempDir())
	defer c.CloseDb()

	c.Options().PortalUrl = srv.URL
	c.Options().JoinToken = "t0k3n"

	// Prepare theme dir with app.js
	tempTheme, err := os.MkdirTemp("", "pp-theme-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempTheme) }()
	c.SetThemePath(tempTheme)
	assert.NoError(t, os.WriteFile(filepath.Join(tempTheme, "app.js"), []byte("// app\n"), fs.ModeFile))

	assert.NoError(t, InitConfig(c))
	// Should have skipped request because app.js already exists.
	assert.Equal(t, 0, served)
	_, statErr := os.Stat(filepath.Join(tempTheme, "style.css"))
	assert.Error(t, statErr)
}
