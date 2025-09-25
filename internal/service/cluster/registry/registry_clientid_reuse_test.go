package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// When both a conflicting UUID and an existing ClientID are provided, the UUID-first
// rule prevents hijacking: the update applies to the UUID's row and does not move
// the ClientID from its original node.
func TestClientRegistry_ClientIDReuse_CannotHijackExistingUUID(t *testing.T) {
	c := cfg.NewTestConfig("cluster-registry-cid-hijack")
	defer c.CloseDb()
	assert.NoError(t, c.Init())

	r, _ := NewClientRegistryWithConfig(c)
	// Seed two independent nodes
	a := &Node{UUID: rnd.UUIDv7(), Name: "pp-a", Role: "instance"}
	b := &Node{UUID: rnd.UUIDv7(), Name: "pp-b", Role: "service"}
	assert.NoError(t, r.Put(a))
	assert.NoError(t, r.Put(b))

	// Attempt to update UUID=b while passing ClientID of a
	assert.NoError(t, r.Put(&Node{UUID: b.UUID, ClientID: a.ClientID, Role: "service"}))

	// a stays attached to its original UUID and ClientID
	gotA, err := r.FindByNodeUUID(a.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, gotA) {
		assert.Equal(t, a.ClientID, gotA.ClientID)
		assert.True(t, rnd.IsUUID(gotA.UUID))
		assert.True(t, rnd.IsUID(gotA.ClientID, entity.ClientUID))
	}
	// b remains the same client row (not replaced by a)
	gotB, err := r.FindByNodeUUID(b.UUID)
	assert.NoError(t, err)
	if assert.NotNil(t, gotB) {
		assert.Equal(t, b.ClientID, gotB.ClientID)
		assert.True(t, rnd.IsUUID(gotB.UUID))
		assert.True(t, rnd.IsUID(gotB.ClientID, entity.ClientUID))
	}
}

// If a target UUID does not exist yet, providing an existing ClientID with a new UUID
// migrates the row to the new UUID. This mirrors restore flows where a node's ClientID
// is reused for a regenerated or reassigned UUID.
func TestClientRegistry_ClientIDReuse_ChangesUUIDWhenTargetMissing(t *testing.T) {
	c := cfg.NewTestConfig("cluster-registry-cid-move")
	defer c.CloseDb()
	assert.NoError(t, c.Init())

	r, _ := NewClientRegistryWithConfig(c)
	// Seed one node
	a := &Node{UUID: rnd.UUIDv7(), Name: "pp-x", Role: "instance"}
	assert.NoError(t, r.Put(a))

	// Move the row to a new UUID by referencing the same ClientID and a new UUID
	newUUID := rnd.UUIDv7()
	assert.NoError(t, r.Put(&Node{UUID: newUUID, ClientID: a.ClientID}))

	// Old UUID no longer resolves
	_, err := r.FindByNodeUUID(a.UUID)
	assert.Error(t, err)

	// New UUID points to the same client row (ClientID unchanged)
	got, err := r.FindByNodeUUID(newUUID)
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, a.ClientID, got.ClientID)
		assert.Equal(t, newUUID, got.UUID)
		assert.True(t, rnd.IsUUID(got.UUID))
		assert.True(t, rnd.IsUID(got.ClientID, entity.ClientUID))
	}
}
