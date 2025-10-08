package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Ensure List() excludes clients that look like nodes by role but have no NodeUUID.
func TestClientRegistry_ListExcludesNodeRoleWithoutUUID(t *testing.T) {
	c := cfg.NewMinimalTestConfigWithDb("cluster-registry-list-exclude-node-role", t.TempDir())
	defer c.CloseDb()

	// Bad records: node-like roles but empty NodeUUID
	bad1 := entity.NewClient().SetName("pp-bad1").SetRole("instance")
	assert.NoError(t, bad1.Create())
	bad2 := entity.NewClient().SetName("pp-bad2").SetRole("service")
	assert.NoError(t, bad2.Create())

	// Good record: proper NodeUUID
	good := entity.NewClient().SetName("pp-good").SetRole("instance")
	good.NodeUUID = rnd.UUIDv7()
	assert.NoError(t, good.Create())

	r, _ := NewClientRegistryWithConfig(c)
	list, err := r.List()
	assert.NoError(t, err)

	// Only the UUID-backed record should be present
	if assert.Equal(t, 1, len(list)) {
		assert.Equal(t, "pp-good", list[0].Name)
		assert.NotEmpty(t, list[0].UUID)
		assert.NotEqual(t, "pp-bad1", list[0].Name)
		assert.NotEqual(t, "pp-bad2", list[0].Name)
	}
}
