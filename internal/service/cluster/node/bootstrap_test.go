package node

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
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
	"gopkg.in/yaml.v2"

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

	// Default NodeRole() resolves to instance; no Portal configured.
	assert.Equal(t, cluster.RoleInstance, c.NodeRole())
	assert.NoError(t, InitConfig(c))
}

func TestInitConfig_ServiceRole(t *testing.T) {
	c := newBootstrapTestConfig(t, "bootstrap-service")

	c.Options().NodeRole = cluster.RoleService

	assert.NoError(t, InitConfig(c))
}

func TestInitConfig_ClusterOIDCWithoutBootstrap(t *testing.T) {
	// A registered instance must wire its OIDC RP from the persisted node credentials
	// even when both bootstrap toggles are disabled, since OIDC derivation is not a
	// bootstrap-policy concern.
	origJoin, origTheme := cluster.BootstrapAutoJoinEnabled, cluster.BootstrapAutoThemeEnabled
	cluster.BootstrapAutoJoinEnabled = false
	cluster.BootstrapAutoThemeEnabled = false
	t.Cleanup(func() {
		cluster.BootstrapAutoJoinEnabled = origJoin
		cluster.BootstrapAutoThemeEnabled = origTheme
	})

	c := newBootstrapTestConfig(t, "init-cluster-oidc")
	c.Options().NodeRole = cluster.RoleInstance
	c.Options().ClusterOIDC = true
	c.Options().SiteUrl = "https://app.localssl.dev/i/pro-1/"
	c.Options().NodeClientID = cluster.ExampleClientID
	c.Options().NodeClientSecret = cluster.ExampleClientSecret

	assert.NoError(t, InitConfig(c))
	assert.Equal(t, cluster.ExampleClientID, c.OIDCClient())
	assert.Equal(t, cluster.ExampleClientSecret, c.OIDCSecret())
	assert.Equal(t, "https://app.localssl.dev/", c.OIDCUri().String())
}

func TestBootstrapClusterNode(t *testing.T) {
	t.Run("NoConfig", func(t *testing.T) {
		c := newBootstrapTestConfig(t, "bootstrap-node-noconfig")
		// No Portal URL / join token configured → no-op, no panic.
		assert.NotPanics(t, func() { bootstrapClusterNode(c) })
	})
	t.Run("InvalidPortalURL", func(t *testing.T) {
		c := newBootstrapTestConfig(t, "bootstrap-node-badurl")
		c.Options().PortalUrl = "://nope"
		c.Options().JoinToken = cluster.ExampleJoinToken
		assert.NotPanics(t, func() { bootstrapClusterNode(c) })
	})
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

	// Secret must be stored in the secret file, not written inline to options.yml.
	content, readErr := os.ReadFile(c.OptionsYaml())
	assert.NoError(t, readErr)

	var persisted map[string]any
	assert.NoError(t, yaml.Unmarshal(content, &persisted))
	_, hasInlineSecret := persisted["NodeClientSecret"]
	assert.False(t, hasInlineSecret)

	info, statErr := os.Stat(c.NodeClientSecretFile())
	assert.NoError(t, statErr)
	if statErr == nil {
		assert.Equal(t, fs.ModeSecretFile, info.Mode().Perm())
	}
}

