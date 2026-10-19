package commands

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// TestPersistRegisterResponse checks which registration settings are saved to options.yml.
func TestPersistRegisterResponse(t *testing.T) {
	t.Run("UnusableServer", func(t *testing.T) {
		// Database settings with a server address the instance cannot use, or without a host, are ignored as
		// a whole, and the cluster settings are still saved.
		clusterUUID := "4a47c940-d5de-41b3-88a2-eb816cc659ca"

		logger, ok := log.(*logrus.Logger)
		require.True(t, ok)
		hook := test.NewLocal(logger)
		system := event.SystemLog
		event.SystemLog = logger
		t.Cleanup(func() {
			event.SystemLog = system
			hook.Reset()
		})

		for _, host := range []string{"/run/mysqld/mysqld.sock", ""} {
			hook.Reset()
			conf := config.NewMinimalTestConfig(t.TempDir())
			refused, err := persistRegisterResponse(conf, &cluster.RegisterResponse{
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
			})
			require.NoError(t, err, host)
			assert.True(t, refused, host)

			content, err := os.ReadFile(conf.OptionsYaml())
			require.NoError(t, err)

			var persisted map[string]any
			require.NoError(t, yaml.Unmarshal(content, &persisted))
			assert.Equal(t, clusterUUID, persisted["ClusterUUID"], host)
			assert.Equal(t, cluster.ExampleClientID, persisted["NodeClientID"], host)

			var messages []string
			for _, entry := range hook.AllEntries() {
				messages = append(messages, entry.Message)
			}
			assert.Contains(t, messages, "cluster: ignored database settings with unusable server address "+host+":3306")
			for _, key := range []string{"DatabaseDriver", "DatabaseName", "DatabaseUser", "DatabasePassword", "DatabaseServer"} {
				assert.NotContains(t, persisted, key, host)
			}
		}
	})
	t.Run("NoPassword", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		refused, err := persistRegisterResponse(conf, &cluster.RegisterResponse{
			UUID:     "4a47c940-d5de-41b3-88a2-eb816cc659ca",
			Node:     cluster.Node{ClientID: cluster.ExampleClientID},
			Database: cluster.RegisterDatabase{Driver: dsn.DriverMySQL, Host: "db.local", Port: 3306, Name: "pp_db", User: "pp_user"},
		})
		require.NoError(t, err)
		assert.False(t, refused)

		content, err := os.ReadFile(conf.OptionsYaml())
		require.NoError(t, err)

		var persisted map[string]any
		require.NoError(t, yaml.Unmarshal(content, &persisted))

		for _, key := range []string{"DatabaseDriver", "DatabaseName", "DatabaseUser", "DatabasePassword", "DatabaseServer"} {
			assert.NotContains(t, persisted, key)
		}
	})
	t.Run("IncompleteDatabase", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		refused, err := persistRegisterResponse(conf, &cluster.RegisterResponse{
			UUID: "4a47c940-d5de-41b3-88a2-eb816cc659ca",
			Node: cluster.Node{ClientID: cluster.ExampleClientID},
			Database: cluster.RegisterDatabase{Driver: dsn.DriverMySQL, Host: "db.local", Port: 3306, Name: "pp_db", User: "pp_user",
				Password: "pp_pw", DSN: "pp_user:pp_pw@tcp(db.local:3306)/pp_db?parseTime=true", Incomplete: true},
		})
		require.NoError(t, err)
		assert.False(t, refused)

		content, err := os.ReadFile(conf.OptionsYaml())
		require.NoError(t, err)

		var persisted map[string]any
		require.NoError(t, yaml.Unmarshal(content, &persisted))
		assert.Equal(t, cluster.ExampleClientID, persisted["NodeClientID"])

		for _, key := range []string{"DatabaseDriver", "DatabaseDSN", "DatabaseName", "DatabaseUser", "DatabasePassword", "DatabaseServer"} {
			assert.NotContains(t, persisted, key)
		}
	})
	t.Run("PortalDSN", func(t *testing.T) {
		// The Portal sends a DSN with every password; the separate settings are saved, so that a socket
		// address in the DSN cannot refuse the whole update.
		for _, host := range []string{"db.local", "/run/mysqld/mysqld.sock"} {
			conf := config.NewMinimalTestConfig(t.TempDir())
			refused, err := persistRegisterResponse(conf, &cluster.RegisterResponse{
				UUID: "4a47c940-d5de-41b3-88a2-eb816cc659ca",
				Node: cluster.Node{ClientID: cluster.ExampleClientID},
				Database: cluster.RegisterDatabase{Driver: dsn.DriverMySQL, Host: host, Port: 3306, Name: "pp_db", User: "pp_user",
					Password: "pp_pw", DSN: "pp_user:pp_pw@tcp(" + host + ":3306)/pp_db?parseTime=true"},
			})
			require.NoError(t, err, host)
			assert.Equal(t, host != "db.local", refused, host)

			content, err := os.ReadFile(conf.OptionsYaml())
			require.NoError(t, err)

			var persisted map[string]any
			require.NoError(t, yaml.Unmarshal(content, &persisted))
			assert.Equal(t, cluster.ExampleClientID, persisted["NodeClientID"], host)
			assert.NotContains(t, persisted, "DatabaseDSN", host)

			if host == "db.local" {
				assert.Equal(t, "db.local:3306", persisted["DatabaseServer"])
				assert.Equal(t, "pp_pw", persisted["DatabasePassword"])
			} else {
				assert.NotContains(t, persisted, "DatabasePassword", host)
			}
		}
	})
	t.Run("NotSaved", func(t *testing.T) {
		// Credentials of a complete response that cannot be saved are reported as refused.
		logger, ok := log.(*logrus.Logger)
		require.True(t, ok)
		hook := test.NewLocal(logger)
		system := event.SystemLog
		event.SystemLog = logger
		t.Cleanup(func() {
			event.SystemLog = system
			hook.Reset()
		})

		for name, db := range map[string]cluster.RegisterDatabase{
			"NoName":  {Driver: dsn.DriverMySQL, Host: "db.local", Port: 3306, User: "pp_user", Password: "pp_pw"},
			"NoUser":  {Driver: dsn.DriverMySQL, Host: "db.local", Port: 3306, Name: "pp_db", Password: "pp_pw"},
			"DSNOnly": {Driver: dsn.DriverMySQL, DSN: "pp_user:pp_pw@tcp(db.local:3306)/pp_db?parseTime=true"},
			"NoHost":  {Driver: dsn.DriverMySQL, Host: "[]", Port: 3306, Name: "pp_db", User: "pp_user", Password: "pp_pw"},
		} {
			conf := config.NewMinimalTestConfig(t.TempDir())
			refused, err := persistRegisterResponse(conf, &cluster.RegisterResponse{
				UUID:     "4a47c940-d5de-41b3-88a2-eb816cc659ca",
				Node:     cluster.Node{ClientID: cluster.ExampleClientID},
				Database: db,
			})
			require.NoError(t, err, name)
			assert.True(t, refused, name)

			if name != "NoHost" {
				var messages []string
				for _, entry := range hook.AllEntries() {
					messages = append(messages, entry.Message)
				}
				assert.Contains(t, messages, "cluster: ignored database settings without name, user, or password", name)
			}

			hook.Reset()

			content, err := os.ReadFile(conf.OptionsYaml())
			require.NoError(t, err, name)

			var persisted map[string]any
			require.NoError(t, yaml.Unmarshal(content, &persisted), name)
			assert.Equal(t, cluster.ExampleClientID, persisted["NodeClientID"], name)

			for _, key := range []string{"DatabaseDriver", "DatabaseDSN", "DatabaseName", "DatabaseUser", "DatabasePassword", "DatabaseServer"} {
				assert.NotContains(t, persisted, key, name)
			}
		}
	})
	t.Run("RefusedDatabase", func(t *testing.T) {
		// Database settings the config refuses are skipped as a whole, and the registration is saved.
		conf := config.NewMinimalTestConfig(t.TempDir())
		refused, err := persistRegisterResponse(conf, &cluster.RegisterResponse{
			UUID: "4a47c940-d5de-41b3-88a2-eb816cc659ca",
			Node: cluster.Node{ClientID: cluster.ExampleClientID},
			Database: cluster.RegisterDatabase{Driver: dsn.DriverMySQL, Host: "db.local", Port: 3306, Name: "db?x", User: "pp_user",
				Password: "pp_pw"},
		})
		require.NoError(t, err)
		assert.True(t, refused)

		content, err := os.ReadFile(conf.OptionsYaml())
		require.NoError(t, err)

		var persisted map[string]any
		require.NoError(t, yaml.Unmarshal(content, &persisted))
		assert.Equal(t, cluster.ExampleClientID, persisted["NodeClientID"])

		for _, key := range []string{"DatabaseDriver", "DatabaseName", "DatabaseUser", "DatabasePassword", "DatabaseServer"} {
			assert.NotContains(t, persisted, key)
		}
	})
	t.Run("ReplacesStoredDSN", func(t *testing.T) {
		// Rotated credentials saved as separate settings remove the DSN of an earlier join, which would
		// otherwise take precedence over them.
		conf := config.NewMinimalTestConfig(t.TempDir())
		_, err := conf.SaveOptionsPatch(config.Values{
			"DatabaseDriver": dsn.DriverMySQL,
			"DatabaseDSN":    "pp_user:old_pw@tcp(db.local:3306)/pp_db?parseTime=true",
		})
		require.NoError(t, err)

		refused, err := persistRegisterResponse(conf, &cluster.RegisterResponse{
			UUID: "4a47c940-d5de-41b3-88a2-eb816cc659ca",
			Node: cluster.Node{ClientID: cluster.ExampleClientID},
			Database: cluster.RegisterDatabase{Driver: dsn.DriverMySQL, Host: "db.local", Port: 3306, Name: "pp_db", User: "pp_user",
				Password: "new_pw"},
		})
		require.NoError(t, err)
		assert.False(t, refused)

		content, err := os.ReadFile(conf.OptionsYaml())
		require.NoError(t, err)

		var persisted map[string]any
		require.NoError(t, yaml.Unmarshal(content, &persisted))
		assert.NotContains(t, persisted, "DatabaseDSN")
		assert.Equal(t, "new_pw", persisted["DatabasePassword"])
		assert.Equal(t, "db.local:3306", persisted["DatabaseServer"])
		assert.Equal(t, "", conf.Options().DatabaseDSN)
		assert.Equal(t, "new_pw", conf.DatabasePassword())
		assert.Contains(t, conf.DatabaseDSN(), "pp_user:new_pw@tcp(db.local:3306)/pp_db")
	})
}
