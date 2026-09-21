package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// UUID-first upsert: Put finds existing row by UUID and updates fields.
func TestClientRegistry_PutUpdateByUUID(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-put-uuid")

	r, _ := NewClientRegistryWithConfig(c)
	uuid := rnd.UUIDv7()

	// Create via UUID
	n := &Node{Node: cluster.Node{UUID: uuid, Name: "pp-uuid", Role: cluster.RoleInstance, Labels: map[string]string{"a": "1"}}}
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
	c := newRegistryTestConfig(t, "cluster-registry-find-uuid-latest")

	uuid := rnd.UUIDv7()
	// Create two raw client rows with the same NodeUUID and different UpdatedAt
	c1 := entity.NewClient().SetName("pp-dup-1").SetRole(cluster.RoleInstance)
	c1.NodeUUID = uuid
	assert.NoError(t, c1.Create())
	time.Sleep(1100 * time.Millisecond)
	c2 := entity.NewClient().SetName("pp-dup-2").SetRole(cluster.RoleService)
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

// DeleteAllByUUID retires all rows that share a NodeUUID.
func TestClientRegistry_DeleteAllByUUID(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-delete-all")

	uuid := rnd.UUIDv7()
	// Two rows with same UUID
	a := entity.NewClient().SetName("pp-del-a").SetRole(cluster.RoleInstance)
	a.NodeUUID = uuid
	assert.NoError(t, a.Create())
	b := entity.NewClient().SetName("pp-del-b").SetRole(cluster.RoleService)
	b.NodeUUID = uuid
	assert.NoError(t, b.Create())

	r, _ := NewClientRegistryWithConfig(c)
	assert.NoError(t, r.DeleteAllByUUID(uuid))

	// No row is current anymore, so the node cannot be resolved or authenticated.
	assert.Nil(t, entity.FindClientByNodeUUID(uuid))
	_, err := r.FindByNodeUUID(uuid)
	assert.ErrorIs(t, err, ErrNotFound)

	// The rows are retained, so the UUID stays taken and cannot be claimed again.
	retired := entity.FindClientsByNodeUUID(uuid)
	assert.Len(t, retired, 2)
	for i := range retired {
		assert.True(t, retired[i].Deleted())
	}

	reclaim := &Node{}
	reclaim.UUID = uuid
	reclaim.Name = "pp-del-c"

	assert.ErrorIs(t, r.Create(reclaim), ErrIdentifierMismatch)
}

// List() should only include clients that represent cluster nodes (i.e., have a NodeUUID).
func TestClientRegistry_ListOnlyUUID(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-list-only-uuid")

	// Create one client with empty NodeUUID (non-node), and one proper node
	nonNode := entity.NewClient().SetName("webapp").SetRole(acl.RoleClient.String())
	assert.NoError(t, nonNode.Create())
	node := entity.NewClient().SetName("pp-node").SetRole(cluster.RoleInstance)
	node.NodeUUID = rnd.UUIDv7()
	assert.NoError(t, node.Create())

	r, _ := NewClientRegistryWithConfig(c)
	list, err := r.List()
	assert.NoError(t, err)

	// The MariaDB test database is shared across tests (unlike the per-test SQLite
	// file), so List() also returns nodes created elsewhere. Assert on membership
	// rather than the exact count: the NodeUUID-backed record is listed and the
	// non-node client is not.
	if found := listNodeByName(list, "pp-node"); assert.NotNil(t, found, "node should be listed") {
		assert.NotEmpty(t, found.UUID)
	}

	assert.Nil(t, listNodeByName(list, "webapp"), "non-node client must be excluded")
}

// Put must refuse identifiers that name different records, so neither is attached to the other.
func TestClientRegistry_PutRejectsMismatchedIdentifiers(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-put-prefers-uuid")

	r, _ := NewClientRegistryWithConfig(c)
	// Seed two separate records
	n1 := &Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-a", Role: cluster.RoleInstance}}
	assert.NoError(t, r.Put(n1))
	n2 := &Node{Node: cluster.Node{Name: "pp-b", Role: cluster.RoleService}}
	assert.NoError(t, r.Put(n2))

	// The UUID resolves to n1 and the ClientID to n2, so neither may be written.
	upd := &Node{Node: cluster.Node{UUID: n1.UUID, ClientID: n2.ClientID, Role: cluster.RoleService}}
	assert.ErrorIs(t, r.Put(upd), ErrIdentifierMismatch)

	got1, err := r.FindByNodeUUID(n1.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, got1) {
		assert.Equal(t, cluster.RoleInstance, got1.Role)
		assert.Equal(t, n1.ClientID, got1.ClientID)
	}
	// n2 should remain unchanged
	got2 := entity.FindClientByUID(n2.ClientID)
	if assert.NotNil(t, got2) {
		assert.Equal(t, cluster.RoleService, got2.ClientRole)
		assert.NotEqual(t, got2.ClientUID, got1.ClientID)
	}
}

// Deletion releases a node name while retiring its UUID, so a replacement node joins
// under the same name as a new registration instead of writing into the retired record.
func TestClientRegistry_PutAfterDeleteReusesName(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-put-after-delete")

	r, _ := NewClientRegistryWithConfig(c)

	old := &Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-replaced", Role: cluster.RoleInstance}}
	assert.NoError(t, r.Put(old))
	assert.NoError(t, r.Delete(old.UUID))

	// The name resolves to nothing, so it may be claimed again.
	_, err := r.FindByName("pp-replaced")
	assert.ErrorIs(t, err, ErrNotFound)

	// The replacement pins no UUID, which is the shape that resolves by name alone.
	replacement := &Node{Node: cluster.Node{Name: "pp-replaced", Role: cluster.RoleInstance}}
	assert.NoError(t, r.Put(replacement))

	// A separate record, so the replacement inherits neither the identifier nor the row.
	assert.NotEqual(t, old.ClientID, replacement.ClientID)

	got, err := r.FindByName("pp-replaced")
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, replacement.ClientID, got.ClientID)
		assert.NotEqual(t, old.UUID, got.UUID)
	}

	// The retired registration stays retired and keeps its UUID reserved.
	if retired := entity.FindClientByUID(old.ClientID); assert.NotNil(t, retired) {
		assert.True(t, retired.Deleted())
	}

	assert.Len(t, entity.FindClientsByNodeUUID(old.UUID), 1)
}

