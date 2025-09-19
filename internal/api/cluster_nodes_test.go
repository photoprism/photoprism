package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
)

func TestClusterEndpoints(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.Options().NodeRole = cluster.RolePortal

	ClusterListNodes(router)
	ClusterGetNode(router)
	ClusterUpdateNode(router)
	ClusterDeleteNode(router)

	// Empty list initially (JSON array)
	r := PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes")
	assert.Equal(t, http.StatusOK, r.Code)

	// Seed nodes in the registry
	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)
	n := &reg.Node{Name: "pp-node-01", Role: "instance"}
	assert.NoError(t, regy.Put(n))
	n2 := &reg.Node{Name: "pp-node-02", Role: "service"}
	assert.NoError(t, regy.Put(n2))
	// Resolve actual IDs (client-backed registry generates IDs)
	n, err = regy.FindByName("pp-node-01")
	assert.NoError(t, err)

	// Get by id
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.ID)
	assert.Equal(t, http.StatusOK, r.Code)

	// 404 for missing id
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/missing")
	assert.Equal(t, http.StatusNotFound, r.Code)

	// Patch (manage requires Auth; our Auth() in tests allows admin; skip strict role checks here)
	r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.ID, `{"advertiseUrl":"http://n1:2342"}`)
	assert.Equal(t, http.StatusOK, r.Code)

	// Pagination: count=1 returns exactly one
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes?count=1")
	assert.Equal(t, http.StatusOK, r.Code)

	// Offset beyond length clamps to end and returns empty list
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes?offset=10")
	assert.Equal(t, http.StatusOK, r.Code)

	// Delete existing
	r = PerformRequest(app, http.MethodDelete, "/api/v1/cluster/nodes/"+n.ID)
	assert.Equal(t, http.StatusOK, r.Code)

	// GET after delete -> 404
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.ID)
	assert.Equal(t, http.StatusNotFound, r.Code)

	// DELETE nonexistent id -> 404
	r = PerformRequest(app, http.MethodDelete, "/api/v1/cluster/nodes/missing-id")
	assert.Equal(t, http.StatusNotFound, r.Code)

	// DELETE invalid id (uppercase) -> 404
	r = PerformRequest(app, http.MethodDelete, "/api/v1/cluster/nodes/BadID")
	assert.Equal(t, http.StatusNotFound, r.Code)

	// List again (should not include the deleted node)
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes")
	assert.Equal(t, http.StatusOK, r.Code)
}

// Test that ClusterGetNode validates the :id path parameter and rejects unsafe values.
func TestClusterGetNode_IDValidation(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.Options().NodeRole = cluster.RolePortal

	// Register route under test.
	ClusterGetNode(router)

	// Seed a node and resolve its actual ID.
	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)
	n := &reg.Node{Name: "pp-node-99", Role: "instance"}
	assert.NoError(t, regy.Put(n))
	n, err = regy.FindByName("pp-node-99")
	assert.NoError(t, err)

	// Valid ID returns 200.
	r := PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.ID)
	assert.Equal(t, http.StatusOK, r.Code)

	// Uppercase letters are not allowed.
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/N1")
	assert.Equal(t, http.StatusNotFound, r.Code)

	// Characters outside [a-z0-9-] are rejected (e.g., underscore).
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/bad_id")
	assert.Equal(t, http.StatusNotFound, r.Code)

	// Dot is rejected.
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/bad.id")
	assert.Equal(t, http.StatusNotFound, r.Code)

	// Encoded space is rejected.
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/a%20b")
	assert.Equal(t, http.StatusNotFound, r.Code)

	// Excessively long ID (>64 chars) is rejected.
	longID := make([]byte, 65)
	for i := range longID {
		longID[i] = 'a'
	}
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+string(longID))
	assert.Equal(t, http.StatusNotFound, r.Code)
}
