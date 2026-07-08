package registry

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestMain ensures SQLite test DB artifacts are purged after the suite runs.
func TestMain(m *testing.M) {
	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	// Run unit tests.
	code := m.Run()

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}

func TestClientRegistry_GetAndDelete(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-delete")

	r, _ := NewClientRegistryWithConfig(c)

	// Missing / invalid uuid
	if _, err := r.Get("not-a-uuid"); err == nil {
		t.Fatalf("expected error for invalid uuid")
	}

	// Create node with an instance-declared group policy (write-control
	// GroupsSrc marks this write's provenance).
	n := &Node{Node: cluster.Node{Name: "pp-del", Role: cluster.RoleInstance, UUID: rnd.UUIDv7()}}
	n.AllowGroups = []string{"g-ops"}
	n.GroupsSrc = entity.ClientGroupsSrcNode
	assert.NoError(t, r.Put(n))
	assert.NotEmpty(t, n.ClientID)
	assert.True(t, rnd.IsUID(n.ClientID, entity.ClientUID))
	assert.True(t, rnd.IsUUID(n.UUID))
	// Put reflects the persisted provenance into the read DTO field.
	assert.Equal(t, entity.ClientGroupsSrcNode, n.Node.GroupsSrc)

	// Get by UUID
	got, err := r.Get(n.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, n.UUID, got.UUID)
		assert.Equal(t, "pp-del", got.Name)
		assert.True(t, rnd.IsUUID(got.UUID))
		assert.True(t, rnd.IsUID(got.ClientID, entity.ClientUID))
		assert.Equal(t, []string{"g-ops"}, got.AllowGroups)
		// toNode maps ClientData.GroupsSrc onto the read DTO field; got.GroupsSrc
		// (registry.Node) is the separate write-control field, hence the qualifier.
		assert.Equal(t, entity.ClientGroupsSrcNode, got.Node.GroupsSrc)
	}

	// Delete by UUID
	assert.NoError(t, r.Delete(n.UUID))

	// Now missing
	_, err = r.Get(n.UUID)
	assert.Error(t, err)
	_, err = r.FindByName("pp-del")
	assert.Error(t, err)

	// Deleting again yields not found
	assert.Error(t, r.Delete(n.UUID))
}

func TestClientRegistry_ListOrderByUpdatedAtDesc(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-order")

	r, _ := NewClientRegistryWithConfig(c)

	a := &Node{Node: cluster.Node{Name: "pp-a", Role: cluster.RoleInstance, UUID: rnd.UUIDv7()}}
	b := &Node{Node: cluster.Node{Name: "pp-b", Role: "service", UUID: rnd.UUIDv7()}}
	assert.NoError(t, r.Put(a))
	// Ensure distinct UpdatedAt values (DBs often have second precision)
	time.Sleep(1100 * time.Millisecond)
	assert.NoError(t, r.Put(b))

	// Update a to make it most recent
	time.Sleep(1100 * time.Millisecond)
	assert.NoError(t, r.Put(&Node{Node: cluster.Node{ClientID: a.ClientID, Name: a.Name}}))

	list, err := r.List()
	assert.NoError(t, err)
	if assert.GreaterOrEqual(t, len(list), 2) {
		// First should be the most recently updated (a)
		assert.Equal(t, "pp-a", list[0].Name)
		// Basic ID shape checks
		assert.True(t, rnd.IsUUID(list[0].UUID))
		assert.True(t, rnd.IsUID(list[0].ClientID, entity.ClientUID))
	}
}

