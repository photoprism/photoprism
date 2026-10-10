package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// migrateFailEntity is a model whose table cannot be created, so that a schema migration fails.
type migrateFailEntity struct {
	ID   uint   `gorm:"primary_key"`
	Name string `gorm:"type:BROKEN((("`
}

// TableName returns the entity table name.
func (migrateFailEntity) TableName() string {
	return "config_migrate_fail"
}

// TestConfig_MigrateDbVersion pins that a version is recorded as migrated only after its schema migration succeeded.
func TestConfig_MigrateDbVersion(t *testing.T) {
	if TestConfig().DatabaseDriver() != dsn.DriverSQLite3 {
		t.Skip("requires a SQLite database of its own")
	}

	c := NewMinimalTestConfigWithDb("migrate-version", t.TempDir())

	t.Cleanup(func() {
		delete(entity.Entities, "config_migrate_fail")
		_ = c.CloseDb()
		entity.SetDbProvider(TestConfig())
		TestConfig().Propagate()
	})

	// An unknown version is never recorded, so give the build one.
	c.options.Version = "261009-test"
	require.NoError(t, c.MigrateDb(false, nil))

	version := migrate.FirstOrCreateVersion(c.Db(), migrate.NewVersion(c.Version(), c.Edition()))
	require.NotNil(t, version)
	require.False(t, version.NeedsMigration())

	// Migrate the schema again, as "migrations run --failed" does, with a table that cannot be created.
	entity.Entities["config_migrate_fail"] = &migrateFailEntity{}
	assert.Error(t, c.MigrateDb(true, nil))

	version = migrate.FirstOrCreateVersion(c.Db(), migrate.NewVersion(c.Version(), c.Edition()))
	require.NotNil(t, version)
	assert.True(t, version.NeedsMigration(), "a failed migration is retried on the next start")
	assert.NotEmpty(t, version.Error)

	// Running selected migrations does not migrate the schema, so the version stays unrecorded.
	delete(entity.Entities, "config_migrate_fail")
	require.NoError(t, c.MigrateDb(false, []string{"20221015-100000"}))

	version = migrate.FirstOrCreateVersion(c.Db(), migrate.NewVersion(c.Version(), c.Edition()))
	require.NotNil(t, version)
	assert.True(t, version.NeedsMigration())

	require.NoError(t, c.MigrateDb(false, nil))

	version = migrate.FirstOrCreateVersion(c.Db(), migrate.NewVersion(c.Version(), c.Edition()))
	require.NotNil(t, version)
	assert.False(t, version.NeedsMigration())
	assert.Empty(t, version.Error)

	// A failure under an unknown version records nothing, as a success would.
	c.options.Version = ""
	entity.Entities["config_migrate_fail"] = &migrateFailEntity{}
	assert.Error(t, c.MigrateDb(true, nil))
	assert.Empty(t, migrate.UnknownVersion.Error)
	assert.Nil(t, migrate.UnknownVersion.Find(c.Db()))
}

// TestConfig_versionEdition checks that Portal builds record schema versions under an edition of their own.
func TestConfig_versionEdition(t *testing.T) {
	prevRole := DefaultNodeRole
	t.Cleanup(func() { DefaultNodeRole = prevRole })

	c := NewMinimalTestConfig(t.TempDir())

	t.Run("Community", func(t *testing.T) {
		DefaultNodeRole = cluster.RoleInstance
		c.options.Edition = ""
		assert.Equal(t, "ce", c.versionEdition())
	})
	t.Run("Instance", func(t *testing.T) {
		DefaultNodeRole = cluster.RoleInstance
		c.options.Edition = "ultimate"
		assert.Equal(t, "ultimate", c.versionEdition())
	})
	t.Run("Portal", func(t *testing.T) {
		DefaultNodeRole = cluster.RolePortal
		c.options.Edition = "ultimate"
		assert.Equal(t, "portal-ultimate", c.versionEdition())
	})
	t.Run("PortalEdition", func(t *testing.T) {
		DefaultNodeRole = cluster.RoleInstance
		c.options.Edition = Portal
		assert.Equal(t, Portal, c.versionEdition())
	})
}
