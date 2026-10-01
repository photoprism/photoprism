package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// TestPersistRegisterResponse checks which registration settings are saved to options.yml.
func TestPersistRegisterResponse(t *testing.T) {
	t.Run("UnusableServer", func(t *testing.T) {
		// Database settings with a server address the instance cannot use are ignored as a whole, and the
		// cluster settings are still saved.
		clusterUUID := "4a47c940-d5de-41b3-88a2-eb816cc659ca"

		for _, host := range []string{"/run/mysqld/mysqld.sock", ""} {
			conf := config.NewMinimalTestConfig(t.TempDir())
			require.NoError(t, persistRegisterResponse(conf, &cluster.RegisterResponse{
				UUID: clusterUUID,
				Node: cluster.Node{ClientID: cluster.ExampleClientID},
				Database: cluster.RegisterDatabase{
					Driver:   dsn.DriverMySQL,
					Host:     host,
					Port:     3306,
					Name:     "pp_db",
					User:     "pp_user",
					Password: "pwd",
				},
			}), host)

			content, err := os.ReadFile(conf.OptionsYaml())
			require.NoError(t, err)

			var persisted map[string]any
			require.NoError(t, yaml.Unmarshal(content, &persisted))
			assert.Equal(t, clusterUUID, persisted["ClusterUUID"], host)
			assert.Equal(t, cluster.ExampleClientID, persisted["NodeClientID"], host)

			if host == "" {
				assert.Equal(t, "pp_db", persisted["DatabaseName"])
				assert.Equal(t, ":3306", persisted["DatabaseServer"])
			} else {
				for _, key := range []string{"DatabaseDriver", "DatabaseName", "DatabaseUser", "DatabasePassword", "DatabaseServer"} {
					assert.NotContains(t, persisted, key)
				}
			}
		}
	})
}
