package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/dsn"

	"github.com/photoprism/photoprism/internal/service/cluster"
)

// resetDatabaseOptions clears all DB-related option fields so tests start from defaults even if
// storage/testdata/config/options.yml contains legacy values such as DatabaseDsn.
func resetDatabaseOptions(c *Config) {
	c.options.DatabaseDriver = ""
	c.options.DatabaseDSN = ""
	c.options.Deprecated.DatabaseDsn = ""
	c.options.DatabaseServer = ""
	c.options.DatabaseName = ""
	c.options.DatabaseUser = ""
	c.options.DatabasePassword = ""
}

func TestConfig_DatabaseDriver(t *testing.T) {
	t.Run("DefaultsToSQLite", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)

		assert.Equal(t, dsn.DriverSQLite3, c.DatabaseDriver())
	})
	t.Run("NormalizesDeprecatedDSN", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)

		c.options.DatabaseDriver = dsn.DriverMySQL
		c.options.Deprecated.DatabaseDsn = "user:pass@tcp(localhost:3306)/photoprism"

		assert.Equal(t, dsn.DriverMySQL, c.DatabaseDriver())
		assert.Equal(t, "user:pass@tcp(localhost:3306)/photoprism", c.options.DatabaseDSN)
		assert.Empty(t, c.options.Deprecated.DatabaseDsn)
	})
}

func TestConfig_DatabaseDriverName(t *testing.T) {
	t.Run("DefaultsToSQLite", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		assert.Equal(t, "SQLite", c.DatabaseDriverName())
	})
	t.Run("Whitespace", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = "   "
		assert.Equal(t, "SQLite", c.DatabaseDriverName())
	})
	t.Run("Auto", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = " AUTO "
		assert.Equal(t, "SQLite", c.DatabaseDriverName())
	})
	t.Run("MySQLReportsAsMariaDB", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = dsn.DriverMySQL
		assert.Equal(t, "MariaDB", c.DatabaseDriverName())
	})
	t.Run("MariaDBAliasReportsAsMariaDB", func(t *testing.T) {
		// "mariadb" collapses onto DriverMySQL in ParseDriver; format stays "MariaDB".
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = dsn.DriverMariaDB
		assert.Equal(t, "MariaDB", c.DatabaseDriverName())
	})
	t.Run("PostgresReportsAsPostgreSQL", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = dsn.DriverPostgres
		assert.Equal(t, "PostgreSQL", c.DatabaseDriverName())
	})
	t.Run("DeprecatedTiDBReportsAsSQLite", func(t *testing.T) {
		// DatabaseDriver() warns and rewrites "tidb" to SQLite3 before this runs.
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = "tidb"
		assert.Equal(t, "SQLite", c.DatabaseDriverName())
	})
	t.Run("UnknownDriverReportsAsSQLite", func(t *testing.T) {
		// DatabaseDriver() coerces unknown drivers to SQLite3 before this runs.
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = "oracle"
		assert.Equal(t, "SQLite", c.DatabaseDriverName())
	})
}

func TestConfig_DatabaseVersion(t *testing.T) {
	c := TestConfig()

	assert.NotEmpty(t, c.DatabaseVersion())
	assert.True(t, c.IsDatabaseVersion("v3.45"))
}

func TestConfig_DatabaseSsl(t *testing.T) {
	c := TestConfig()

	assert.False(t, c.DatabaseSsl())
}

func TestConfig_normalizeDatabaseDSN(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.Deprecated.DatabaseDsn = "foo:b@r@tcp(honeypot:1234)/baz?charset=utf8mb4,utf8&parseTime=true"
	c.options.DatabaseDriver = dsn.DriverMySQL

	assert.Equal(t, "honeypot:1234", c.DatabaseServer())
	assert.Equal(t, "honeypot", c.DatabaseHost())
	assert.Equal(t, 1234, c.DatabasePort())
	assert.Equal(t, "baz", c.DatabaseName())
	assert.Equal(t, "foo", c.DatabaseUser())
	assert.Equal(t, "b@r", c.DatabasePassword())
}

