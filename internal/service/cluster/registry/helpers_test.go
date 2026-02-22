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
