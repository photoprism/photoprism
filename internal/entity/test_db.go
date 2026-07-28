package entity

import (
	"database/sql"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// testDbNameLen is the maximum length of a MySQL database name.
const testDbNameLen = 64

// testDbNameRegexp matches characters that must not appear in a database name.
var testDbNameRegexp = regexp.MustCompile(`[^a-z0-9_]+`)

// testDbDsn caches the isolated DSN that has been created for a configured DSN.
var testDbDsn = map[string]string{}
var testDbDsnMutex sync.Mutex

// TestDbDSN returns a data source name that gives the package running the tests
// a database of its own, and creates that database if it does not exist yet.
//
// SQLite resolves the DSN to a file in the package directory and is returned
// unchanged. On MySQL and Postgres, all packages would otherwise share a
// single schema and truncate each other's fixtures as soon as they run in parallel.
func TestDbDSN(driver, dbDsn string) string {
	if driver == dsn.DriverSQLite3 || dbDsn == "" {
		return dbDsn
	}

	testDbDsnMutex.Lock()
	defer testDbDsnMutex.Unlock()

	if cached, ok := testDbDsn[dbDsn]; ok {
		return cached
	}

	parsedDSN := dsn.Parse(dbDsn)
	switch parsedDSN.Driver {
	case dsn.DriverMariaDB, dsn.DriverMySQL, dsn.DriverPostgreSQL, dsn.DriverPostgres:
	default:
		// A test that isolates its database with PHOTOPRISM_TEST_DSN alone leaves
		// the driver pointing at MySQL, so its file path lands here and the config
		// error that follows terminates the test binary.
		log.Warnf("testdb: %s is not a postgres or mariadb dsn, set PHOTOPRISM_TEST_DSN_NAME to match it", clean.Log(dbDsn))
		return dbDsn
	}

	name := testDbName(parsedDSN.Name)

	if err := createTestDb(&parsedDSN, name); err != nil {
		log.Warnf("testdb: %s (create test database %s)", err, clean.Log(name))
		return dbDsn
	}

	parsedDSN.Name = name
	testDbDsn[dbDsn] = parsedDSN.ToString()

	return testDbDsn[dbDsn]
}

// testDbName returns a database name for the current working directory, which is
// the source directory of the package whose tests are running.
func testDbName(baseName string) string {
	wd, err := os.Getwd()

	if err != nil {
		log.Warnf("testdb: %s (test database name)", err)
		wd = baseName
	}

	hash := fnv.New32a()
	_, _ = hash.Write([]byte(wd))

	suffix := fmt.Sprintf("_%s_%08x", cleanTestDbName(filepath.Base(wd)), hash.Sum32())
	baseName = cleanTestDbName(baseName)

	if maxLen := testDbNameLen - len(suffix); len(baseName) > maxLen {
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
func createTestDb(conf *dsn.DSN, name string) error {
	if name != cleanTestDbName(name) {
		return fmt.Errorf("invalid database name %s", clean.Log(name))
	}

	var db *sql.DB
	var err error

	switch conf.Driver {
	case dsn.DriverMariaDB, dsn.DriverMySQL:
		//nolint:gosec // G701: the name is checked against cleanTestDbName above.
		db, err = sql.Open(conf.Driver, conf.ToString())

		if err != nil {
			return err
		}

		defer db.Close()
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", name))
	case dsn.DriverPostgreSQL, dsn.DriverPostgres:
		db, err = sql.Open("pgx", conf.ToString())

		if err != nil {
			log.Errorf("pgx open failed with %v", err)
			return err
		}

		defer db.Close()
		q := fmt.Sprintf("SELECT datname FROM pg_database WHERE datname = '%s';", name) //nolint:gosec // G701: the name is checked against cleanTestDbName above.
		if rows, err2 := db.Query(q); err2 == nil {
			defer rows.Close()
			if !rows.Next() {
				//nolint:gosec // G701: the name is checked against cleanTestDbName above.
				_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s OWNER %s", name, conf.User))
			}
		} else {
			return err2
		}
	default:
		err = fmt.Errorf("testdb: unsupported driver %s detected", conf.Driver)
	}

	return err
}