func TestConfig_ParseDatabaseDSN(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.DatabaseDSN = "foo:b@r@tcp(honeypot:1234)/baz?charset=utf8mb4,utf8&parseTime=true"
	c.options.DatabaseDriver = dsn.DriverSQLite3

	assert.Equal(t, "", c.DatabaseServer())
	assert.Equal(t, "", c.DatabaseHost())
	assert.Equal(t, 0, c.DatabasePort())
	assert.Equal(t, "foo:b@r@tcp(honeypot:1234)/baz?charset=utf8mb4,utf8&parseTime=true", c.DatabaseName())
	assert.Equal(t, "", c.DatabaseUser())
	assert.Equal(t, "", c.DatabasePassword())

	c.options.DatabaseDriver = dsn.DriverMySQL

	assert.Equal(t, "honeypot:1234", c.DatabaseServer())
	assert.Equal(t, "honeypot", c.DatabaseHost())
	assert.Equal(t, 1234, c.DatabasePort())
	assert.Equal(t, "baz", c.DatabaseName())
	assert.Equal(t, "foo", c.DatabaseUser())
	assert.Equal(t, "b@r", c.DatabasePassword())

	c.options.DatabaseDriver = dsn.DriverSQLite3

	assert.Equal(t, "", c.DatabaseServer())
	assert.Equal(t, "", c.DatabaseHost())
	assert.Equal(t, 0, c.DatabasePort())
	assert.Equal(t, "foo:b@r@tcp(honeypot:1234)/baz?charset=utf8mb4,utf8&parseTime=true", c.DatabaseName())
	assert.Equal(t, "", c.DatabaseUser())
	assert.Equal(t, "", c.DatabasePassword())

	t.Run("ManualServerConfig", func(t *testing.T) {
		target := NewConfig(CliTestContext())
		resetDatabaseOptions(target)

		target.options.DatabaseDriver = dsn.DriverMySQL
		target.options.DatabaseServer = "db.internal:3306"
		target.options.DatabaseName = "photoprism"
		target.options.DatabaseUser = "app"
		target.options.DatabasePassword = "secret"
		target.options.DatabaseDSN = "foo:b@r@tcp(otherhost:3307)/other?charset=utf8mb4,utf8&parseTime=true"

		target.ParseDatabaseDSN()

		assert.Equal(t, "otherhost:3307", target.options.DatabaseServer)
		assert.Equal(t, "otherhost", target.DatabaseHost())
		assert.Equal(t, "other", target.options.DatabaseName)
		assert.Equal(t, "foo", target.options.DatabaseUser)
		assert.Equal(t, "b@r", target.options.DatabasePassword)
	})
	t.Run("SQLiteSkipWhenServerPreset", func(t *testing.T) {
		cfg := NewConfig(CliTestContext())
		resetDatabaseOptions(cfg)

		cfg.options.DatabaseDriver = dsn.DriverSQLite3
		cfg.options.DatabaseDSN = "file:/data/app.db?_busy_timeout=5000"
		cfg.options.DatabaseServer = "/tmp/mysql.sock"
		cfg.options.DatabaseName = "existing-name"
		cfg.options.DatabaseUser = "existing-user"
		cfg.options.DatabasePassword = "existing-pass"

		cfg.ParseDatabaseDSN()

		assert.Equal(t, "/tmp/mysql.sock", cfg.options.DatabaseServer)
		assert.Equal(t, "existing-name", cfg.options.DatabaseName)
		assert.Equal(t, "existing-user", cfg.options.DatabaseUser)
		assert.Equal(t, "existing-pass", cfg.options.DatabasePassword)
	})
}

func TestConfig_DatabaseServer(t *testing.T) {
	c := NewConfig(CliTestContext())
	resetDatabaseOptions(c)
	assert.Equal(t, "", c.DatabaseServer())
	c.options.DatabaseServer = "test"
	assert.Equal(t, "", c.DatabaseServer())
}

