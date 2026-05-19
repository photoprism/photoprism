package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Duplicate names: FindByName should return the most recently updated.
func TestClientRegistry_DuplicateNamePrefersLatest(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-dupes")

	// Create two clients directly to simulate duplicates with same name.
	c1 := entity.NewClient().SetName("pp-dupe").SetRole(cluster.RoleInstance)
	assert.NoError(t, c1.Create())
	// Stagger times
	time.Sleep(10 * time.Millisecond)
	c2 := entity.NewClient().SetName("pp-dupe").SetRole(cluster.RoleService)
	assert.NoError(t, c2.Create())

	r, _ := NewClientRegistryWithConfig(c)
	n, err := r.FindByName("pp-dupe")
	assert.NoError(t, err)
	if assert.NotNil(t, n) {
		// Latest should be c2
		assert.Equal(t, c2.ClientUID, n.ClientID)
		assert.Equal(t, "service", n.Role)
		// IDs have expected format
		assert.True(t, rnd.IsUID(n.ClientID, entity.ClientUID))
	}
}

// Role change path: Put should update ClientRole via mapping.
func TestClientRegistry_RoleChange(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-role")

	r, _ := NewClientRegistryWithConfig(c)
	n := &Node{Node: cluster.Node{Name: "pp-role", Role: cluster.RoleService}}
	assert.NoError(t, r.Put(n))
	got, err := r.FindByName("pp-role")
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, "service", got.Role)
	}
	// Change to instance
	upd := &Node{Node: cluster.Node{ClientID: got.ClientID, Name: got.Name, Role: cluster.RoleInstance}}
	assert.NoError(t, r.Put(upd))
	got2, err := r.FindByName("pp-role")
	assert.NoError(t, err)
	if assert.NotNil(t, got2) {
		assert.Equal(t, cluster.RoleInstance, got2.Role)
	}
}

func TestClientRegistry_FindByName_NormalizesLegacyAliasAppToInstance(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-legacy-app")

	legacy := entity.NewClient()
	legacy.ClientName = "pp-legacy-app"
	legacy.ClientRole = "app"
	assert.NoError(t, legacy.Create())

	r, _ := NewClientRegistryWithConfig(c)
	n, err := r.FindByName("pp-legacy-app")
	assert.NoError(t, err)
	if assert.NotNil(t, n) {
		assert.Equal(t, cluster.RoleInstance, n.Role)
	}
}
