package registry

import (
	"testing"

	cfg "github.com/photoprism/photoprism/internal/config"
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