func TestConfig_DatabaseHost(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		assert.Equal(t, "", c.DatabaseHost())
	})
	t.Run("PostgresDefaultsToLocalhost", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = dsn.DriverPostgres
		assert.Equal(t, "localhost", c.DatabaseHost())
	})
	t.Run("PostgresUsesConfiguredHost", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = dsn.DriverPostgres
		c.options.DatabaseServer = "postgres.internal:5433"
		assert.Equal(t, "postgres.internal", c.DatabaseHost())
	})
}

func TestConfig_DatabasePort(t *testing.T) {
	t.Run("SQLite", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		assert.Equal(t, 0, c.DatabasePort())
	})
	t.Run("PostgresDefaultsTo5432", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = dsn.DriverPostgres
		assert.Equal(t, 5432, c.DatabasePort())
	})
	t.Run("PostgresUsesConfiguredPort", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		resetDatabaseOptions(c)
		c.options.DatabaseDriver = dsn.DriverPostgres
		c.options.DatabaseServer = "postgres.internal:5433"
		assert.Equal(t, 5433, c.DatabasePort())
	})
}

func TestConfig_DatabasePortString(t *testing.T) {
	c := NewConfig(CliTestContext())
	resetDatabaseOptions(c)
	assert.Equal(t, "", c.DatabasePortString())
}

func TestConfig_DatabaseName(t *testing.T) {
	c := NewConfig(CliTestContext())
	resetDatabaseOptions(c)
	assert.Equal(t, ProjectRoot+"/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseName())
}

func TestConfig_DatabaseUser(t *testing.T) {
	c := NewConfig(CliTestContext())
	resetDatabaseOptions(c)
	assert.Equal(t, "", c.DatabaseUser())
}

func TestConfig_DatabasePassword(t *testing.T) {
	c := NewConfig(CliTestContext())
	resetDatabaseOptions(c)
	assert.Equal(t, "", c.DatabasePassword())

	// Test setting the password via secret file.
	_ = os.Setenv(FlagFileVar("DATABASE_PASSWORD"), "testdata/secret_database")
	assert.Equal(t, "", c.DatabasePassword())
	c.Options().DatabaseDriver = dsn.DriverMySQL
	assert.Equal(t, "StoryOfAmélie", c.DatabasePassword())
	c.Options().DatabaseDriver = dsn.DriverSQLite3
	_ = os.Setenv(FlagFileVar("DATABASE_PASSWORD"), "")

	assert.Equal(t, "", c.DatabasePassword())
}

func TestDatabaseProvisionPrefix(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)
		assert.Equal(t, cluster.DefaultDatabaseProvisionPrefix, conf.DatabaseProvisionPrefix())
	})
	t.Run("SanitizeAndTrim", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)
		conf.options.DatabaseProvisionPrefix = "  My Custom-Prefix!!  "

		got := conf.DatabaseProvisionPrefix()

		assert.Equal(t, "my_custom_prefix", got)
		assert.LessOrEqual(t, len(got), cluster.DatabaseProvisionPrefixMaxLen)
		assert.Equal(t, got, conf.options.DatabaseProvisionPrefix)
	})
}

func TestShouldAutoRotateDatabase(t *testing.T) {
	t.Run("PortalAlwaysFalse", func(t *testing.T) {
		conf := NewMinimalTestConfig(t.TempDir())
		conf.Options().Edition = Portal
		conf.Options().NodeRole = cluster.RolePortal
		conf.Options().DatabaseDriver = dsn.DriverMySQL
		assert.False(t, conf.ShouldAutoRotateDatabase())
	})
	t.Run("NonMySQLDriverFalse", func(t *testing.T) {
		conf := NewMinimalTestConfig(t.TempDir())
		conf.Options().DatabaseDriver = dsn.DriverSQLite3
		assert.False(t, conf.ShouldAutoRotateDatabase())
	})
	t.Run("MySQLMissingFieldsTrue", func(t *testing.T) {
		conf := NewMinimalTestConfig(t.TempDir())
		conf.Options().DatabaseDriver = dsn.DriverMySQL
		conf.Options().DatabaseName = "photoprism"
		conf.Options().DatabaseUser = ""
		conf.Options().DatabasePassword = ""
		assert.True(t, conf.ShouldAutoRotateDatabase())
	})
}

