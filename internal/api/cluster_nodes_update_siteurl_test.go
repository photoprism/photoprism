package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Verifies that PATCH /cluster/nodes/{uuid} normalizes/validates siteUrl and persists only when valid.
func TestClusterUpdateNode_SiteUrl(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.Options().NodeRole = cluster.RolePortal

	ClusterUpdateNode(router)
	ClusterGetNode(router)

	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)
	// Seed node
	n := &reg.Node{Name: "pp-node-siteurl", Role: "instance", UUID: rnd.UUIDv7()}
	assert.NoError(t, regy.Put(n))
	n, err = regy.FindByName("pp-node-siteurl")
	assert.NoError(t, err)

	// Invalid scheme: ignored (200 OK but no update)
	r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, `{"siteUrl":"ftp://invalid"}`)
	assert.Equal(t, http.StatusOK, r.Code)
	n2, err := regy.FindByNodeUUID(n.UUID)
	assert.NoError(t, err)
	assert.Equal(t, "", n2.SiteUrl)

	// Valid https URL: persisted and normalized
	r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, `{"siteUrl":"HTTPS://PHOTOS.EXAMPLE.COM"}`)
	assert.Equal(t, http.StatusOK, r.Code)
	n3, err := regy.FindByNodeUUID(n.UUID)
	assert.NoError(t, err)
	assert.Equal(t, "https://photos.example.com", n3.SiteUrl)
}
