package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cfg "github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// newRegistryTestConfig creates a minimal registry test config and closes its database during cleanup.
func newRegistryTestConfig(t *testing.T, name string) *cfg.Config {
	t.Helper()

	c := cfg.NewMinimalTestConfigWithDb(name, t.TempDir())
	t.Cleanup(func() {
		if err := c.CloseDb(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})

	return c
}

// backdateClient moves the client's updated_at into the past, so a later write sorts after it although
// timestamps have only second precision.
func backdateClient(t *testing.T, clientUID string, age time.Duration) {
	t.Helper()

	require.NoError(t, entity.UnscopedDb().Model(&entity.Client{}).Where("client_uid = ?", clientUID).
		UpdateColumn("updated_at", time.Now().UTC().Add(-age)).Error)
}

// listNodeByName returns the listed node with the given name, or nil if absent.
// List queries are not isolated per test on the shared MariaDB test database, so
// tests assert on membership with this helper rather than an exact result count.
func listNodeByName(list []Node, name string) *Node {
	for i := range list {
		if list[i].Name == name {
			return &list[i]
		}
	}

	return nil
}
