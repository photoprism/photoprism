package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// Duplicate names: FindByName should return the most recently updated.
func TestClientRegistry_DuplicateNamePrefersLatest(t *testing.T) {
	c := cfg.NewTestConfig("cluster-registry-dupes")
	defer c.CloseDb()
	assert.NoError(t, c.Init())

	// Create two clients directly to simulate duplicates with same name.
	c1 := entity.NewClient().SetName("pp-dupe").SetRole("instance")
	assert.NoError(t, c1.Create())
	// Stagger times
	time.Sleep(10 * time.Millisecond)
	c2 := entity.NewClient().SetName("pp-dupe").SetRole("service")
	assert.NoError(t, c2.Create())

	r, _ := NewClientRegistryWithConfig(c)
	n, err := r.FindByName("pp-dupe")
	assert.NoError(t, err)
	if assert.NotNil(t, n) {
		// Latest should be c2
		assert.Equal(t, c2.ClientUID, n.ID)
		assert.Equal(t, "service", n.Role)
	}
}

// Role change path: Put should update ClientRole via mapping.
func TestClientRegistry_RoleChange(t *testing.T) {
	c := cfg.NewTestConfig("cluster-registry-role")
	defer c.CloseDb()
	assert.NoError(t, c.Init())

	r, _ := NewClientRegistryWithConfig(c)
	n := &Node{Name: "pp-role", Role: "service"}
	assert.NoError(t, r.Put(n))
	got, err := r.FindByName("pp-role")
	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Equal(t, "service", got.Role)
	}
	// Change to instance
	upd := &Node{ID: got.ID, Name: got.Name, Role: "instance"}
	assert.NoError(t, r.Put(upd))
	got2, err := r.FindByName("pp-role")
	assert.NoError(t, err)
	if assert.NotNil(t, got2) {
		assert.Equal(t, "instance", got2.Role)
	}
}
