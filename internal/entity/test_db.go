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

	"github.com/go-sql-driver/mysql"

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
// unchanged. On MySQL, all packages would otherwise share a single schema and
// truncate each other's fixtures as soon as they run in parallel.
func TestDbDSN(driver, dbDsn string) string {
	if driver != dsn.DriverMySQL || dbDsn == "" {
		return dbDsn
	}

	testDbDsnMutex.Lock()
	defer testDbDsnMutex.Unlock()

	if cached, ok := testDbDsn[dbDsn]; ok {
		return cached
	}

	conf, err := mysql.ParseDSN(dbDsn)

	if err != nil {
		log.Warnf("mysql: %s (parse test database dsn)", err)
		return dbDsn
	}

	name := testDbName(conf.DBName)

	if err = createTestDb(conf, name); err != nil {
		log.Warnf("mysql: %s (create test database %s)", err, clean.Log(name))
		return dbDsn
	}

	conf.DBName = name
	testDbDsn[dbDsn] = conf.FormatDSN()

	return testDbDsn[dbDsn]
}

// testDbName returns a database name for the current working directory, which is
// the source directory of the package whose tests are running.
func testDbName(baseName string) string {
	wd, err := os.Getwd()

	if err != nil {
		log.Warnf("mysql: %s (test database name)", err)
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