func TestConfig_DatabaseDSN(t *testing.T) {
	c := NewConfig(CliTestContext())
	resetDatabaseOptions(c)
	driver := c.DatabaseDriver()
	assert.Equal(t, dsn.DriverSQLite3, driver)
	c.options.DatabaseDSN = ""
	c.options.DatabaseDriver = "MariaDB"
	assert.Equal(t, "photoprism:@tcp(localhost)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=15s", c.DatabaseDSN())
	c.options.DatabaseDriver = "tidb"
	assert.Equal(t, ProjectRoot+"/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseDSN())
	c.options.DatabaseDriver = "Postgres"
	assert.Equal(t, "user=photoprism password= dbname=photoprism host=localhost port=5432 connect_timeout=15 sslmode=disable TimeZone=UTC", c.DatabaseDSN())
	c.options.DatabaseDriver = "SQLite"
	assert.Equal(t, ProjectRoot+"/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseDSN())
	c.options.DatabaseDriver = ""
	assert.Equal(t, ProjectRoot+"/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseDSN())

	t.Run("CustomServer", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = dsn.DriverMySQL
		conf.options.DatabaseServer = "proxy.internal:6032"
		conf.options.DatabaseName = "instancedb"
		conf.options.DatabaseUser = "instance"
		conf.options.DatabasePassword = "secret"
		conf.options.DatabaseTimeout = 42

		want := "instance:secret@tcp(proxy.internal:6032)/instancedb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=42s"
		if got := conf.DatabaseDSN(); got != want {
			t.Fatalf("DatabaseDSN() = %q, want %q", got, want)
		}
	})
	t.Run("UnixSocket", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = dsn.DriverMySQL
		conf.options.DatabaseServer = "/var/run/mysql.sock"
		conf.options.DatabaseName = "instancedb"
		conf.options.DatabaseUser = "instance"
		conf.options.DatabasePassword = "secret"
		conf.options.DatabaseTimeout = 21

		want := "instance:secret@unix(/var/run/mysql.sock)/instancedb?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=21s"
		if got := conf.DatabaseDSN(); got != want {
			t.Fatalf("DatabaseDSN() = %q, want %q", got, want)
		}
	})
	t.Run("PostgresDefaults", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = dsn.DriverPostgres
		conf.options.DatabaseName = "instancedb"
		conf.options.DatabaseUser = "instance"
		conf.options.DatabasePassword = "secret"
		conf.options.DatabaseTimeout = 12

		want := "user=instance password=secret dbname=instancedb host=localhost port=5432 connect_timeout=12 sslmode=disable TimeZone=UTC"
		if got := conf.DatabaseDSN(); got != want {
			t.Fatalf("DatabaseDSN() = %q, want %q", got, want)
		}

		assert.Equal(t, "localhost", conf.DatabaseHost())
		assert.Equal(t, 5432, conf.DatabasePort())
	})
	t.Run("PostgresCustomServer", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = dsn.DriverPostgres
		conf.options.DatabaseServer = "postgres.internal:5433"
		conf.options.DatabaseName = "instancedb"
		conf.options.DatabaseUser = "instance"
		conf.options.DatabasePassword = "secret"
		conf.options.DatabaseTimeout = 9

		want := "user=instance password=secret dbname=instancedb host=postgres.internal port=5433 connect_timeout=9 sslmode=disable TimeZone=UTC"
		if got := conf.DatabaseDSN(); got != want {
			t.Fatalf("DatabaseDSN() = %q, want %q", got, want)
		}

		assert.Equal(t, "postgres.internal", conf.DatabaseHost())
		assert.Equal(t, 5433, conf.DatabasePort())
	})
}