func TestResponseBuilders_RedactionAndOpts(t *testing.T) {
	// Base node with all fields
	n := Node{
		Node: cluster.Node{
			ClientID:     "cs5gfen1bgxz7s9i",
			Name:         "pp-node",
			Role:         cluster.RoleInstance,
			SiteUrl:      "https://photos.example.com",
			AdvertiseUrl: "http://node:2342",
			Labels:       map[string]string{"env": "prod"},
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
			UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
		},
	}
	dbInfo := n.ensureDatabase()
	dbInfo.Name = "dbn"
	dbInfo.User = "dbu"
	dbInfo.RotatedAt = time.Now().UTC().Format(time.RFC3339)
	fullView := true
	n.AllowGroups = []string{"media-acme-admin"}
	n.AllowGroupRoles = map[string]string{"media-acme-admin": "admin"}
	n.GroupsFullView = &fullView
	// Set the read DTO field, not the registry.Node write-control GroupsSrc.
	n.Node.GroupsSrc = entity.ClientGroupsSrcNode

	// Non-admin (default opts): redact advertise/database and access rules
	out := BuildClusterNode(n, NodeOpts{})
	assert.Equal(t, "", out.AdvertiseUrl)
	assert.Nil(t, out.Database)
	assert.Nil(t, out.AllowGroups)
	assert.Nil(t, out.AllowGroupRoles)
	assert.Nil(t, out.GroupsFullView)
	assert.Empty(t, out.GroupsSrc)

	// Include advertise only
	out2 := BuildClusterNode(n, NodeOpts{IncludeAdvertiseUrl: true})
	assert.Equal(t, "http://node:2342", out2.AdvertiseUrl)
	assert.Nil(t, out2.Database)

	// Include advertise + database
	out3 := BuildClusterNode(n, NodeOpts{IncludeAdvertiseUrl: true, IncludeDatabase: true})
	if assert.NotNil(t, out3.Database) {
		assert.Equal(t, "dbn", out3.Database.Name)
		assert.Equal(t, "dbu", out3.Database.User)
	}
	assert.Nil(t, out3.AllowGroups, "access rules require IncludeAccessRules")

	// Include access rules (full admin view)
	out4 := BuildClusterNode(n, NodeOpts{IncludeAdvertiseUrl: true, IncludeDatabase: true, IncludeAccessRules: true})
	assert.Equal(t, []string{"media-acme-admin"}, out4.AllowGroups)
	assert.Equal(t, map[string]string{"media-acme-admin": "admin"}, out4.AllowGroupRoles)
	if assert.NotNil(t, out4.GroupsFullView) {
		assert.True(t, *out4.GroupsFullView)
	}
	assert.Equal(t, entity.ClientGroupsSrcNode, out4.GroupsSrc)

	// BuildClusterNodes on empty input returns empty slice (not nil)
	list := BuildClusterNodes(nil, NodeOpts{})
	assert.NotNil(t, list)
	assert.Equal(t, 0, len(list))
}

