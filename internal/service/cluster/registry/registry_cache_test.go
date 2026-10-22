package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Registry writes go through the entity methods, so cached client lookups see each change at once.
func TestClientRegistry_ClientCache(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-client-cache")

	r, err := NewClientRegistryWithConfig(c)
	require.NoError(t, err)

	uuid := rnd.UUIDv7()
	n := &Node{Node: cluster.Node{UUID: uuid, Name: "pp-cache", Role: cluster.RoleInstance}}
	require.NoError(t, r.Put(n))
	require.NotEmpty(t, n.ClientID)

	// cached returns the client from the client cache.
	cached := func(t *testing.T) *entity.Client {
		t.Helper()
		m, err := entity.CachedClientByUID(n.ClientID)
		require.NoError(t, err)
		return &m
	}

	t.Run("Put", func(t *testing.T) {
		assert.Equal(t, cluster.RoleInstance, cached(t).ClientRole)
		require.NoError(t, r.Put(&Node{Node: cluster.Node{ClientID: n.ClientID, Name: n.Name, Role: cluster.RoleService}}))
		assert.Equal(t, cluster.RoleService, cached(t).ClientRole)
	})
	t.Run("RotateSecret", func(t *testing.T) {
		before := cached(t)
		assert.Empty(t, before.GetData().RotatedAt)
		_, err := r.RotateSecret(uuid)
		require.NoError(t, err)
		after := cached(t)
		assert.NotEmpty(t, after.GetData().RotatedAt)
	})
	t.Run("Delete", func(t *testing.T) {
		assert.False(t, cached(t).Deleted())
		require.NoError(t, r.Delete(uuid))
		assert.True(t, cached(t).Deleted())
	})
	t.Run("PurgeAllByUUID", func(t *testing.T) {
		_ = cached(t)
		require.NoError(t, r.PurgeAllByUUID(uuid))
		_, err := entity.CachedClientByUID(n.ClientID)
		assert.ErrorIs(t, err, entity.ErrClientNotFound)
	})
}
