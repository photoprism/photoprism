package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/internal/service/cluster/provisioner"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestClusterNodesRegister(t *testing.T) {
	t.Run("FeatureDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RoleInstance
		ClusterNodesRegister(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-01"}`)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})

	// Register with existing ClientID requires clientSecret
	t.Run("ExistingClientRequiresSecret", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Pre-create a node via registry and rotate to get a plaintext secret for tests
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-auth", Role: "instance"}}
		assert.NoError(t, regy.Put(n))
		nr, err := regy.RotateSecret(n.UUID)
		assert.NoError(t, err)
		secret := nr.ClientSecret

		// Missing secret → 401
		body := `{"nodeName":"pp-auth","clientId":"` + nr.ClientID + `"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusUnauthorized, r.Code)

		// Wrong secret → 401
		body = `{"nodeName":"pp-auth","clientId":"` + nr.ClientID + `","clientSecret":"WRONG"}`
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusUnauthorized, r.Code)

		// Correct secret → 200 (existing-node path)
		body = `{"nodeName":"pp-auth","clientId":"` + nr.ClientID + `","clientSecret":"` + secret + `"}`
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("MissingToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		ClusterNodesRegister(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-01"}`)
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("CreateNodeSucceedsWithProvisioner", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Provisioner is independent of the main DB; with MariaDB admin DSN configured
		// it should successfully provision and return 201.
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-01"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		body := r.Body.String()
		assert.Contains(t, body, "\"database\"")
		assert.Contains(t, body, "\"secrets\"")
		// New nodes return the client secret; include alias for clarity.
		assert.Contains(t, body, "\"clientSecret\"")
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("UUIDChangeRequiresSecret", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		// Pre-create node with a UUID
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-lock", Role: "instance"}}
		assert.NoError(t, regy.Put(n))

		// Attempt to change UUID via name without client credentials → 409
		newUUID := rnd.UUIDv7()
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-lock","nodeUUID":"`+newUUID+`"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusConflict, r.Code)
	})
	t.Run("BadAdvertiseUrlRejected", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// http scheme for public host must be rejected (require https unless localhost).
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-03","advertiseUrl":"http://example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("GoodAdvertiseUrlAccepted", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// https is allowed for public host
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-04","advertiseUrl":"https://example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// http is allowed for localhost
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-04b","advertiseUrl":"http://localhost:2342"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("SiteUrlValidation", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Reject http siteUrl for public host
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-05","siteUrl":"http://example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)

		// Accept https siteUrl
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-06","siteUrl":"https://photos.example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("NormalizeName", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Mixed separators and case should normalize to DNS label
		body := `{"nodeName":"My.Node/Name:Prod"}`
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
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Empty nodeName → 400
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":""}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RotateSecretPersistsAndRespondsOK", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Pre-create node in registry so handler goes through existing-node path
		// and rotates the secret before attempting DB ensure. Don't reuse the
		// Monitoring fixture client ID to avoid changing its secret, which is
		// used by OAuth tests running in the same package.
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{Name: "pp-node-01", Role: "instance"}}
		assert.NoError(t, regy.Put(n))

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-01","rotateSecret":true}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// Secret should have rotated and been persisted even though DB ensure failed.
		// Fetch by name (most-recently-updated) to avoid flakiness if another test adds
		// a node with the same name and a different id.
		n2, err := regy.FindByName("pp-node-01")
		assert.NoError(t, err)
		// With client-backed registry, plaintext secret is not persisted; only rotation timestamp is updated.
		assert.NotEmpty(t, n2.RotatedAt)
	})
	t.Run("ExistingNodeSiteUrlPersistsAndRespondsOK", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Pre-create node in registry so handler goes through existing-node path.
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{Name: "pp-node-02", Role: "instance"}}
		assert.NoError(t, regy.Put(n))

		// Provisioner is independent; endpoint should respond 200 and persist metadata.
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-02","siteUrl":"https://Photos.Example.COM"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// Ensure normalized/persisted siteUrl.
		n2, err := regy.FindByName("pp-node-02")
		assert.NoError(t, err)
		assert.Equal(t, "https://photos.example.com", n2.SiteUrl)
	})
	t.Run("AssignNodeUUIDWhenMissing", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Register without nodeUUID; server should assign one (UUID v7 preferred).
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-uuid"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// Response must include node.uuid
		body := r.Body.String()
		assert.Contains(t, body, "\"uuid\"")

		// Verify it is persisted in the registry
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n, err := regy.FindByName("pp-node-uuid")
		assert.NoError(t, err)
		if assert.NotNil(t, n) {
			assert.NotEmpty(t, n.UUID)
		}
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
