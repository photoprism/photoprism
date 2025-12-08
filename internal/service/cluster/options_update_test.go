package cluster_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	clusternode "github.com/photoprism/photoprism/internal/service/cluster/node"
)

func TestOptionsUpdate_IsZero(t *testing.T) {
	var u cluster.OptionsUpdate
	assert.True(t, u.IsZero())

	u.SetClusterUUID("1234")
	assert.False(t, u.IsZero())
}

func TestOptionsUpdate_HasDatabaseUpdate(t *testing.T) {
	var u cluster.OptionsUpdate
	assert.False(t, u.HasDatabaseUpdate())

	u.SetDatabaseName("photoprism")
	assert.True(t, u.HasDatabaseUpdate())
}

func TestOptionsUpdate_Apply(t *testing.T) {
	conf := config.NewMinimalTestConfig(t.TempDir())
	conf.Options().OptionsYaml = filepath.Join(conf.ConfigPath(), "options.yml")

	// Seed file with existing values to ensure they are preserved.
	seed := map[string]any{"Existing": "value"}
	b, err := yaml.Marshal(seed)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(conf.OptionsYaml(), b, 0o600))

	update := cluster.OptionsUpdate{}
	update.SetClusterUUID("4a47c940-d5de-41b3-88a2-eb816cc659ca")
	update.SetClusterCIDR("192.0.2.0/24")
	update.SetDatabaseName("cluster_database")
	update.SetDatabaseUser("cluster_user")

	written, err := clusternode.ApplyOptionsUpdate(conf, update)
	require.NoError(t, err)
	assert.True(t, written)

	content, err := os.ReadFile(conf.OptionsYaml())
	require.NoError(t, err)

	var merged map[string]any
	require.NoError(t, yaml.Unmarshal(content, &merged))

	assert.Equal(t, "value", merged["Existing"])
	assert.Equal(t, "4a47c940-d5de-41b3-88a2-eb816cc659ca", merged["ClusterUUID"])
	assert.Equal(t, "192.0.2.0/24", merged["ClusterCIDR"])
	assert.Equal(t, "cluster_database", merged["DatabaseName"])
	assert.Equal(t, "cluster_user", merged["DatabaseUser"])

	// Applying an empty update should be a no-op.
	empty := cluster.OptionsUpdate{}
	written, err = clusternode.ApplyOptionsUpdate(conf, empty)
	require.NoError(t, err)
	assert.False(t, written)
}
