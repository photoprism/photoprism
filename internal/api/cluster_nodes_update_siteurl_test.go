package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Verifies that PATCH /cluster/nodes/{uuid} validates role/url fields consistently.
func TestClusterUpdateNode_SiteUrl(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

	ClusterUpdateNode(router)
	ClusterGetNode(router)

	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)

	// Seed node
	n := &reg.Node{Node: cluster.Node{Name: "pp-node-siteurl", Role: cluster.RoleInstance, UUID: rnd.UUIDv7()}}
	assert.NoError(t, regy.Put(n))
	n, err = regy.FindByName("pp-node-siteurl")
	assert.NoError(t, err)

	// Invalid scheme is rejected.
	r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, `{"SiteUrl":"ftp://invalid"}`)
	assert.Equal(t, http.StatusBadRequest, r.Code)
	n2, err := regy.FindByNodeUUID(n.UUID)
	assert.NoError(t, err)
	assert.Equal(t, "", n2.SiteUrl)

	// Valid https URL: persisted and normalized
	r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, `{"SiteUrl":"HTTPS://PHOTOS.EXAMPLE.COM"}`)
	assert.Equal(t, http.StatusOK, r.Code)
	n3, err := regy.FindByNodeUUID(n.UUID)
	assert.NoError(t, err)
	assert.Equal(t, "https://photos.example.com", n3.SiteUrl)

	// Invalid role is rejected.
	r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, `{"Role":"viewer"}`)
	assert.Equal(t, http.StatusBadRequest, r.Code)
	n4, err := regy.FindByNodeUUID(n.UUID)
	assert.NoError(t, err)
	assert.Equal(t, cluster.RoleInstance, n4.Role)

	// Invalid advertise URL is rejected.
	r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, `{"AdvertiseUrl":"ftp://node.example"}`)
	assert.Equal(t, http.StatusBadRequest, r.Code)

	// Valid role alias + advertise URL are normalized.
	r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, `{"Role":"app","AdvertiseUrl":"HTTP://N1:2342"}`)
	assert.Equal(t, http.StatusOK, r.Code)
	n5, err := regy.FindByNodeUUID(n.UUID)
	assert.NoError(t, err)
	assert.Equal(t, cluster.RoleInstance, n5.Role)
	assert.Equal(t, "http://n1:2342", n5.AdvertiseUrl)
}
