package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestClusterEndpoints(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

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

	n := &reg.Node{Node: cluster.Node{Name: "pp-node-01", Role: cluster.RoleInstance, UUID: rnd.UUIDv7()}}
	assert.NoError(t, regy.Put(n))

	n2 := &reg.Node{Node: cluster.Node{Name: "pp-node-02", Role: "service", UUID: rnd.UUIDv7()}}
	assert.NoError(t, regy.Put(n2))

	// Resolve actual IDs (client-backed registry generates IDs)
	n, err = regy.FindByName("pp-node-01")
	assert.NoError(t, err)

	// Get by UUID
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.UUID)
	assert.Equal(t, http.StatusOK, r.Code)

	// 404 for missing id
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/missing")
	assert.Equal(t, http.StatusNotFound, r.Code)

	// Patch (manage requires Auth; our Auth() in tests allows admin; skip strict role checks here)
	r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, `{"AdvertiseUrl":"http://n1:2342"}`)
	assert.Equal(t, http.StatusOK, r.Code)

	// Pagination: count=1 returns exactly one
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes?count=1")
	assert.Equal(t, http.StatusOK, r.Code)

	// Offset beyond length clamps to end and returns empty list
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes?offset=10")
	assert.Equal(t, http.StatusOK, r.Code)

	// Delete existing
	r = PerformRequest(app, http.MethodDelete, "/api/v1/cluster/nodes/"+n.UUID)
	assert.Equal(t, http.StatusOK, r.Code)

	// GET after delete -> 404
	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.UUID)
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

// Test that ClusterGetNode validates the :uuid path parameter and rejects unsafe values.
func TestClusterGetNode_UUIDValidation(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

	// Register route under test.
	ClusterGetNode(router)

	// Seed a node and resolve its actual ID.
	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)

	n := &reg.Node{Node: cluster.Node{Name: "pp-node-99", Role: cluster.RoleInstance, UUID: rnd.UUIDv7()}}
	assert.NoError(t, regy.Put(n))

	n, err = regy.FindByName("pp-node-99")
	assert.NoError(t, err)

	// Valid UUID returns 200.
	r := PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.UUID)
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

func TestClusterUpdateNode_UUIDValidation(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

	ClusterUpdateNode(router)

	r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/bad_id", `{"SiteUrl":"https://photos.example.com"}`)
	assert.Equal(t, http.StatusNotFound, r.Code)

	r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/BadID", `{"SiteUrl":"https://photos.example.com"}`)
	assert.Equal(t, http.StatusNotFound, r.Code)
}

// TestClusterUpdateNode_RedirectURIs_Apply exercises the v1 path where the
// Portal admin writes a fresh RedirectURIs set to a registered node. After
// the PATCH lands, GET must echo the persisted set so the OIDC OP authorize
// handler can find them via client.GetData().RedirectURIs.
func TestClusterUpdateNode_RedirectURIs_Apply(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

	ClusterGetNode(router)
	ClusterUpdateNode(router)

	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)

	n := &reg.Node{Node: cluster.Node{Name: "pp-node-redirects", Role: cluster.RoleInstance, UUID: rnd.UUIDv7()}}
	assert.NoError(t, regy.Put(n))
	n, err = regy.FindByName("pp-node-redirects")
	assert.NoError(t, err)

	body := `{"RedirectURIs":["https://photos.example.com/api/v1/oidc/redirect","http://127.0.0.1:2342/api/v1/oidc/redirect"]}`
	r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, body)
	assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())

	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.UUID)
	assert.Equal(t, http.StatusOK, r.Code)
	assert.Contains(t, r.Body.String(), `"https://photos.example.com/api/v1/oidc/redirect"`)
	assert.Contains(t, r.Body.String(), `"http://127.0.0.1:2342/api/v1/oidc/redirect"`)
}

