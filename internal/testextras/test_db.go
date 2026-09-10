package testextras

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	// Required drivers for database/sql
	// ToDo: replace with gorm calls
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// testDbNameLen is the maximum length of a MySQL database name.
const testDbNameLen = 64

// testDbNameRegexp matches characters that must not appear in a database name.
var testDbNameRegexp = regexp.MustCompile(`[^a-z0-9_]+`)

// testDbDsn caches the isolated DSN that has been created for a configured DSN.
var testDbDsn = map[string]string{}
var testDbDsnMutex sync.Mutex

// TestDbFromDSN returns a data source name that gives the package running the tests
// a database of its own, and creates that database if it does not exist yet.
// Calls TestDbDSN to do the actual work
func TestDbFromDSN(dbDSN dsn.DSN) string {
	return TestDbDSN(dbDSN.Driver, dbDSN.Name)
}

// TestDbDSN returns a data source name that gives the package running the tests
// a database of its own, and creates that database if it does not exist yet.
//
// The drivers admin DSN is retrieved from the environment and used to create the
// database. All packages would otherwise share a single schema and truncate each
// other's fixtures as soon as they run in parallel.
func TestDbDSN(driver, dbName string) string {
	if dbName == "" {
		log.Warn("testdb: no dbname was provided in call")
		return dbName
	}

	parsedDSN := dsn.TestDSNFromEnv(driver, dbName)
	stringDSN := parsedDSN.ToString()
	testDbDsnMutex.Lock()
	defer testDbDsnMutex.Unlock()

	if cached, ok := testDbDsn[stringDSN]; ok {
		return cached
	}

	switch driver {
	case dsn.DriverMariaDB, dsn.DriverMySQL, dsn.DriverPostgreSQL, dsn.DriverPostgres, dsn.DriverSQLite3:
	default:
		// A test that isolates its database with PHOTOPRISM_TEST_DSN alone leaves
		// the driver pointing at MySQL, so its file path lands here and the config
		// error that follows terminates the test binary.
		log.Warnf("testdb: %s is not a supported driver, set PHOTOPRISM_TEST_DSN_NAME correctly", clean.Log(driver))
		return ""
	}

	name := testDbName(parsedDSN.Name)

	if err := createTestDb(&parsedDSN, name); err != nil {
		log.Warnf("testdb: %s (create test database %s)", err, clean.Log(name))
		return parsedDSN.ToString()
	}

	if driver == dsn.DriverSQLite3 {
		name += ".db"
	}

	parsedDSN.Name = name
	testDbDsn[stringDSN] = parsedDSN.ToString()

	return testDbDsn[stringDSN]
}

// testDbName returns a database name for the current working directory, which is
// the source directory of the package whose tests are running.
// And adds a unique identifier after the name to guarantee no database name clashes
func testDbName(baseName string) string {
	wd, err := os.Getwd()

	if err != nil {
		log.Warnf("testdb: %s (test database name)", err)
		wd = baseName
	}

	breaker := rnd.GenerateUID('d')

	suffix := fmt.Sprintf("_%s_%s", cleanTestDbName(filepath.Base(wd)), breaker)
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

	driver, dsname := dsn.PhotoPrismDriverToDriverDSN(conf.Driver)

	switch conf.Driver {
	case dsn.DriverMariaDB, dsn.DriverMySQL:
		//nolint:gosec // G701: the name is checked against cleanTestDbName above.
		db, err = sql.Open(driver, dsname)

		if err != nil {
			return err
		}

		defer db.Close()
		str := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", name)
		_, err = db.Exec(str)
		if err == nil {
			str := fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%';", name, conf.User)
			_, err = db.Exec(str)
		}
	case dsn.DriverPostgreSQL, dsn.DriverPostgres:
		db, err = sql.Open("pgx", dsname)

		if err != nil {
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
	case dsn.DriverSQLite3:
		// SQLite will create the database file on open, so nothing to do here?
		newDSN := dsn.TestDSNFromEnv(driver, name)
		db, err = sql.Open("sqlite3", newDSN.ToString())

		if err != nil {
			return err
		}
		_, _ = db.Query("select 1") // Force the database to be created

		defer db.Close()

	default:
		err = fmt.Errorf("testdb: unsupported driver %s detected", conf.Driver)
	}

	return err
}

// TestDbRemoveByName will remove the database that has been created by TestDbDSN
// But it finds it by the base name and the cached value from that.
// Use this where you have generated separate databases directly.
func TestDbRemoveByName(driver, name string) error {
	eDSN := dsn.TestDSNFromEnv(driver, name)
	return TestDbRemoveByDSN(driver, eDSN.ToString())
}

// TestDbRemoveByDSN will remove the database that has been created by TestDbDSN
// It will only remove a database whose name has been cached.
func TestDbRemoveByDSN(driver, dbDsn string) error {
	if dbDsn == "" {
		return nil
	}

	testDbDsnMutex.Lock()
	defer testDbDsnMutex.Unlock()

	var cached string
	var ok bool
	if cached, ok = testDbDsn[dbDsn]; !ok {
		log.Warnf("testdb: database dsn %s not found in cache", dbDsn)
		return nil
	}

	name := dsn.Parse(cached).Name

	driver, dsname := dsn.PhotoPrismDriverToDriverDSN(driver)
	var db *sql.DB
	var err error

	switch driver {
	case dsn.DriverMariaDB, dsn.DriverMySQL:
		//nolint:gosec // G701: the name is checked against cleanTestDbName above.
		db, err = sql.Open(driver, dsname)

		if err != nil {
			return err
		}

		defer db.Close()
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", name))
	case dsn.DriverPostgreSQL, dsn.DriverPostgres:
		db, err = sql.Open("pgx", dsname)

		if err != nil {
			return err
		}

		defer db.Close()
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", name))
	case dsn.DriverSQLite3:
		cachedDSN := dsn.Parse(cached)
		if cachedDSN.Server == "" {
			if _, err = os.Stat(cachedDSN.Name); err == nil {
				err = os.Remove(cachedDSN.Name)
			} else {
				err = nil
			}
		} else {
			filePath := strings.Replace(cachedDSN.Server, "file:", "", 1)
			fileName := filepath.Join(filePath, cachedDSN.Name)
			if _, err = os.Stat(fileName); err == nil {
				err = os.Remove(fileName)
			} else {
				err = nil
			}
		}
	default:
		err = fmt.Errorf("testdb: unsupported driver %s detected", driver)
	}

	if err == nil {
		delete(testDbDsn, dbDsn)
	}
	return err
}

// TestDbCleanup cleans up the created test database
func TestDbCleanup(code int) int {
	if code == 0 {
		driver, _ := dsn.PhotoPrismTestToDriverDSN()
		if err := TestDbRemoveByName(driver, "testdb"); err != nil {
			log.Errorf("remove database: %v", err)
			return 1
		}
	}
	return code
}
