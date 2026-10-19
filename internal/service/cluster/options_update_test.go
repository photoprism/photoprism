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

func TestOptionsUpdate_ClearDatabase(t *testing.T) {
	var u cluster.OptionsUpdate
	u.SetNodeClientID(cluster.ExampleClientID)
	u.SetDatabaseDriver("mysql")
	u.SetDatabaseDSN("user:pw@tcp(db:3306)/photoprism")
	u.SetDatabaseServer("db:3306")
	u.SetDatabaseName("photoprism")
	u.SetDatabaseUser("user")
	u.SetDatabasePassword("pw")
	u.ClearDatabase()
	assert.False(t, u.HasDatabaseUpdate())
	assert.False(t, u.IsZero())
	require.NotNil(t, u.NodeClientID)
}

func TestOptionsUpdate_SetDatabase(t *testing.T) {
	complete := cluster.RegisterDatabase{Host: "db.local", Port: 3306, Name: "pp_db", User: "pp_user", Password: "pp_pw"}
	t.Run("Success", func(t *testing.T) {
		var u cluster.OptionsUpdate
		server, ok := u.SetDatabase(complete)
		assert.True(t, ok)
		assert.Equal(t, "db.local:3306", server)
		require.NotNil(t, u.DatabaseDriver)
		assert.Equal(t, "mysql", *u.DatabaseDriver)
		assert.Equal(t, "db.local:3306", *u.DatabaseServer)
		assert.Equal(t, "pp_db", *u.DatabaseName)
		assert.Equal(t, "pp_user", *u.DatabaseUser)
		assert.Equal(t, "pp_pw", *u.DatabasePassword)
		assert.Nil(t, u.DatabaseDSN)
	})
	t.Run("Driver", func(t *testing.T) {
		var u cluster.OptionsUpdate
		d := complete
		d.Port, d.Driver = 0, "mariadb"
		server, ok := u.SetDatabase(d)
		assert.True(t, ok)
		assert.Equal(t, "db.local", server)
		assert.Equal(t, "db.local", *u.DatabaseServer)
		assert.Equal(t, "mariadb", *u.DatabaseDriver)
		assert.Equal(t, "pp_pw", *u.DatabasePassword)
	})
	t.Run("NoHost", func(t *testing.T) {
		for _, port := range []int{3306, 0} {
			var u cluster.OptionsUpdate
			d := complete
			d.Host, d.Port = "", port
			_, ok := u.SetDatabase(d)
			assert.False(t, ok, port)
			assert.True(t, u.IsZero(), port)
		}
	})
	t.Run("UnusableServer", func(t *testing.T) {
		var u cluster.OptionsUpdate
		d := complete
		d.Host = "/run/mysqld/mysqld.sock"
		server, ok := u.SetDatabase(d)
		assert.False(t, ok)
		assert.Equal(t, "/run/mysqld/mysqld.sock:3306", server)
		assert.True(t, u.IsZero())
	})
	t.Run("NotComplete", func(t *testing.T) {
		for name, change := range map[string]func(*cluster.RegisterDatabase){
			"Name":       func(d *cluster.RegisterDatabase) { d.Name = "" },
			"User":       func(d *cluster.RegisterDatabase) { d.User = "" },
			"Password":   func(d *cluster.RegisterDatabase) { d.Password = "" },
			"Incomplete": func(d *cluster.RegisterDatabase) { d.Incomplete = true },
		} {
			var u cluster.OptionsUpdate
			d := complete
			d.Host = "/run/mysqld/mysqld.sock"
			d.DSN = "pp_user:pp_pw@tcp(db.local:3306)/pp_db"
			change(&d)
			_, ok := u.SetDatabase(d)
			assert.True(t, ok, name)
			assert.True(t, u.IsZero(), name)
		}
	})
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
	update.SetPortalLoginUrl("https://portal.example.com/portal/login")

	written, err := conf.SaveClusterOptionsUpdate(update)
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
	assert.Equal(t, "https://portal.example.com/portal/login", merged["PortalLoginUrl"])

	// Applying an empty update should be a no-op.
	empty := cluster.OptionsUpdate{}
	written, err = conf.SaveClusterOptionsUpdate(empty)
	require.NoError(t, err)
	assert.False(t, written)
}
