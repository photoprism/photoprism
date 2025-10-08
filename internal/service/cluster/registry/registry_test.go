package registry

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cfg "github.com/photoprism/photoprism/internal/config"
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
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-delete", t.TempDir())
	defer c.CloseDb()

	r, _ := NewClientRegistryWithConfig(c)

	// Missing / invalid uuid
	if _, err := r.Get("not-a-uuid"); err == nil {
		t.Fatalf("expected error for invalid uuid")
	}

	// Create node
	n := &Node{Node: cluster.Node{Name: "pp-del", Role: "instance", UUID: rnd.UUIDv7()}}
	assert.NoError(t, r.Put(n))
	assert.NotEmpty(t, n.ClientID)
	assert.True(t, rnd.IsUID(n.ClientID, entity.ClientUID))
	assert.True(t, rnd.IsUUID(n.UUID))

	// Get by UUID
	got, err := r.Get(n.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, n.UUID, got.UUID)
		assert.Equal(t, "pp-del", got.Name)
		assert.True(t, rnd.IsUUID(got.UUID))
		assert.True(t, rnd.IsUID(got.ClientID, entity.ClientUID))
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
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-order", t.TempDir())
	defer c.CloseDb()

	r, _ := NewClientRegistryWithConfig(c)

	a := &Node{Node: cluster.Node{Name: "pp-a", Role: "instance", UUID: rnd.UUIDv7()}}
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
			Role:         "instance",
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

	// Non-admin (default opts): redact advertise/database
	out := BuildClusterNode(n, NodeOpts{})
	assert.Equal(t, "", out.AdvertiseUrl)
	assert.Nil(t, out.Database)

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

	// BuildClusterNodes on empty input returns empty slice (not nil)
	list := BuildClusterNodes(nil, NodeOpts{})
	assert.NotNil(t, list)
	assert.Equal(t, 0, len(list))
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
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-map", t.TempDir())
	defer c.CloseDb()

	m := entity.NewClient().SetName("pp-map").SetRole("instance")
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
		assert.Equal(t, "instance", n.Role)
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

func TestClientRegistry_GetClusterNodeByUUID(t *testing.T) {
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-getbyuuid", t.TempDir())
	defer c.CloseDb()
	assert.NoError(t, c.Init())

	r, _ := NewClientRegistryWithConfig(c)
	// Insert a node with NodeUUID
	nu := rnd.UUIDv7()
	n := &Node{Node: cluster.Node{Name: "pp-getuuid", Role: "instance", UUID: nu}}
	assert.NoError(t, r.Put(n))

	// Fetch DTO by NodeUUID
	dto, err := r.GetClusterNodeByUUID(nu, NodeOpts{})
	assert.NoError(t, err)
	assert.Equal(t, "pp-getuuid", dto.Name)
	assert.Equal(t, nu, dto.UUID)
	assert.True(t, rnd.IsUUID(dto.UUID))
}

func TestClientRegistry_FindByName_NormalizesDNSLabel(t *testing.T) {
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-findname", t.TempDir())
	defer c.CloseDb()
	assert.NoError(t, c.Init())

	r, _ := NewClientRegistryWithConfig(c)
	// Create canonical node name
	n := &Node{Node: cluster.Node{Name: "my-node-prod", Role: "instance"}}
	assert.NoError(t, r.Put(n))
	// Lookup using mixed separators and case
	got, err := r.FindByName("My.Node/Prod")
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, "my-node-prod", got.Name)
	}
}
