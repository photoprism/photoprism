package dsn

import (
	"os"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

// PhotoPrismTestToDriverDSN function to transform PHOTOPRISM_TEST_DSN environment variables to driver and dsn strings
func PhotoPrismTestToDriverDSN() (driver string, dsn string) {
	dsnName := os.Getenv("PHOTOPRISM_TEST_DSN_NAME")
	return PhotoPrismDriverToDriverDSN(dsnName)
}

// PhotoPrismDriverToDriverDSN function to transform named driver to driver and dsn strings
func PhotoPrismDriverToDriverDSN(driverName string) (driver string, dsn string) {
	switch driverName {
	case DriverSQLite3, "sqlitefile":
		driver = DriverSQLite3
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN_SQLITEFILE")
	case DriverMariaDB, DriverMySQL:
		driver = DriverMySQL
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN_MARIADB")
	case "mysql8":
		driver = DriverMySQL
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN_MYSQL8")
	case DriverPostgres, DriverPostgreSQL:
		driver = DriverPostgres
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN_POSTGRES")
	default:
		driver = "sqlite"
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN_SQLITE")
	}

	return driver, dsn
}

// TestDSNFromEnv generates a testing DSN for all supported databases using the environment settings.
// The dbname is used for the database, user and password as per test database standard config.
func TestDSNFromEnv(driver string, dbname string) (dbDSN DSN) {
	_, envDSNStr := PhotoPrismDriverToDriverDSN(driver)
	envDSN := Parse(envDSNStr)
	switch driver {
	case DriverPostgres, DriverPostgreSQL:
		dbDSN = DSN{Driver: DriverPostgreSQL, Name: dbname, Server: envDSN.Server, User: dbname, Password: dbname}
	case DriverSQLite3:
		if !strings.HasSuffix(dbname, ".db") {
			dbname += ".db"
		}
		dbDSN = DSN{Driver: DriverSQLite3, Server: envDSN.Server, Name: clean.TypeLower(dbname)}
	case DriverMySQL:
		fallthrough
	default:
		dbDSN = DSN{Driver: DriverMariaDB, Net: envDSN.Net, Name: dbname, Server: envDSN.Server, User: dbname, Password: dbname}
	}
	return dbDSN
}