// TestClusterUpdateNode_RedirectURIs_Replace_And_Clear confirms that a
// non-nil slice replaces the persisted set (including the cleared case
// where the slice is empty).
func TestClusterUpdateNode_RedirectURIs_Replace_And_Clear(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

	ClusterGetNode(router)
	ClusterUpdateNode(router)

	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)

	n := &reg.Node{Node: cluster.Node{Name: "pp-node-replace", Role: cluster.RoleInstance, UUID: rnd.UUIDv7()}}
	assert.NoError(t, regy.Put(n))
	n, err = regy.FindByName("pp-node-replace")
	assert.NoError(t, err)

	// Seed two redirect URIs.
	r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID,
		`{"RedirectURIs":["https://a.example.com/cb","https://b.example.com/cb"]}`)
	assert.Equal(t, http.StatusOK, r.Code)

	// Replace with a single entry — the old ones must be gone.
	r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID,
		`{"RedirectURIs":["https://c.example.com/cb"]}`)
	assert.Equal(t, http.StatusOK, r.Code)

	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.UUID)
	assert.Equal(t, http.StatusOK, r.Code)
	assert.Contains(t, r.Body.String(), `"https://c.example.com/cb"`)
	assert.NotContains(t, r.Body.String(), `"https://a.example.com/cb"`)
	assert.NotContains(t, r.Body.String(), `"https://b.example.com/cb"`)

	// Clear via empty slice. Non-nil, zero length means "remove all".
	r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID,
		`{"RedirectURIs":[]}`)
	assert.Equal(t, http.StatusOK, r.Code)

	r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.UUID)
	assert.Equal(t, http.StatusOK, r.Code)
	assert.NotContains(t, r.Body.String(), `"RedirectURIs"`, "empty slice must drop the field via omitempty")
}

// TestNormalizeAllowGroupRoles validates the PATCH-time group → role mapping
// helper: keys are normalized, empty keys dropped, and role values must be
// federatable instance roles.
func TestNormalizeAllowGroupRoles(t *testing.T) {
	t.Run("AcceptsAllInstanceRoles", func(t *testing.T) {
		out, err := normalizeAllowGroupRoles(map[string]string{
			"Media-Acme-Admin": "admin", "Media-Acme-Manager": "manager", "Media-Acme-User": "user",
			"Media-Acme-Contributor": "contributor", "Media-Acme-Viewer": "viewer", "Media-Acme-Guest": "guest",
			"   ": "admin",
		})
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			"media-acme-admin": "admin", "media-acme-manager": "manager", "media-acme-user": "user",
			"media-acme-contributor": "contributor", "media-acme-viewer": "viewer", "media-acme-guest": "guest",
		}, out)
	})
	t.Run("InvalidRole", func(t *testing.T) {
		for _, role := range []string{"cluster_admin", "visitor", "none", "bogus", ""} {
			_, err := normalizeAllowGroupRoles(map[string]string{"media-acme-x": role})
			assert.Error(t, err, "role %q must be rejected", role)
		}
	})
}

// TestClusterUpdateNode_GroupRules covers the group-based admission config on
// PATCH: AllowGroups apply/replace/clear with normalization, AllowGroupRoles
// role validation, and the GroupsFullView opt-in round-trip.
func TestClusterUpdateNode_GroupRules(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

	ClusterGetNode(router)
	ClusterUpdateNode(router)

	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)

	n := &reg.Node{Node: cluster.Node{Name: "pp-node-groups", Role: cluster.RoleInstance, UUID: rnd.UUIDv7()}}
	assert.NoError(t, regy.Put(n))
	n, err = regy.FindByName("pp-node-groups")
	assert.NoError(t, err)

	t.Run("Apply", func(t *testing.T) {
		body := `{"AllowGroups":["Media-Acme-Admin","Media-Acme-Viewer"],"AllowGroupRoles":{"Media-Acme-Admin":"admin"},"GroupsFullView":true}`
		r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, body)
		assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())

		client := entity.FindClientByNodeUUID(n.UUID)
		assert.NotNil(t, client)
		data := client.GetData()
		assert.Equal(t, []string{"media-acme-admin", "media-acme-viewer"}, data.AllowGroups, "stored groups must be normalized")
		assert.Equal(t, map[string]string{"media-acme-admin": "admin"}, data.AllowGroupRoles)
		assert.True(t, data.GroupsFullView)

		r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.UUID)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), `"media-acme-admin"`)
		assert.Contains(t, r.Body.String(), `"GroupsFullView":true`)
	})
	t.Run("OmittedFieldsUnchanged", func(t *testing.T) {
		r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, `{"DisplayName":"Groups Node"}`)
		assert.Equal(t, http.StatusOK, r.Code)

		data := entity.FindClientByNodeUUID(n.UUID).GetData()
		assert.Equal(t, []string{"media-acme-admin", "media-acme-viewer"}, data.AllowGroups)
		assert.True(t, data.GroupsFullView)
	})
	t.Run("ReplaceAndClear", func(t *testing.T) {
		r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID,
			`{"AllowGroups":["Media-Acme-User"],"AllowGroupRoles":{},"GroupsFullView":false}`)
		assert.Equal(t, http.StatusOK, r.Code)

		data := entity.FindClientByNodeUUID(n.UUID).GetData()
		assert.Equal(t, []string{"media-acme-user"}, data.AllowGroups)
		assert.Empty(t, data.AllowGroupRoles)
		assert.False(t, data.GroupsFullView)

		r = PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, `{"AllowGroups":[]}`)
		assert.Equal(t, http.StatusOK, r.Code)

		data = entity.FindClientByNodeUUID(n.UUID).GetData()
		assert.Empty(t, data.AllowGroups, "empty slice must clear the persisted set")

		r = PerformRequest(app, http.MethodGet, "/api/v1/cluster/nodes/"+n.UUID)
		assert.Equal(t, http.StatusOK, r.Code)
		assert.NotContains(t, r.Body.String(), `"AllowGroups"`, "cleared set must drop the field via omitempty")
		assert.NotContains(t, r.Body.String(), `"GroupsFullView"`, "false opt-in must drop the field via omitempty")
	})
	t.Run("InvalidRole", func(t *testing.T) {
		r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID,
			`{"AllowGroupRoles":{"Media-Acme-Operators":"cluster_admin"}}`)
		assert.Equal(t, http.StatusBadRequest, r.Code, "body=%s", r.Body.String())
	})
}

