package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Basic FindByClientID flow with Put and DTO mapping.
func TestClientRegistry_FindByClientID(t *testing.T) {
	c := cfg.NewTestConfig("cluster-registry-find-clientid")
	defer c.CloseDb()
	assert.NoError(t, c.Init())

	r, _ := NewClientRegistryWithConfig(c)
	n := &Node{Name: "pp-find-client", Role: "instance", UUID: rnd.UUIDv7()}
	assert.NoError(t, r.Put(n))

	got, err := r.FindByClientID(n.ClientID)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, n.ClientID, got.ClientID)
		assert.Equal(t, n.UUID, got.UUID)
		assert.True(t, rnd.IsUID(got.ClientID, entity.ClientUID))
		assert.True(t, rnd.IsUUID(got.UUID))
	}
}

// Simulate client ID changing after a restore: old row removed, new row created with same NodeUUID.
func TestClientRegistry_ClientIDChangedAfterRestore(t *testing.T) {
	c := cfg.NewTestConfig("cluster-registry-clientid-restore")
	defer c.CloseDb()
	assert.NoError(t, c.Init())

	uuid := rnd.UUIDv7()
	// Original row
	a := entity.NewClient().SetName("pp-restore").SetRole("instance")
	a.NodeUUID = uuid
	assert.NoError(t, a.Create())
	oldID := a.ClientUID

	// Simulate restore: remove old row, create new row for same node UUID with new UID
	assert.NoError(t, a.Delete())
	time.Sleep(1100 * time.Millisecond)
	b := entity.NewClient().SetName("pp-restore").SetRole("instance")
	b.NodeUUID = uuid
	assert.NoError(t, b.Create())

	r, _ := NewClientRegistryWithConfig(c)

	// Old ClientID no longer valid
	_, err := r.FindByClientID(oldID)
	assert.Error(t, err)

	// UUID lookup still works and returns the new row
	got, err := r.FindByNodeUUID(uuid)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, b.ClientUID, got.ClientID)
		assert.Equal(t, uuid, got.UUID)
		assert.True(t, rnd.IsUUID(got.UUID))
		assert.True(t, rnd.IsUID(got.ClientID, entity.ClientUID))
	}
}

// Names swapped between two nodes: UUIDs must remain authoritative.
func TestClientRegistry_SwapNames_UUIDAuthoritative(t *testing.T) {
	c := cfg.NewTestConfig("cluster-registry-swap-names")
	defer c.CloseDb()
	assert.NoError(t, c.Init())

	r, _ := NewClientRegistryWithConfig(c)
	a := &Node{UUID: rnd.UUIDv7(), Name: "pp-a", Role: "instance"}
	b := &Node{UUID: rnd.UUIDv7(), Name: "pp-b", Role: "service"}
	assert.NoError(t, r.Put(a))
	assert.NoError(t, r.Put(b))

	// Swap names via UUID-targeted updates
	assert.NoError(t, r.Put(&Node{UUID: a.UUID, Name: "pp-b"}))
	time.Sleep(1100 * time.Millisecond)
	assert.NoError(t, r.Put(&Node{UUID: b.UUID, Name: "pp-a"}))

	// UUID lookups map to the correct updated names
	gotA, err := r.FindByNodeUUID(a.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, gotA) {
		assert.Equal(t, "pp-b", gotA.Name)
		assert.True(t, rnd.IsUUID(gotA.UUID))
	}
	gotB, err := r.FindByNodeUUID(b.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, gotB) {
		assert.Equal(t, "pp-a", gotB.Name)
		assert.True(t, rnd.IsUUID(gotB.UUID))
	}

	// Name-based lookup chooses latest update for each name; both exist and are valid
	byNameA, err := r.FindByName("pp-a")
	assert.NoError(t, err)
	if assert.NotNil(t, byNameA) {
		assert.Equal(t, b.UUID, byNameA.UUID)
		assert.True(t, rnd.IsUUID(byNameA.UUID))
	}
	byNameB, err := r.FindByName("pp-b")
	assert.NoError(t, err)
	if assert.NotNil(t, byNameB) {
		assert.Equal(t, a.UUID, byNameB.UUID)
		assert.True(t, rnd.IsUUID(byNameB.UUID))
	}
}

// Ensure DB driver and fields round-trip through Put → toNode → BuildClusterNode.
func TestClientRegistry_DBDriverAndFields(t *testing.T) {
	c := cfg.NewTestConfig("cluster-registry-dbdriver")
	defer c.CloseDb()
	assert.NoError(t, c.Init())

	r, _ := NewClientRegistryWithConfig(c)
	n := &Node{UUID: rnd.UUIDv7(), Name: "pp-db", Role: "instance"}
	n.Database.Name = "photoprism_d123"
	n.Database.User = "photoprism_u123"
	n.Database.Driver = "mysql"
	n.Database.RotatedAt = time.Now().UTC().Format(time.RFC3339)
	assert.NoError(t, r.Put(n))

	got, err := r.FindByNodeUUID(n.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, "photoprism_d123", got.Database.Name)
		assert.Equal(t, "photoprism_u123", got.Database.User)
		assert.Equal(t, "mysql", got.Database.Driver)
	}

	// Build DTO with DB included
	dto := BuildClusterNode(*got, NodeOpts{IncludeAdvertiseUrl: true, IncludeDatabase: true})
	if assert.NotNil(t, dto.Database) {
		assert.Equal(t, "mysql", dto.Database.Driver)
		assert.Equal(t, "photoprism_d123", dto.Database.Name)
	}
}
