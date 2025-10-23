package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// UUID-first upsert: Put finds existing row by UUID and updates fields.
func TestClientRegistry_PutUpdateByUUID(t *testing.T) {
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-put-uuid", t.TempDir())
	defer c.CloseDb()

	r, _ := NewClientRegistryWithConfig(c)
	uuid := rnd.UUIDv7()

	// Create via UUID
	n := &Node{Node: cluster.Node{UUID: uuid, Name: "pp-uuid", Role: "instance", Labels: map[string]string{"a": "1"}}}
	assert.NoError(t, r.Put(n))
	assert.NotEmpty(t, n.ClientID)
	assert.True(t, rnd.IsUUID(n.UUID))
	assert.True(t, rnd.IsUID(n.ClientID, entity.ClientUID))

	// Update same record by UUID only; change name and labels
	upd := &Node{Node: cluster.Node{UUID: uuid, Name: "pp-uuid-new", Labels: map[string]string{"a": "2", "b": "x"}}}
	assert.NoError(t, r.Put(upd))

	got, err := r.FindByNodeUUID(uuid)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		// Still the same underlying client row
		assert.Equal(t, n.ClientID, got.ClientID)
		assert.Equal(t, "pp-uuid-new", got.Name)
		assert.Equal(t, "2", got.Labels["a"])
		assert.Equal(t, "x", got.Labels["b"])
		assert.True(t, rnd.IsUUID(got.UUID))
		assert.True(t, rnd.IsUID(got.ClientID, entity.ClientUID))
	}
}

// Latest-by-UpdatedAt when multiple rows share the same NodeUUID (historical duplicates).
func TestClientRegistry_FindByNodeUUID_PrefersLatest(t *testing.T) {
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-find-uuid-latest", t.TempDir())
	defer c.CloseDb()

	uuid := rnd.UUIDv7()
	// Create two raw client rows with the same NodeUUID and different UpdatedAt
	c1 := entity.NewClient().SetName("pp-dup-1").SetRole("instance")
	c1.NodeUUID = uuid
	assert.NoError(t, c1.Create())
	time.Sleep(1100 * time.Millisecond)
	c2 := entity.NewClient().SetName("pp-dup-2").SetRole("service")
	c2.NodeUUID = uuid
	assert.NoError(t, c2.Create())

	r, _ := NewClientRegistryWithConfig(c)
	got, err := r.FindByNodeUUID(uuid)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		// Should return the most recently updated row (c2)
		assert.Equal(t, c2.ClientUID, got.ClientID)
		assert.Equal(t, "service", got.Role)
		assert.Equal(t, "pp-dup-2", got.Name)
	}
}

// DeleteAllByUUID removes all rows that share a NodeUUID.
func TestClientRegistry_DeleteAllByUUID(t *testing.T) {
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-delete-all", t.TempDir())
	defer c.CloseDb()

	uuid := rnd.UUIDv7()
	// Two rows with same UUID
	a := entity.NewClient().SetName("pp-del-a").SetRole("instance")
	a.NodeUUID = uuid
	assert.NoError(t, a.Create())
	b := entity.NewClient().SetName("pp-del-b").SetRole("service")
	b.NodeUUID = uuid
	assert.NoError(t, b.Create())

	r, _ := NewClientRegistryWithConfig(c)
	assert.NoError(t, r.DeleteAllByUUID(uuid))

	// Ensure no rows remain for this UUID
	assert.Empty(t, entity.FindClientsByNodeUUID(uuid))
}

// List() should only include clients that represent cluster nodes (i.e., have a NodeUUID).
func TestClientRegistry_ListOnlyUUID(t *testing.T) {
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-list-only-uuid", t.TempDir())
	defer c.CloseDb()

	// Create one client with empty NodeUUID (non-node), and one proper node
	nonNode := entity.NewClient().SetName("webapp").SetRole("client")
	assert.NoError(t, nonNode.Create())
	node := entity.NewClient().SetName("pp-node").SetRole("instance")
	node.NodeUUID = rnd.UUIDv7()
	assert.NoError(t, node.Create())

	r, _ := NewClientRegistryWithConfig(c)
	list, err := r.List()
	assert.NoError(t, err)
	// Only the NodeUUID-backed record should be present
	if assert.Equal(t, 1, len(list)) {
		assert.Equal(t, "pp-node", list[0].Name)
		assert.NotEmpty(t, list[0].UUID)
	}
}

// Put should prefer UUID over ClientID when both are provided, avoiding cross-attachment.
func TestClientRegistry_PutPrefersUUIDOverClientID(t *testing.T) {
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-put-prefers-uuid", t.TempDir())
	defer c.CloseDb()

	r, _ := NewClientRegistryWithConfig(c)
	// Seed two separate records
	n1 := &Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-a", Role: "instance"}}
	assert.NoError(t, r.Put(n1))
	n2 := &Node{Node: cluster.Node{Name: "pp-b", Role: "service"}}
	assert.NoError(t, r.Put(n2))

	// Now attempt to update by UUID of n1 while also passing n2.ClientID:
	// implementation must use UUID and not attach to n2.
	upd := &Node{Node: cluster.Node{UUID: n1.UUID, ClientID: n2.ClientID, Role: "service"}}
	assert.NoError(t, r.Put(upd))

	got1, err := r.FindByNodeUUID(n1.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, got1) {
		assert.Equal(t, "service", got1.Role)
		assert.Equal(t, n1.ClientID, got1.ClientID)
	}
	// n2 should remain unchanged
	got2 := entity.FindClientByUID(n2.ClientID)
	if assert.NotNil(t, got2) {
		assert.Equal(t, "service", got2.ClientRole)
		assert.NotEqual(t, got2.ClientUID, got1.ClientID)
	}
}