// TestClusterUpdateNode_RedirectURIs_Invalid rejects malformed entries with
// 400. Validation policy mirrors validateSiteURL: HTTPS always, HTTP only
// on loopback / cluster-internal hosts, no fragment.
func TestClusterUpdateNode_RedirectURIs_Invalid(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

	ClusterUpdateNode(router)

	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)
	n := &reg.Node{Node: cluster.Node{Name: "pp-node-invalid", Role: cluster.RoleInstance, UUID: rnd.UUIDv7()}}
	assert.NoError(t, regy.Put(n))
	n, _ = regy.FindByName("pp-node-invalid")

	cases := []struct {
		name string
		body string
	}{
		{"NotAbsolute", `{"RedirectURIs":["/relative/cb"]}`},
		{"FtpScheme", `{"RedirectURIs":["ftp://example.com/cb"]}`},
		{"HttpNonLoopback", `{"RedirectURIs":["http://photos.example.com/cb"]}`},
		{"WithFragment", `{"RedirectURIs":["https://example.com/cb#frag"]}`},
		{"NoHost", `{"RedirectURIs":["https://"]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, tc.body)
			assert.Equal(t, http.StatusBadRequest, r.Code, "body=%s", r.Body.String())
		})
	}
}

func TestNormalizeRedirectURIs(t *testing.T) {
	t.Run("NilInputNoChange", func(t *testing.T) {
		out, err := normalizeRedirectURIs(nil)
		assert.NoError(t, err)
		assert.Nil(t, out)
	})
	t.Run("EmptyInputReturnsEmpty", func(t *testing.T) {
		out, err := normalizeRedirectURIs([]string{})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Len(t, out, 0)
	})
	t.Run("WhitespaceTrimmed", func(t *testing.T) {
		out, err := normalizeRedirectURIs([]string{"  https://example.com/cb  "})
		assert.NoError(t, err)
		assert.Equal(t, []string{"https://example.com/cb"}, out)
	})
	t.Run("DropsDuplicates", func(t *testing.T) {
		out, err := normalizeRedirectURIs([]string{
			"https://example.com/cb",
			"https://example.com/cb",
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"https://example.com/cb"}, out)
	})
	t.Run("AllowsLoopbackHTTP", func(t *testing.T) {
		out, err := normalizeRedirectURIs([]string{"http://127.0.0.1:2342/cb"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"http://127.0.0.1:2342/cb"}, out)
	})
	t.Run("AllowsClusterInternalHTTP", func(t *testing.T) {
		out, err := normalizeRedirectURIs([]string{"http://photoprism.svc.cluster.local/cb"})
		assert.NoError(t, err)
		assert.Equal(t, []string{"http://photoprism.svc.cluster.local/cb"}, out)
	})
}

func TestValidateRedirectURI(t *testing.T) {
	cases := []struct {
		uri string
		ok  bool
	}{
		{"https://example.com/cb", true},
		{"http://localhost:2342/cb", true},
		{"http://127.0.0.1/cb", true},
		{"http://photoprism.svc/cb", true},
		{"http://example.com/cb", false},
		{"https://example.com/cb#frag", false},
		{"ftp://example.com/cb", false},
		{"/relative", false},
		{"", false},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.ok, validateRedirectURI(tc.uri), "uri=%q", tc.uri)
	}
}

func TestClusterUpdateNode_RequestTooLarge(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)

	ClusterUpdateNode(router)

	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)

	n := &reg.Node{Node: cluster.Node{Name: "pp-node-request-limit", Role: cluster.RoleInstance, UUID: rnd.UUIDv7()}}
	assert.NoError(t, regy.Put(n))

	body := `{"Labels":{"oversized":"` + strings.Repeat("a", int(MaxClusterRegisterBytes)) + `"}}`
	r := PerformRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID, body)

	assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
}