func TestResolveNodeOIDCClient(t *testing.T) {
	// portalInstance returns a config with cluster OIDC enabled, a shared-domain
	// SiteUrl, and the node registered (client credentials persisted).
	portalInstance := func(t *testing.T, name string) *config.Config {
		c := newBootstrapTestConfig(t, name)
		c.Options().NodeRole = cluster.RoleInstance
		c.Options().ClusterOIDC = true
		c.Options().SiteUrl = "https://app.localssl.dev/i/pro-1/"
		c.Options().NodeClientID = cluster.ExampleClientID
		c.Options().NodeClientSecret = cluster.ExampleClientSecret
		return c
	}

	t.Run("DerivesFromNodeCredentialsAndDefaultsIssuer", func(t *testing.T) {
		c := portalInstance(t, "oidc-node-derive")
		resolveNodeOIDCClient(c)
		assert.Equal(t, cluster.ExampleClientID, c.OIDCClient())
		assert.Equal(t, cluster.ExampleClientSecret, c.OIDCSecret())
		// The issuer defaults to the instance's own origin root (the shared-domain
		// Portal OP), so a single flag suffices.
		assert.Equal(t, "https://app.localssl.dev/", c.OIDCUri().String())
	})
	t.Run("HonorsExplicitIssuer", func(t *testing.T) {
		c := portalInstance(t, "oidc-node-issuer")
		c.Options().OIDCUri = "https://portal.example.com/"
		resolveNodeOIDCClient(c)
		assert.Equal(t, cluster.ExampleClientID, c.OIDCClient())
		assert.Equal(t, "https://portal.example.com/", c.OIDCUri().String(), "an explicit issuer must be respected")
	})
	t.Run("DisabledIsNoOp", func(t *testing.T) {
		c := portalInstance(t, "oidc-node-disabled")
		c.Options().ClusterOIDC = false
		resolveNodeOIDCClient(c)
		assert.Equal(t, "", c.OIDCClient())
		assert.Equal(t, "", c.OIDCSecret())
	})
	t.Run("ExplicitClientWins", func(t *testing.T) {
		c := portalInstance(t, "oidc-node-explicit")
		c.Options().OIDCClient = "cs5cpu17n6gj2qo5"
		c.Options().OIDCSecret = "explicit-secret"
		c.Options().OIDCUri = "https://keycloak.example.com/realms/main"
		resolveNodeOIDCClient(c)
		assert.Equal(t, "cs5cpu17n6gj2qo5", c.OIDCClient(), "an explicit client id must not be overwritten")
		assert.Equal(t, "explicit-secret", c.OIDCSecret())
	})
	t.Run("NotRegisteredLeavesClientEmpty", func(t *testing.T) {
		c := portalInstance(t, "oidc-node-unregistered")
		c.Options().NodeClientID = ""
		c.Options().NodeClientSecret = ""
		resolveNodeOIDCClient(c)
		assert.Equal(t, "", c.OIDCClient(), "without node credentials nothing is derived")
		assert.Equal(t, "", c.OIDCSecret())
	})
	t.Run("ServiceRoleNotDerived", func(t *testing.T) {
		c := portalInstance(t, "oidc-node-service")
		c.Options().NodeRole = cluster.RoleService
		resolveNodeOIDCClient(c)
		assert.Equal(t, "", c.OIDCClient())
	})
}

func TestRegisterAuthToken_UsesJoinTokenWithoutNodeCredentials(t *testing.T) {
	c := newBootstrapTestConfig(t, "bootstrap-auth-join")
	portal, err := url.Parse("https://portal.example.test")
	assert.NoError(t, err)

	token, err := registerAuthToken(c, portal, cluster.ExampleJoinToken)
	assert.NoError(t, err)
	assert.Equal(t, cluster.ExampleJoinToken, token)
}

