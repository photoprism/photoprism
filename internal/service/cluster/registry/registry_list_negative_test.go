package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Ensure List() excludes clients that look like nodes by role but have no NodeUUID.
func TestClientRegistry_ListExcludesNodeRoleWithoutUUID(t *testing.T) {
	c := newRegistryTestConfig(t, "cluster-registry-list-exclude-node-role")

	// Bad records: node-like roles but empty NodeUUID
	bad1 := entity.NewClient().SetName("pp-bad1").SetRole(cluster.RoleInstance)
	assert.NoError(t, bad1.Create())
	bad2 := entity.NewClient().SetName("pp-bad2").SetRole(cluster.RoleService)
	assert.NoError(t, bad2.Create())

	// Good record: proper NodeUUID
	good := entity.NewClient().SetName("pp-good").SetRole(cluster.RoleInstance)
	good.NodeUUID = rnd.UUIDv7()
	assert.NoError(t, good.Create())

	r, _ := NewClientRegistryWithConfig(c)
	list, err := r.List()
	assert.NoError(t, err)

	// The MariaDB test database is shared across tests (unlike the per-test SQLite
	// file), so List() also returns nodes created elsewhere. Assert on membership
	// rather than the exact count: the UUID-backed node is listed and the node-role
	// records without a NodeUUID are not.
	if found := listNodeByName(list, "pp-good"); assert.NotNil(t, found, "UUID-backed node should be listed") {
		assert.NotEmpty(t, found.UUID)
	}

	assert.Nil(t, listNodeByName(list, "pp-bad1"), "node role without UUID must be excluded")
	assert.Nil(t, listNodeByName(list, "pp-bad2"), "node role without UUID must be excluded")
}