// PurgeAllByUUID releases a UUID by removing every record that holds it, including the
// retired ones that keep it reserved.
func TestClientRegistry_PurgeAllByUUID(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-purge-all")

	r, _ := NewClientRegistryWithConfig(c)

	t.Run("MixedLiveAndRetired", func(t *testing.T) {
		uuid := rnd.UUIDv7()

		retired := entity.NewClient().SetName("pp-purge-a").SetRole(cluster.RoleInstance)
		retired.NodeUUID = uuid
		assert.NoError(t, retired.Create())
		assert.NoError(t, retired.Delete())

		live := entity.NewClient().SetName("pp-purge-b").SetRole(cluster.RoleService)
		live.NodeUUID = uuid
		assert.NoError(t, live.Create())

		// One retired and one live record, so the UUID is taken twice over.
		assert.Len(t, entity.FindClientsByNodeUUID(uuid), 2)

		assert.NoError(t, r.PurgeAllByUUID(uuid))

		// Nothing is left to reserve it, so it may be claimed again.
		assert.Empty(t, entity.FindClientsByNodeUUID(uuid))
		assert.Nil(t, entity.FindClientByUID(retired.ClientUID))
		assert.Nil(t, entity.FindClientByUID(live.ClientUID))

		reclaim := &Node{}
		reclaim.UUID = uuid
		reclaim.Name = "pp-purge-successor"
		assert.NoError(t, r.Create(reclaim))
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.ErrorIs(t, r.PurgeAllByUUID(rnd.UUIDv7()), ErrNotFound)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.ErrorIs(t, r.PurgeAllByUUID(""), ErrNotFound)
	})
}

// FindRetiredByNodeUUID reaches a record that no ordinary lookup resolves anymore.
func TestClientRegistry_FindRetiredByNodeUUID(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-find-retired")

	r, _ := NewClientRegistryWithConfig(c)

	t.Run("Retired", func(t *testing.T) {
		n := &Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-retired-find", Role: cluster.RoleInstance}}
		assert.NoError(t, r.Put(n))
		assert.NoError(t, r.Delete(n.UUID))

		got, err := r.FindRetiredByNodeUUID(n.UUID)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, n.ClientID, got.ClientID)
		}

		// The ordinary lookup still reports nothing, so the two do not overlap.
		_, err = r.FindByNodeUUID(n.UUID)
		assert.ErrorIs(t, err, ErrNotFound)
	})
	t.Run("LiveIsNotRetired", func(t *testing.T) {
		n := &Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-live-find", Role: cluster.RoleInstance}}
		assert.NoError(t, r.Put(n))

		_, err := r.FindRetiredByNodeUUID(n.UUID)
		assert.ErrorIs(t, err, ErrNotFound)
	})
	t.Run("PrefersRetiredOverLiveDuplicate", func(t *testing.T) {
		uuid := rnd.UUIDv7()

		retired := entity.NewClient().SetName("pp-retired-dup-a").SetRole(cluster.RoleInstance)
		retired.NodeUUID = uuid
		assert.NoError(t, retired.Create())
		assert.NoError(t, retired.Delete())

		live := entity.NewClient().SetName("pp-retired-dup-b").SetRole(cluster.RoleInstance)
		live.NodeUUID = uuid
		assert.NoError(t, live.Create())

		// Each lookup answers for its own half of the split.
		got, err := r.FindRetiredByNodeUUID(uuid)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, retired.ClientUID, got.ClientID)
		}

		current, err := r.FindByNodeUUID(uuid)
		assert.NoError(t, err)
		if assert.NotNil(t, current) {
			assert.Equal(t, live.ClientUID, current.ClientID)
		}
	})
	t.Run("Empty", func(t *testing.T) {
		_, err := r.FindRetiredByNodeUUID("")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

// Put must not write into a retired record that a ClientID names, because its identifiers
// are reserved rather than available.
func TestClientRegistry_PutRejectsRetiredClientID(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-put-retired-id")

	r, _ := NewClientRegistryWithConfig(c)

	n := &Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-retired-id", Role: cluster.RoleInstance}}
	assert.NoError(t, r.Put(n))
	assert.NoError(t, r.Delete(n.UUID))

	upd := &Node{}
	upd.ClientID = n.ClientID
	upd.Name = "pp-retired-id-new"
	upd.Role = cluster.RoleInstance

	assert.ErrorIs(t, r.Put(upd), ErrIdentifierMismatch)

	// The record stays retired rather than being revived by the refused write.
	if retired := entity.FindClientByUID(n.ClientID); assert.NotNil(t, retired) {
		assert.True(t, retired.Deleted())
		assert.Equal(t, "pp-retired-id", retired.ClientName)
	}
}