func TestApplyGroupConfig(t *testing.T) {
	boolPtr := func(v bool) *bool { return &v }
	declared := func(groups []string, roles map[string]string, fullView bool) *Node {
		n := &Node{GroupsSrc: entity.ClientGroupsSrcNode}
		n.AllowGroups = groups
		n.AllowGroupRoles = roles
		if fullView {
			n.GroupsFullView = boolPtr(true)
		}
		return n
	}

	t.Run("NodeDeclares", func(t *testing.T) {
		data := &entity.ClientData{}
		applyGroupConfig(data, declared([]string{"g1"}, map[string]string{"g1": "admin"}, true))
		assert.Equal(t, []string{"g1"}, data.AllowGroups)
		assert.Equal(t, map[string]string{"g1": "admin"}, data.AllowGroupRoles)
		assert.True(t, data.GroupsFullView)
		assert.Equal(t, entity.ClientGroupsSrcNode, data.GroupsSrc)
	})
	t.Run("NodeNeverClobbersManual", func(t *testing.T) {
		data := &entity.ClientData{AllowGroups: []string{"pinned"}, GroupsSrc: entity.ClientGroupsSrcManual}
		applyGroupConfig(data, declared([]string{"g1"}, nil, false))
		assert.Equal(t, []string{"pinned"}, data.AllowGroups)
		assert.Equal(t, entity.ClientGroupsSrcManual, data.GroupsSrc)
	})
	t.Run("EmptyDeclarationClearsNodeValues", func(t *testing.T) {
		data := &entity.ClientData{AllowGroups: []string{"g1"}, GroupsFullView: true, GroupsSrc: entity.ClientGroupsSrcNode}
		applyGroupConfig(data, declared(nil, nil, false))
		assert.Empty(t, data.AllowGroups)
		assert.False(t, data.GroupsFullView)
		assert.Equal(t, "", data.GroupsSrc)
	})
	t.Run("EmptyDeclarationKeepsUnmanagedValues", func(t *testing.T) {
		data := &entity.ClientData{AllowGroups: []string{"legacy"}}
		applyGroupConfig(data, declared(nil, nil, false))
		assert.Equal(t, []string{"legacy"}, data.AllowGroups, "values without provenance must not be cleared")
	})
	t.Run("ManualEditPins", func(t *testing.T) {
		data := &entity.ClientData{AllowGroups: []string{"g1"}, GroupsSrc: entity.ClientGroupsSrcNode}
		n := &Node{GroupsSrc: entity.ClientGroupsSrcManual}
		n.AllowGroups = []string{"g2"}
		applyGroupConfig(data, n)
		assert.Equal(t, []string{"g2"}, data.AllowGroups)
		assert.Equal(t, entity.ClientGroupsSrcManual, data.GroupsSrc)
	})
	t.Run("ManualClearAllUnpins", func(t *testing.T) {
		data := &entity.ClientData{AllowGroups: []string{"g1"}, GroupsFullView: true, GroupsSrc: entity.ClientGroupsSrcManual}
		n := &Node{GroupsSrc: entity.ClientGroupsSrcManual}
		n.AllowGroups = []string{}
		n.AllowGroupRoles = map[string]string{}
		n.GroupsFullView = boolPtr(false)
		applyGroupConfig(data, n)
		assert.Empty(t, data.AllowGroups)
		assert.Equal(t, "", data.GroupsSrc, "a fully cleared manual config must un-pin")
	})
	t.Run("ManualNoFieldsNoChange", func(t *testing.T) {
		data := &entity.ClientData{AllowGroups: []string{"g1"}, GroupsSrc: entity.ClientGroupsSrcNode}
		applyGroupConfig(data, &Node{GroupsSrc: entity.ClientGroupsSrcManual})
		assert.Equal(t, []string{"g1"}, data.AllowGroups)
		assert.Equal(t, entity.ClientGroupsSrcNode, data.GroupsSrc)
	})
	t.Run("UnmanagedCallerLegacySemantics", func(t *testing.T) {
		data := &entity.ClientData{AllowGroups: []string{"g1"}, GroupsSrc: entity.ClientGroupsSrcNode}
		n := &Node{}
		n.AllowGroups = []string{"g2"}
		applyGroupConfig(data, n)
		assert.Equal(t, []string{"g2"}, data.AllowGroups)
		assert.Equal(t, entity.ClientGroupsSrcNode, data.GroupsSrc, "no provenance update without a source")
	})
}

func TestNodeOptsForSession_AdminVsNonAdmin(t *testing.T) {
	// Admin: SuperAdmin=true suffices for IsAdmin()
	admin := &entity.User{SuperAdmin: true}
	sAdmin, _ := entity.NewSession(0, 0), (&entity.User{})
	sAdmin.SetUser(admin)
	optsA := NodeOptsForSession(sAdmin)
	assert.True(t, optsA.IncludeAdvertiseUrl)
	assert.True(t, optsA.IncludeDatabase)

	// Non-admin: empty session/user
	s := &entity.Session{}
	opts := NodeOptsForSession(s)
	assert.False(t, opts.IncludeAdvertiseUrl)
	assert.False(t, opts.IncludeDatabase)

	// Nil session defaults to redacted
	optsNil := NodeOptsForSession(nil)
	assert.False(t, optsNil.IncludeAdvertiseUrl)
	assert.False(t, optsNil.IncludeDatabase)
}

