package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/internal/service/cluster/provisioner"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestClusterNodesRegister(t *testing.T) {
	t.Run("FeatureDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RoleInstance
		ClusterNodesRegister(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-01"}`)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("ExistingNodeMutationRequiresOAuthToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-auth", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))
		nr, err := regy.RotateSecret(n.UUID)
		assert.NoError(t, err)

		// Join tokens must not mutate existing registrations.
		body := `{"NodeName":"pp-auth","Labels":{"env":"prod"}}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusConflict, r.Code)
		assert.Contains(t, r.Body.String(), "already registered")

		// Existing-node mutations with a valid node access token must succeed.
		token := oauthNodeAccessToken(t, app, router, conf, nr.ClientID, nr.ClientSecret)
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-auth","RotateSecret":true}`, token)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("MissingToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		ClusterNodesRegister(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-01"}`)
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		body := `{"NodeName":"pp-node-big","Labels":{"env":"` + strings.Repeat("a", 300*1024) + `"}}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
	t.Run("ClusterCIDRBlocksClientIPOutsideRange", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		prevClusterCIDR := conf.Options().ClusterCIDR
		t.Cleanup(func() {
			conf.Options().ClusterCIDR = prevClusterCIDR
		})
		conf.Options().ClusterCIDR = "192.0.2.0/24"
		ClusterNodesRegister(router)

		r := AuthenticatedRequestWithBodyAndIP(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-cidr-blocked"}`, cluster.ExampleJoinToken, "198.51.100.9")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("ClusterCIDRAllowsClientIPInRange", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		prevClusterCIDR := conf.Options().ClusterCIDR
		t.Cleanup(func() {
			conf.Options().ClusterCIDR = prevClusterCIDR
		})
		conf.Options().ClusterCIDR = "192.0.2.0/24"
		ClusterNodesRegister(router)

		r := AuthenticatedRequestWithBodyAndIP(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-cidr-allowed"}`, cluster.ExampleJoinToken, "192.0.2.42")
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("ForbiddenFromCDN", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/cluster/nodes/register", nil)
		req.Header.Set(header.CdnHost, "edge.example")
		req.Header.Set(header.Auth, header.AuthBearer+" "+cluster.ExampleJoinToken)
		req.Header.Set(header.Accept, "application/json")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("CreateNodeWithoutRotateSkipsProvisioner", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Provisioner is independent of the main DB; with MariaDB admin DSN configured
		// it should successfully provision and return 201.
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-01"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		body := r.Body.String()
		assert.Contains(t, body, "\"Database\"")
		assert.Contains(t, body, "\"Secrets\"")
		assert.Contains(t, body, "\"ClientSecret\"")
		assert.Equal(t, "", gjson.Get(body, "Database.Name").String())
		assert.False(t, gjson.Get(body, "AlreadyProvisioned").Bool())
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("CreateNodeRotateDatabaseProvisioned", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-rotate","RotateDatabase":true}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		body := r.Body.String()
		assert.NotEqual(t, "", gjson.Get(body, "Database.Name").String())
		assert.NotEqual(t, "", gjson.Get(body, "Database.Password").String())
		assert.True(t, gjson.Get(body, "AlreadyProvisioned").Bool())
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("UUIDChangeWithJoinTokenReturnsConflict", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		// Pre-create node with a UUID
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-lock", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))

		// Attempt to change UUID via join token must fail with conflict.
		newUUID := rnd.UUIDv7()
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-lock","NodeUUID":"`+newUUID+`"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusConflict, r.Code)
	})
	t.Run("AdvertiseUrlHttpAllowed", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// http scheme is allowed for cluster-internal traffic, even on public hostnames.
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-03","AdvertiseUrl":"http://example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("GoodAdvertiseUrlAccepted", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// https is allowed for public host
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-04","AdvertiseUrl":"https://example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// http is allowed for localhost
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-04b","AdvertiseUrl":"http://localhost:2342"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("SiteUrlValidation", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Reject http SiteUrl for public host
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-05","SiteUrl":"http://example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)

		// Accept https SiteUrl
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-06","SiteUrl":"https://photos.example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("NormalizeName", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Mixed separators and case should normalize to DNS label
		body := `{"NodeName":"My.Node/Name:Prod"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n, err := regy.FindByName("my-node-name-prod")
		assert.NoError(t, err)
		if assert.NotNil(t, n) {
			assert.Equal(t, "my-node-name-prod", n.Name)
		}
	})
	t.Run("BadName", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Empty nodeName → 400
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":""}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RotateSecretPersistsAndRespondsOK", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Pre-create node in registry so handler goes through existing-node path
		// and rotates the secret before attempting DB ensure. Don't reuse the
		// Monitoring fixture client ID to avoid changing its secret, which is
		// used by OAuth tests running in the same package.
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-node-01", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))
		n, err = regy.RotateSecret(n.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, n) {
			return
		}

		token := oauthNodeAccessToken(t, app, router, conf, n.ClientID, n.ClientSecret)
		body := `{"NodeName":"pp-node-01","RotateSecret":true}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// Secret should have rotated and been persisted even though DB ensure failed.
		// Fetch by name (most-recently-updated) to avoid flakiness if another test adds
		// a node with the same name and a different id.
		n2, err := regy.FindByName("pp-node-01")
		assert.NoError(t, err)
		// With client-backed registry, plaintext secret is not persisted; only rotation timestamp is updated.
		if assert.NotNil(t, n2) {
			assert.NotEmpty(t, n2.RotatedAt)
		}
	})
	t.Run("RotateSecretRequiresMatchingNodeClientCredentials", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		victim := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-victim", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(victim))
		victim, err = regy.RotateSecret(victim.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, victim) {
			return
		}

		attacker := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-attacker", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(attacker))
		attacker, err = regy.RotateSecret(attacker.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, attacker) {
			return
		}

		// Using another node's valid access token must not authorize rotating the victim.
		token := oauthNodeAccessToken(t, app, router, conf, attacker.ClientID, attacker.ClientSecret)
		body := `{"NodeName":"pp-victim","RotateSecret":true}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusConflict, r.Code)
		assert.NotContains(t, r.Body.String(), "\"ClientSecret\"")
		assert.NotContains(t, r.Body.String(), "\"Password\"")
	})
	t.Run("ExistingNodeMutationRequiresWriteScope", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		node := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-scope-check", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(node))
		node, err = regy.RotateSecret(node.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, node) {
			return
		}

		client := entity.FindClientByUID(node.ClientID)
		if assert.NotNil(t, client) {
			client.SetScope("metrics")
			assert.NoError(t, client.Save())
		}

		// Tokens that do not include cluster permissions must be denied.
		token := oauthNodeAccessToken(t, app, router, conf, node.ClientID, node.ClientSecret)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-scope-check","SiteUrl":"https://scope.example.com"}`, token)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("RotateDatabaseWithJoinTokenReturnsConflict", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{Name: "pp-node-db-rotate", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-db-rotate","RotateDatabase":true}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusConflict, r.Code)
	})
	t.Run("ExistingNodeSiteUrlPersistsAndRespondsOK", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Pre-create node in registry so handler goes through existing-node path.
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-node-02", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))

		n, err = regy.RotateSecret(n.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, n) {
			return
		}

		// Provisioner is independent; endpoint should respond 200 and persist metadata.
		token := oauthNodeAccessToken(t, app, router, conf, n.ClientID, n.ClientSecret)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-02","SiteUrl":"https://Photos.Example.COM"}`, token)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// Ensure normalized/persisted SiteUrl.
		n2, err := regy.FindByName("pp-node-02")
		assert.NoError(t, err)
		assert.Equal(t, "https://photos.example.com", n2.SiteUrl)
	})
	t.Run("ExistingNodeWithoutRotateDoesNotProvisionDatabase", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		// If provisioning is called unexpectedly, this invalid DSN should cause a 409.
		conf.Options().DatabaseProvisionDSN = "invalid-dsn"
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-node-no-provision", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))
		n, err = regy.RotateSecret(n.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, n) {
			return
		}

		token := oauthNodeAccessToken(t, app, router, conf, n.ClientID, n.ClientSecret)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-no-provision","SiteUrl":"https://photos.example.com"}`, token)
		assert.Equal(t, http.StatusOK, r.Code)
		body := r.Body.String()
		assert.False(t, gjson.Get(body, "AlreadyProvisioned").Bool())
		assert.Equal(t, "", gjson.Get(body, "Database.Name").String())
		assert.Equal(t, "", gjson.Get(body, "Database.User").String())
	})
	t.Run("AssignNodeUUIDWhenMissing", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Register without nodeUUID; server should assign one (UUID v7 preferred).
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-uuid"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// Response must include Node.UUID
		body := r.Body.String()
		assert.NotEmpty(t, gjson.Get(body, "Node.UUID").String())

		// Verify it is persisted in the registry
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n, err := regy.FindByName("pp-node-uuid")
		assert.NoError(t, err)
		if assert.NotNil(t, n) {
			assert.NotEmpty(t, n.UUID)
		}
	})
	t.Run("ThemeHintProvided", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		themeDir := conf.PortalThemePath()
		assert.NoError(t, os.MkdirAll(themeDir, fs.ModeDir))
		assert.NoError(t, os.WriteFile(filepath.Join(themeDir, fs.AppJsFile), []byte("// app\n"), fs.ModeFile))
		assert.NoError(t, os.WriteFile(filepath.Join(themeDir, fs.VersionTxtFile), []byte(" 2.0.0\n"), fs.ModeFile))
		t.Cleanup(func() { _ = os.RemoveAll(themeDir) })

		body := `{"NodeName":"pp-node-theme","Theme":"1.0.0"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		assert.Equal(t, "2.0.0", gjson.Get(r.Body.String(), "Theme").String())
		cleanupRegisterProvisioning(t, conf, r)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		node, err := regy.FindByName("pp-node-theme")
		assert.NoError(t, err)
		if assert.NotNil(t, node) {
			assert.Equal(t, "1.0.0", node.Theme)
		}

		node, err = regy.RotateSecret(node.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, node) {
			return
		}
		token := oauthNodeAccessToken(t, app, router, conf, node.ClientID, node.ClientSecret)

		body = `{"NodeName":"pp-node-theme","Theme":"2.0.0"}`
		r2 := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r2.Code)
		assert.Equal(t, "2.0.0", gjson.Get(r2.Body.String(), "Theme").String())
		cleanupRegisterProvisioning(t, conf, r2)
	})
}

func cleanupRegisterProvisioning(t *testing.T, conf *config.Config, r *httptest.ResponseRecorder) {
	t.Helper()

	if r.Code != http.StatusOK && r.Code != http.StatusCreated {
		return
	}

	var resp cluster.RegisterResponse
	if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal register response: %v", err)
	}

	if !resp.AlreadyProvisioned {
		return
	}

	name := resp.Database.Name
	user := resp.Database.User

	if conf != nil && (name == "" || user == "") && resp.Node.Name != "" && resp.Node.UUID != "" {
		genName, genUser, _ := provisioner.GenerateCredentials(conf, resp.Node.UUID, resp.Node.Name)
		if name == "" {
			name = genName
		}
		if user == "" {
			user = genUser
		}
	}

	if name == "" && user == "" {
		return
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := provisioner.DropCredentials(ctx, name, user); err != nil {
			t.Fatalf("drop credentials for %s/%s: %v", name, user, err)
		}
	})
}

func AuthenticatedRequestWithBodyAndIP(r http.Handler, method, path, body string, authToken string, clientIP string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)
	req.RemoteAddr = clientIP + ":12345"
	header.SetAuthorization(req, authToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func oauthNodeAccessToken(t testing.TB, app http.Handler, router *gin.RouterGroup, conf *config.Config, clientID, clientSecret string) string {
	return oauthNodeAccessTokenWithScope(t, app, router, conf, clientID, clientSecret, "cluster")
}

func oauthNodeAccessTokenWithScope(t testing.TB, app http.Handler, router *gin.RouterGroup, conf *config.Config, clientID, clientSecret, scope string) string {
	t.Helper()

	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() {
		conf.SetAuthMode(prevAuthMode)
	})

	OAuthToken(router)

	data := url.Values{
		"grant_type":    {authn.GrantClientCredentials.String()},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"scope":         {scope},
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Add(header.ContentType, header.ContentTypeForm)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "oauth token request failed: %s", w.Body.String())

	return gjson.Get(w.Body.String(), "access_token").String()
}

// TestValidateAdvertiseURL ensures the validator accepts HTTP and HTTPS for advertise URLs.
func TestValidateAdvertiseURL(t *testing.T) {
	cases := []struct {
		u  string
		ok bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"http://localhost:2342", true},
		{"http://photoprism.default.svc", true},
		{"http://photoprism.default.svc.cluster.local", true},
		{"http://photoprism.internal", true},
		{"https://127.0.0.1", true},
		{"ftp://example.com", false},
		{"https://", false},
		{"", false},
	}
	for _, c := range cases {
		if got := validateAdvertiseURL(c.u); got != c.ok {
			t.Fatalf("validateAdvertiseURL(%q) = %v, want %v", c.u, got, c.ok)
		}
	}
}

// TestValidateSiteURL enforces HTTPS for non-local site URLs.
func TestValidateSiteURL(t *testing.T) {
	cases := []struct {
		u  string
		ok bool
	}{
		{"https://photos.example.com", true},
		{"http://photos.example.com", false},
		{"http://127.0.0.1:2342", true},
		{"mailto:me@example.com", false},
		{"://bad", false},
	}
	for _, c := range cases {
		if got := validateSiteURL(c.u); got != c.ok {
			t.Fatalf("validateSiteURL(%q) = %v, want %v", c.u, got, c.ok)
		}
	}
}
