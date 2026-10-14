package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// TestConfig_ClientUserListShape pins how the list fields of the client config encode in an empty library.
// Lenses and countries are omitted because their default rows are always present. Album categories and
// years stay null, as GORM v1 leaves a plucked slice nil when nothing matches.
func TestConfig_ClientUserListShape(t *testing.T) {
	c := NewIsolatedTestConfig("client-list-shape", t.TempDir(), true)

	// Only SQLite gets a database of its own here; on MariaDB the package database holds fixtures.
	if c.DatabaseDriver() != dsn.DriverSQLite3 {
		t.Skip("requires an empty SQLite database")
	}

	t.Cleanup(func() {
		_ = c.CloseDb()
		entity.SetDbProvider(TestConfig())
		TestConfig().Propagate()
	})
	require.NoError(t, c.Init())
	c.InitDb()

	data, err := json.Marshal(c.ClientUser(false))
	require.NoError(t, err)

	fields := map[string]json.RawMessage{}
	require.NoError(t, json.Unmarshal(data, &fields))

	expected := map[string]string{
		"albumCategories": "null",
		"cameras":         "[]",
		"years":           "null",
		"categories":      "[]",
	}

	for name, value := range expected {
		assert.JSONEq(t, value, string(fields[name]), name)
	}
}