func TestToNode_Mapping(t *testing.T) {
	newRegistryTestConfig(t, "cluster-registry-map")

	m := entity.NewClient().SetName("pp-map").SetRole(cluster.RoleInstance)
	m.NodeUUID = rnd.UUIDv7()
	m.ClientURL = "http://pp-map:2342"
	data := m.GetData()
	data.Labels = map[string]string{"tier": "gold"}
	data.SiteURL = "https://photos.example.com"
	data.Database = &entity.ClientDatabase{Name: "dbn", User: "dbu", RotatedAt: time.Now().UTC().Format(time.RFC3339)}
	m.SetData(data)
	assert.NoError(t, m.Create())

	n := toNode(m)
	if assert.NotNil(t, n) {
		assert.Equal(t, "pp-map", n.Name)
		assert.Equal(t, cluster.RoleInstance, n.Role)
		assert.Equal(t, "http://pp-map:2342", n.AdvertiseUrl)
		assert.Equal(t, "gold", n.Labels["tier"])
		assert.Equal(t, "https://photos.example.com", n.SiteUrl)
		assert.Equal(t, "dbn", n.Database.Name)
		assert.Equal(t, "dbu", n.Database.User)
		_, err := time.Parse(time.RFC3339, n.CreatedAt)
		assert.NoError(t, err)
		_, err = time.Parse(time.RFC3339, n.UpdatedAt)
		assert.NoError(t, err)
	}
}

// TestClientRegistry_RedirectURIs_RoundTrip confirms RedirectURIs flow
// through Put → reload → toNode unchanged, that a nil slice on update
// means "no change" (preserves the previous set), and that a non-nil
// slice (even empty) replaces it.
func TestClientRegistry_RedirectURIs_RoundTrip(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-redirecturis")
	assert.NoError(t, c.Init())
	r, _ := NewClientRegistryWithConfig(c)

	uri1 := "https://photos.example.com/api/v1/oidc/redirect"
	uri2 := "http://127.0.0.1:2342/api/v1/oidc/redirect"

	n := &Node{Node: cluster.Node{
		Name:         "pp-redir",
		Role:         cluster.RoleInstance,
		UUID:         rnd.UUIDv7(),
		RedirectURIs: []string{uri1, uri2},
	}}
	assert.NoError(t, r.Put(n))

	reloaded, err := r.FindByNodeUUID(n.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, reloaded) {
		assert.Equal(t, []string{uri1, uri2}, reloaded.RedirectURIs)
	}

	// nil on a subsequent Put preserves the previously persisted set.
	reloaded.RedirectURIs = nil
	reloaded.Name = "pp-redir-renamed"
	assert.NoError(t, r.Put(reloaded))

	again, err := r.FindByNodeUUID(n.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, again) {
		assert.Equal(t, []string{uri1, uri2}, again.RedirectURIs, "nil slice must not clear persisted redirect uris")
	}

	// Empty non-nil slice replaces (clears) the persisted set.
	again.RedirectURIs = []string{}
	assert.NoError(t, r.Put(again))

	cleared, err := r.FindByNodeUUID(n.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, cleared) {
		assert.Empty(t, cleared.RedirectURIs, "empty slice must clear the persisted set")
	}
}

func TestClientRegistry_GetClusterNodeByUUID(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-getbyuuid")
	assert.NoError(t, c.Init())

	r, _ := NewClientRegistryWithConfig(c)
	// Insert a node with NodeUUID
	nu := rnd.UUIDv7()
	n := &Node{Node: cluster.Node{Name: "pp-getuuid", Role: cluster.RoleInstance, UUID: nu}}
	assert.NoError(t, r.Put(n))

	// Fetch DTO by NodeUUID
	dto, err := r.GetClusterNodeByUUID(nu, NodeOpts{})
	assert.NoError(t, err)
	assert.Equal(t, "pp-getuuid", dto.Name)
	assert.Equal(t, nu, dto.UUID)
	assert.True(t, rnd.IsUUID(dto.UUID))
}

func TestClientRegistry_FindByName_NormalizesDNSLabel(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-findname")
	assert.NoError(t, c.Init())

	r, _ := NewClientRegistryWithConfig(c)
	// Create canonical node name
	n := &Node{Node: cluster.Node{Name: "my-node-prod", Role: cluster.RoleInstance}}
	assert.NoError(t, r.Put(n))
	// Lookup using mixed separators and case
	got, err := r.FindByName("My.Node/Prod")
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, "my-node-prod", got.Name)
	}
}
