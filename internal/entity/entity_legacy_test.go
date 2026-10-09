package entity

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/entity/sqlcount"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

// legacyDDL matches statements that change the schema, including after leading comments. MySQL runs
// the content of /*! ... */ comments, so a statement in one is matched.
var legacyDDL = regexp.MustCompile(`(?is)^\s*(?:(?:(?:--|#)[^\n]*(?:\n|$)|/\*(?:[^!].*?)?\*/)\s*)*(?:/\*!\d*\s*)?(CREATE|ALTER|DROP|RENAME|TRUNCATE)\b`)

// migrateLegacySchema migrates a database created from a legacy schema twice through a counted connection,
// and returns the statements of the second run, which must not change the schema again.
func migrateLegacySchema(t *testing.T, driver, dataSourceName string) []string {
	t.Helper()

	p, err := sqlcount.OpenGorm(driver, dataSourceName)
	require.NoError(t, err)
	t.Cleanup(func() { _ = p.Close() })

	db := p.Db()

	// The database holds the legacy schema, which lacks columns added since.
	found, err := DbHasTable(db, "photos")
	require.NoError(t, err)
	require.True(t, found, "legacy photos table")
	found, err = DbHasColumn(db, "photos", "photo_caption")
	require.NoError(t, err)
	require.False(t, found, "legacy photos table without photo_caption")

	// start migrates the schema the way InitDb does on the first start of a new release.
	start := func() {
		opt := migrate.Opt(true, false, nil)
		require.NoError(t, db.AutoMigrate(&migrate.Migration{}, &migrate.Version{}).Error)
		DeprecatedTables.Drop(db)
		require.NoError(t, Entities.Migrate(db, opt))
	}

	// The first run upgrades the legacy schema.
	start()

	var failed []string
	require.NoError(t, db.Model(&migrate.Migration{}).Where("error <> '' OR finished_at IS NULL").Pluck("id", &failed).Error)
	assert.Empty(t, failed, "failed migrations")

	for name := range Entities {
		found, hasErr := DbHasTable(db, name)
		require.NoError(t, hasErr)
		assert.True(t, found, "table %s exists after migrating", name)
	}

	found, err = DbHasColumn(db, "photos", "photo_caption")
	require.NoError(t, err)
	assert.True(t, found, "photo_caption exists after migrating")

	// A second run, as on the next start of a new release, finds nothing to change.
	p.Counter.Start()
	start()

	return p.Counter.Stop()
}

// assertNoDDL checks that none of the statements changes the schema.
func assertNoDDL(t *testing.T, statements []string) {
	t.Helper()

	require.NotEmpty(t, statements)

	for _, s := range statements {
		assert.False(t, legacyDDL.MatchString(s), "second migration changes the schema: %s", s)
	}
}

// TestLegacySchema_SQLite migrates the SQLite schema of an early release to the current one.
func TestLegacySchema_SQLite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	fileName := filepath.Join(t.TempDir(), "legacy.db")
	require.NoError(t, fs.Copy("migrate/testdata/migrate_sqlite3", fileName, true))

	assertNoDDL(t, migrateLegacySchema(t, dsn.DriverSQLite3, fileName))
}

// TestLegacySchema_MariaDB migrates the MariaDB schema of an early release to the current one.
func TestLegacySchema_MariaDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conn, ok := dbConn.(*DbConn)
	require.True(t, ok, "test database provider")

	if conn.Driver != dsn.DriverMySQL {
		t.Skip("requires the MariaDB test database")
	}

	if _, err := exec.LookPath("mariadb"); err != nil {
		t.Skip("requires the mariadb client")
	}

	conf, err := mysql.ParseDSN(conn.Dsn)
	require.NoError(t, err)

	// Use a database of its own, so the legacy schema does not replace the fixtures; its name starts
	// with "acceptance_", so resetting the test databases also drops it.
	name := testDbName("acceptance_legacy")
	require.NotEqual(t, conf.DBName, name)

	serverConf := conf.Clone()
	serverConf.DBName = ""
	server, err := sql.Open(dsn.DriverMySQL, serverConf.FormatDSN())
	require.NoError(t, err)
	t.Cleanup(func() { _ = server.Close() })

	_, err = server.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", name))
	require.NoError(t, err)
	_, err = server.Exec(fmt.Sprintf("CREATE DATABASE `%s`", name))
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = server.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", name)) })

	// The dump creates and selects a database of its own, so load only what follows its USE statement.
	b, err := os.ReadFile("migrate/testdata/migrate_mysql.sql")
	require.NoError(t, err)
	_, schema, found := strings.Cut(string(b), "\nUSE migrate;\n")
	require.True(t, found, "USE statement in dump")
	dump := filepath.Join(t.TempDir(), "legacy.sql")
	require.NoError(t, os.WriteFile(dump, []byte(schema), fs.ModeFile)) //nolint:gosec // G703: test-owned temporary file.

	host, port := conf.Addr, "3306"

	if i := strings.LastIndex(conf.Addr, ":"); i > 0 {
		host, port = conf.Addr[:i], conf.Addr[i+1:]
	}

	//nolint:gosec // G204: the arguments come from the test configuration and a temporary file.
	cmd := exec.Command("mariadb", "--no-defaults", "--protocol", "tcp", "-h", host, "-P", port, "-u", conf.User, name, "-e", "source "+dump)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+conf.Passwd)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	conf.DBName = name
	assertNoDDL(t, migrateLegacySchema(t, dsn.DriverMySQL, conf.FormatDSN()))
}
