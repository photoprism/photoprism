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
	// Backdate c1 so c2 is unambiguously the most recently updated. Timestamps are
	// stored with second precision, so a sub-second stagger would not separate them.
	assert.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).
		Where("client_uid = ?", c1.ClientUID).
		UpdateColumn("updated_at", entity.Now().Add(-time.Hour)).Error)
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

// DisplayName: instance-reported names update freely until an admin pins one;
// an admin override is sticky across registrations, and clearing it un-pins.
func TestClientRegistry_DisplayNameOverride(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-displayname")
	r, _ := NewClientRegistryWithConfig(c)

	// Instance reports a display name on first registration.
	assert.NoError(t, r.Put(&Node{Node: cluster.Node{Name: "pp-dn", Role: cluster.RoleInstance, DisplayName: "First Title"}}))
	got, err := r.FindByName("pp-dn")
	assert.NoError(t, err)
	if !assert.NotNil(t, got) {
		return
	}
	assert.Equal(t, "First Title", got.DisplayName)

	// A later instance registration updates the reported name (no admin pin yet).
	assert.NoError(t, r.Put(&Node{Node: cluster.Node{ClientID: got.ClientID, Name: got.Name, DisplayName: "Second Title"}}))
	got, _ = r.FindByName("pp-dn")
	assert.Equal(t, "Second Title", got.DisplayName)

	// Admin override (SrcManual) pins the value.
	assert.NoError(t, r.Put(&Node{Node: cluster.Node{ClientID: got.ClientID, Name: got.Name, DisplayName: "Admin Pinned"}, NameSrc: entity.SrcManual}))
	got, _ = r.FindByName("pp-dn")
	assert.Equal(t, "Admin Pinned", got.DisplayName)

	// A subsequent instance registration must NOT overwrite the admin override.
	assert.NoError(t, r.Put(&Node{Node: cluster.Node{ClientID: got.ClientID, Name: got.Name, DisplayName: "Instance Again"}}))
	got, _ = r.FindByName("pp-dn")
	assert.Equal(t, "Admin Pinned", got.DisplayName)

	// Admin clears the override (empty + SrcManual): un-pins and falls back, then
	// the next instance registration repopulates it.
	assert.NoError(t, r.Put(&Node{Node: cluster.Node{ClientID: got.ClientID, Name: got.Name, DisplayName: ""}, NameSrc: entity.SrcManual}))
	got, _ = r.FindByName("pp-dn")
	assert.Equal(t, "", got.DisplayName)
	assert.NoError(t, r.Put(&Node{Node: cluster.Node{ClientID: got.ClientID, Name: got.Name, DisplayName: "After Clear"}}))
	got, _ = r.FindByName("pp-dn")
	assert.Equal(t, "After Clear", got.DisplayName)
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
