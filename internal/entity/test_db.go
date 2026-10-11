package entity

import (
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// Maximum database name lengths of MySQL and PostgreSQL.
const (
	testDbNameLen         = 64
	testPostgresDbNameLen = 63
)

// testDbNameRegexp matches characters that must not appear in a database name.
var testDbNameRegexp = regexp.MustCompile(`[^a-z0-9_]+`)

// testDbDsn caches the isolated DSN that has been created for a configured DSN.
var testDbDsn = map[string]string{}
var testDbDsnMutex sync.Mutex

// TestDbDSN returns a data source name that gives the package running the tests a database of its own,
// and creates that database if it does not exist yet. SQLite DSNs resolve to a file in the package
// directory and are returned unchanged; on MySQL and PostgreSQL, packages running in parallel would
// otherwise truncate each other's fixtures.
func TestDbDSN(driver, dbDsn string) string {
	if dbDsn == "" {
		return dbDsn
	}

	switch dsn.ParseDriver(driver) {
	case dsn.DriverMySQL, dsn.DriverPostgres:
	default:
		return dbDsn
	}

	testDbDsnMutex.Lock()
	defer testDbDsnMutex.Unlock()

	if cached, ok := testDbDsn[dbDsn]; ok {
		return cached
	}

	var isolated string
	var err error

	if dsn.ParseDriver(driver) == dsn.DriverPostgres {
		isolated, err = testPostgresDbDSN(dbDsn)
	} else {
		isolated, err = testMySQLDbDSN(dbDsn)
	}

	if err != nil {
		log.Warnf("%s: %s (create test database)", dsn.ParseDriver(driver), err)
		return dbDsn
	}

	testDbDsn[dbDsn] = isolated

	return isolated
}

// testMySQLDbDSN creates the MySQL test database of the package and returns its DSN.
func testMySQLDbDSN(dbDsn string) (string, error) {
	conf, err := mysql.ParseDSN(dbDsn)

	if err != nil {
		// A test that isolates its database with PHOTOPRISM_TEST_DSN alone leaves
		// the driver pointing at MySQL, so its file path lands here and the config
		// error that follows terminates the test binary.
		return "", fmt.Errorf("%s is not a mysql dsn, set PHOTOPRISM_TEST_DRIVER to match it (%s)", clean.Log(dsn.Mask(dbDsn)), err)
	}

	name := testDbName(conf.DBName, testDbNameLen)

	if err = createTestDb(conf, name); err != nil {
		return "", fmt.Errorf("%s (%s)", err, clean.Log(name))
	}

	conf.DBName = name

	return conf.FormatDSN(), nil
}

// testPostgresDbDSN creates the PostgreSQL test database of the package and returns its DSN, which
// must be a connection URI whose user may create databases.
func testPostgresDbDSN(dbDsn string) (string, error) {
	u, err := url.Parse(dbDsn)

	if err != nil || u.Scheme == "" {
		return "", fmt.Errorf("%s is not a postgres connection uri", clean.Log(dsn.Mask(dbDsn)))
	}

	name := testDbName(strings.TrimPrefix(u.Path, "/"), testPostgresDbNameLen)

	if err = createPostgresTestDb(dbDsn, name); err != nil {
		return "", fmt.Errorf("%s (%s)", err, clean.Log(name))
	}

	u.Path = "/" + name
	u.RawPath = ""

	return u.String(), nil
}

// testDbName returns a database name for the current working directory, which is
// the source directory of the package whose tests are running.
func testDbName(baseName string, maxNameLen int) string {
	wd, err := os.Getwd()

	if err != nil {
		log.Warnf("mysql: %s (test database name)", err)
		wd = baseName
	}

	hash := fnv.New32a()
	_, _ = hash.Write([]byte(wd))

	suffix := fmt.Sprintf("_%s_%08x", cleanTestDbName(filepath.Base(wd)), hash.Sum32())
	baseName = cleanTestDbName(baseName)

	if maxLen := maxNameLen - len(suffix); len(baseName) > maxLen {
		baseName = baseName[:maxLen]
	}

	return baseName + suffix
}

// cleanTestDbName removes the characters that a database name must not contain.
func cleanTestDbName(s string) string {
	return testDbNameRegexp.ReplaceAllString(strings.ToLower(s), "")
}

// createTestDb creates the database with the specified name if it does not exist.
//
// The name must have been built by testDbName, which restricts it to the
// characters a database identifier may contain, as it cannot be parameterized.
func createTestDb(conf *mysql.Config, name string) error {
	if name != cleanTestDbName(name) {
		return fmt.Errorf("invalid database name %s", clean.Log(name))
	}

	serverConf := conf.Clone()
	serverConf.DBName = ""

	db, err := sql.Open(dsn.DriverMySQL, serverConf.FormatDSN())

	if err != nil {
		return err
	}

	defer db.Close()

	//nolint:gosec // G701: the name is checked against cleanTestDbName above.
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", name))

	return err
}

// createPostgresTestDb creates the PostgreSQL database with the specified name if it does not exist,
// connecting to the database of the DSN, or to the "postgres" database if that does not exist.
// The name must come from testDbName, which restricts it to identifier characters, as it cannot be
// parameterized.
func createPostgresTestDb(serverDsn, name string) error {
	if name != cleanTestDbName(name) {
		return fmt.Errorf("invalid database name %s", clean.Log(name))
	}

	config, err := pgx.ParseConfig(serverDsn)

	// The parse error is not returned, as it may quote parts of the password.
	if err != nil {
		return errors.New("invalid postgres connection string")
	}

	db := stdlib.OpenDB(*config)
	defer db.Close()

	var exists bool

	if err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil && config.Database != "postgres" {
		config.Database = "postgres"
		db = stdlib.OpenDB(*config)
		defer db.Close()
		err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists)
	}

	if err != nil || exists {
		return err
	}

	// Text sorts by Unicode rules, case-insensitively first, as on MariaDB, if the server supports ICU.
	//nolint:gosec // G701: the name is checked against cleanTestDbName above.
	if _, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s ENCODING 'UTF8' LOCALE_PROVIDER icu ICU_LOCALE 'und' TEMPLATE template0", name)); err == nil {
		return nil
	}

	//nolint:gosec // G701: the name is checked against cleanTestDbName above.
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s ENCODING 'UTF8'", name))

	return err
}
