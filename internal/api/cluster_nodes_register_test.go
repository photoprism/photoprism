package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
)

func TestClusterNodesRegister(t *testing.T) {
	t.Run("FeatureDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RoleInstance
		ClusterNodesRegister(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-01"}`)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})

	t.Run("MissingToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		ClusterNodesRegister(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-01"}`)
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})

	t.Run("DriverConflict", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = "t0k3n"
		ClusterNodesRegister(router)

		// With SQLite driver in tests, provisioning should fail with conflict.
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-01"}`, "t0k3n")
		assert.Equal(t, http.StatusConflict, r.Code)
		assert.Contains(t, r.Body.String(), "portal database must be MySQL/MariaDB")
	})

	t.Run("BadName", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = "t0k3n"
		ClusterNodesRegister(router)

		// Empty nodeName → 400
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":""}`, "t0k3n")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("RotateSecretPersistsDespiteDBConflict", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = "t0k3n"
		ClusterNodesRegister(router)

		// Pre-create node in registry so handler goes through existing-node path
		// and rotates the secret before attempting DB ensure.
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{ID: "test-id", Name: "pp-node-01", Role: "instance"}
		assert.NoError(t, regy.Put(n))

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-01","rotateSecret":true}`, "t0k3n")
		assert.Equal(t, http.StatusConflict, r.Code) // DB conflict under SQLite

		// Secret should have rotated and been persisted even though DB ensure failed.
		// Fetch by name (most-recently-updated) to avoid flakiness if another test adds
		// a node with the same name and a different id.
		n2, err := regy.FindByName("pp-node-01")
		assert.NoError(t, err)
		// With client-backed registry, plaintext secret is not persisted; only rotation timestamp is updated.
		assert.NotEmpty(t, n2.SecretRot)
	})

	t.Run("ExistingNodeSiteUrlPersistsEvenOnDBConflict", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().JoinToken = "t0k3n"
		ClusterNodesRegister(router)

		// Pre-create node in registry so handler goes through existing-node path.
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Name: "pp-node-02", Role: "instance"}
		assert.NoError(t, regy.Put(n))

		// With SQLite driver in tests, provisioning should fail with 409, but metadata should still persist.
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"nodeName":"pp-node-02","siteUrl":"https://Photos.Example.COM"}`, "t0k3n")
		assert.Equal(t, http.StatusConflict, r.Code)

		// Ensure normalized/persisted siteUrl.
		n2, err := regy.FindByName("pp-node-02")
		assert.NoError(t, err)
		assert.Equal(t, "https://photos.example.com", n2.SiteUrl)
	})
}