func TestRegisterAuthToken_UsesOAuthWithNodeCredentials(t *testing.T) {
	var tokenCalls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/oauth/token" {
			http.NotFound(w, r)
			return
		}

		tokenCalls++
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(
			t,
			"Basic "+base64.StdEncoding.EncodeToString([]byte(cluster.ExampleClientID+":"+cluster.ExampleClientSecret)),
			r.Header.Get("Authorization"),
		)
		assert.NoError(t, r.ParseForm())
		assert.Equal(t, "client_credentials", r.Form.Get("grant_type"))
		assert.Equal(t, "cluster", r.Form.Get("scope"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"oauth-node-token","token_type":"Bearer"}`))
	}))
	defer srv.Close()

	c := newBootstrapTestConfig(t, "bootstrap-auth-oauth")
	c.Options().NodeClientID = cluster.ExampleClientID
	c.Options().NodeClientSecret = cluster.ExampleClientSecret

	portal, err := url.Parse(srv.URL)
	assert.NoError(t, err)

	token, err := registerAuthToken(c, portal, cluster.ExampleJoinToken)
	assert.NoError(t, err)
	assert.Equal(t, "oauth-node-token", token)
	assert.Equal(t, 1, tokenCalls)
}

func TestRegisterAuthToken_OAuthFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/oauth/token" {
			http.NotFound(w, r)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
	}))

	defer srv.Close()

	c := newBootstrapTestConfig(t, "bootstrap-auth-failure")
	c.Options().NodeClientID = cluster.ExampleClientID
	c.Options().NodeClientSecret = "stale-secret"

	portal, err := url.Parse(srv.URL)
	assert.NoError(t, err)

	_, err = registerAuthToken(c, portal, cluster.ExampleJoinToken)

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "portal access token request failed")
		assert.Contains(t, err.Error(), "401")
	}
}

func TestInitConfig_DoesNotRetryWithJoinTokenAfterOAuthFailure(t *testing.T) {
	var tokenCalls int
	var registerCalls int

	prevTheme := cluster.BootstrapAutoThemeEnabled
	cluster.BootstrapAutoThemeEnabled = false
	t.Cleanup(func() {
		cluster.BootstrapAutoThemeEnabled = prevTheme
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/oauth/token":
			tokenCalls++
			w.WriteHeader(http.StatusUnauthorized)
		case "/api/v1/cluster/nodes/register":
			registerCalls++
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(cluster.RegisterResponse{})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newBootstrapTestConfig(t, "bootstrap-no-refresh-retry")
	c.Options().PortalUrl = srv.URL
	c.Options().JoinToken = cluster.ExampleJoinToken
	c.Options().NodeName = "pp-node-01"
	c.Options().NodeRole = cluster.RoleInstance
	c.Options().NodeClientID = cluster.ExampleClientID
	c.Options().NodeClientSecret = "stale-secret"

	assert.NoError(t, InitConfig(c))
	assert.Equal(t, 1, tokenCalls)
	assert.Equal(t, 0, registerCalls)
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
			_ = json.NewEncoder(w).Encode(map[string]string{
				"access_token": "tok",
				"token_type":   "Bearer",
			})
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
			_ = json.NewEncoder(w).Encode(map[string]string{
				"access_token": "tok",
				"token_type":   "Bearer",
			})
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
	assert.Equal(t, dsn.DriverSQLite3, c.DatabaseDriver())
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

func TestBuildRegisterPayload_DisplayName(t *testing.T) {
	c := newBootstrapTestConfig(t, "node-displayname")
	reset := func() {
		c.Options().SiteName = ""
		c.Options().AppName = ""
		c.Options().SiteTitle = ""
		c.Options().SiteCaption = ""
		c.Options().Name = ""
	}
	t.Run("PrefersSiteName", func(t *testing.T) {
		reset()
		c.Options().SiteName = "Acme Media"
		c.Options().AppName = "Family Photos"
		c.Options().SiteTitle = "Our Trip"
		assert.Equal(t, "Acme Media", buildRegisterPayload(c).DisplayName)
	})
	t.Run("PrefersAppName", func(t *testing.T) {
		reset()
		c.Options().AppName = "Family Photos"
		c.Options().SiteTitle = "Our Trip"
		c.Options().SiteCaption = "Tagline"
		assert.Equal(t, "Family Photos", buildRegisterPayload(c).DisplayName)
	})
	t.Run("FallsBackToSiteTitle", func(t *testing.T) {
		reset()
		c.Options().SiteTitle = "Our Trip"
		assert.Equal(t, "Our Trip", buildRegisterPayload(c).DisplayName)
	})
	t.Run("IgnoresSiteCaption", func(t *testing.T) {
		// SiteCaption is excluded: Plus/Pro default it to the shared marketing
		// description, so it is not a distinctive per-instance label.
		reset()
		c.Options().SiteCaption = "Browse Your Life"
		assert.Equal(t, "", buildRegisterPayload(c).DisplayName)
	})
	t.Run("EmptyWhenUnbranded", func(t *testing.T) {
		reset()
		assert.Equal(t, "", buildRegisterPayload(c).DisplayName)
	})
	t.Run("AppNameSurvivesNameAliasing", func(t *testing.T) {
		// The Pro edition aliases Name to AppName; reading the raw AppName option
		// means DisplayName still reports it instead of treating it as a default.
		reset()
		c.Options().AppName = "Studio One"
		c.Options().Name = "Studio One"
		assert.Equal(t, "Studio One", buildRegisterPayload(c).DisplayName)
	})
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
