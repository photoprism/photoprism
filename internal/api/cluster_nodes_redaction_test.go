package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Verifies redaction differences between admin and non-admin on list endpoint.
func TestClusterListNodes_Redaction(t *testing.T) {
	// Remove the fixture record
	require.NoError(t, entity.UnscopedDb().Delete(entity.Client{}, "client_uid = ?", entity.ClientFixtures.Get("node").ClientUID).Error)
	defer func() {
		require.NoError(t, entity.Db().Create(entity.ClientFixtures.Pointer("node")).Error)
	}()

	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

	ClusterListNodes(router)

	// Seed one node with internal URL and DB metadata.
	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)

	// Nodes are UUID-first; seed with a UUID v7 so the registry includes it in List().
	n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-node-redact", Role: cluster.RoleInstance, AdvertiseUrl: "http://pp-node:2342", SiteUrl: "https://photos.example.com"}}
	n.Database = &cluster.NodeDatabase{Name: "pp_db", User: "pp_user"}
	assert.NoError(t, regy.Put(n))

	// Admin session shows internal fields
	tokenAdmin := AuthenticateAdmin(app, router)
	r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/cluster/nodes", tokenAdmin)
	assert.Equal(t, http.StatusOK, r.Code)

	node := listedNode(r.Body.String(), n.UUID)
	require.True(t, node.Exists(), "seeded node should be listed")

	// Admins see the client identifier, AdvertiseUrl, and Database.
	assert.Equal(t, n.ClientID, node.Get("ClientID").String())
	assert.Equal(t, "http://pp-node:2342", node.Get("AdvertiseUrl").String())
	assert.True(t, node.Get("Database").Exists())
}

// listedNode returns the node with the given UUID from a cluster node list response.
func listedNode(body, uuid string) gjson.Result {
	return gjson.Get(body, `#(UUID=="`+uuid+`")`)
}

// Verifies redaction for client-scoped sessions (no user attached).
func TestClusterListNodes_Redaction_ClientScope(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

	// Public mode resolves every request to the admin visitor, which would hide the redaction.
	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

	ClusterListNodes(router)

	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)

	// List() selects on node_uuid, so a record seeded without one is not returned at all.
	n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-node-redact2", Role: cluster.RoleInstance, AdvertiseUrl: "http://pp-node2:2342", SiteUrl: "https://photos2.example.com"}}
	n.Database = &cluster.NodeDatabase{Name: "pp_db2", User: "pp_user2"}
	assert.NoError(t, regy.Put(n))

	// Create client session with cluster scope and no user (redacted view expected).
	sess, err := entity.AddClientSession("test-client", conf.SessionMaxAge(), "cluster", authn.GrantClientCredentials, nil)
	assert.NoError(t, err)
	token := sess.AuthToken()

	r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/cluster/nodes", token)
	assert.Equal(t, http.StatusOK, r.Code)

	node := listedNode(r.Body.String(), n.UUID)
	require.True(t, node.Exists(), "seeded node should be listed")

	// Redacted: ClientID, AdvertiseUrl, and Database; SiteUrl is visible.
	assert.Empty(t, node.Get("ClientID").String())
	assert.Empty(t, node.Get("AdvertiseUrl").String())
	assert.False(t, node.Get("Database").Exists())
	assert.Equal(t, "https://photos2.example.com", node.Get("SiteUrl").String())
}