func TestConfig_DatabaseDSNFlags(t *testing.T) {
	t.Run("NoDatabaseDSN", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		assert.True(t, conf.NoDatabaseDSN())
		assert.False(t, conf.HasDatabaseDSN())
	})
	t.Run("DeprecatedDatabaseDsn", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = dsn.DriverMySQL
		conf.options.Deprecated.DatabaseDsn = "user:pass@tcp(db.internal:3306)/photoprism"

		assert.False(t, conf.NoDatabaseDSN())
		assert.True(t, conf.HasDatabaseDSN())
		assert.Equal(t, "user:pass@tcp(db.internal:3306)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=15s", conf.DatabaseDSN())
		assert.Empty(t, conf.options.Deprecated.DatabaseDsn)
	})
}

func TestConfig_ReportDatabaseDSN(t *testing.T) {
	conf := NewConfig(CliTestContext())
	resetDatabaseOptions(conf)

	assert.Equal(t, dsn.DriverSQLite3, conf.DatabaseDriver())
	assert.True(t, conf.ReportDatabaseDSN())

	conf.options.DatabaseDriver = dsn.DriverMySQL
	conf.options.DatabaseDSN = ""
	assert.False(t, conf.ReportDatabaseDSN())

	conf.options.DatabaseDSN = "user:pass@tcp(db.internal:3306)/photoprism"
	assert.True(t, conf.ReportDatabaseDSN())
}

func TestConfig_DatabaseFile(t *testing.T) {
	c := NewConfig(CliTestContext())
	// Ensure SQLite defaults
	resetDatabaseOptions(c)
	driver := c.DatabaseDriver()
	assert.Equal(t, dsn.DriverSQLite3, driver)
	c.options.DatabaseDSN = ""
	assert.Equal(t, ProjectRoot+"/storage/testdata/index.db", c.DatabaseFile())
	assert.Equal(t, ProjectRoot+"/storage/testdata/index.db?_busy_timeout=5000", c.DatabaseDSN())
}

func TestConfig_DatabaseTimeout(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, 15, c.DatabaseTimeout())
	c.options.DatabaseTimeout = 1
	assert.Equal(t, 1, c.DatabaseTimeout())
	c.options.DatabaseTimeout = -1
	assert.Equal(t, 15, c.DatabaseTimeout())
	c.options.DatabaseTimeout = 120
	assert.Equal(t, 60, c.DatabaseTimeout())
	c.options.DatabaseTimeout = 0
	assert.Equal(t, 15, c.DatabaseTimeout())
	c.options.DatabaseTimeout = 15
	assert.Equal(t, 15, c.DatabaseTimeout())
}

func TestConfig_DatabaseConns(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.DatabaseConns = 28
	assert.Equal(t, 28, c.DatabaseConns())

	c.options.DatabaseConns = 3000
	assert.Equal(t, 1024, c.DatabaseConns())
}

func TestConfig_DatabaseConnsIdle(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.options.DatabaseConnsIdle = 14
	c.options.DatabaseConns = 28
	assert.Equal(t, 14, c.DatabaseConnsIdle())

	c.options.DatabaseConnsIdle = -55
	assert.Greater(t, c.DatabaseConnsIdle(), 8)

	c.options.DatabaseConnsIdle = 35
	assert.Equal(t, 28, c.DatabaseConnsIdle())
}

func TestConfig_checkDb(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Setenv("PHOTOPRISM_DATABASE_SKIP_VERSION_CHECK", "true")
	assert.NoError(t, c.checkDb(nil))
	t.Setenv("PHOTOPRISM_DATABASE_SKIP_VERSION_CHECK", "")
	assert.Error(t, c.checkDb(nil))
}
