package dsn

import (
	"fmt"
	"os"
)

// PhotoPrismTestToDriverDSN function to transform PHOTOPRISM_TEST_DSN environment variables to driver and dsn strings
func PhotoPrismTestToDriverDSN(dbn int) (driver string, dsn string) {
	dsnName := os.Getenv("PHOTOPRISM_TEST_DSN_NAME")
	switch dsnName {
	case "sqlite":
		driver = "sqlite"
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN_SQLITE")
	case "sqlitefile":
		driver = "sqlite"
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN_SQLITEFILE")
		if dbn > 0 {
			d := Parse(dsn)
			d.Driver = driver
			d.Name = fmt.Sprintf("%s_%02d.db", d.Name, dbn)
			dsn = d.ToString()
		}
	case "mariadb":
		driver = "mysql"
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN_MARIADB")
		if dbn > 0 {
			d := Parse(dsn)
			d.Driver = driver
			d.Name = fmt.Sprintf("%s_%02d", d.Name, dbn)
			dsn = d.ToString()
		}
	case "mysql8":
		driver = "mysql"
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN_MYSQL8")
		if dbn > 0 {
			d := Parse(dsn)
			d.Driver = driver
			d.Name = fmt.Sprintf("%s_%02d", d.Name, dbn)
			dsn = d.ToString()
		}
	case "postgres":
		driver = "postgres"
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN_POSTGRES")
		if dbn > 0 {
			d := Parse(dsn)
			d.Driver = driver
			d.Name = fmt.Sprintf("%s_%02d", d.Name, dbn)
			dsn = d.ToString()
		}
	default:
		driver = "sqlite"
		dsn = ""
	}

	return driver, dsn
}

// PhotoPrismTestToFolderName gets the folder name to use to enforce folder separation for DBMS tests
func PhotoPrismTestToFolderName() (folderName string) {
	folderName = os.Getenv("PHOTOPRISM_TEST_DSN_NAME")
	if folderName == "" {
		folderName, _ = PhotoPrismTestToDriverDSN(0)
	}
	return folderName
}

// SetDSNToEnv pushes the required parameters into the os environment so tests run correctly.
func SetDSNToEnv(dsn string) {
	dsnName := os.Getenv("PHOTOPRISM_TEST_DSN_NAME")
	switch dsnName {
	case "sqlite":
		_ = os.Setenv("PHOTOPRISM_TEST_DSN_SQLITE", dsn)
	case "sqlitefile":
		_ = os.Setenv("PHOTOPRISM_TEST_DSN_SQLITEFILE", dsn)
	case "mariadb":
		_ = os.Setenv("PHOTOPRISM_TEST_DSN_MARIADB", dsn)
	case "mysql8":
		_ = os.Setenv("PHOTOPRISM_TEST_DSN_MYSQL8", dsn)
	case "postgres":
		_ = os.Setenv("PHOTOPRISM_TEST_DSN_POSTGRES", dsn)
	}
}

// TestDSNPortFromEnv generates a testing DSN for MariaDB or Postgres using the appropriate environment port.
// the dbname is used for the database, user and password as per test database standard config.
func TestDSNPortFromEnv(driver string, dbname string, dbn int) (dbDSN DSN) {
	switch driver {
	case DriverMySQL:
		port := os.Getenv("MARIADB_PORT")
		if port == "" {
			port = "4001"
		}
		dbDSN = DSN{Driver: DriverMariaDB, Net: "tcp", Name: fmt.Sprintf("%s_%02d", dbname, dbn), Server: fmt.Sprintf("mysql:%s", port), User: dbname, Password: dbname}
	case DriverPostgres, DriverPostgreSQL:
		port := os.Getenv("POSTGRES_PORT")
		if port == "" {
			port = "4002"
		}
		dbDSN = DSN{Driver: DriverPostgreSQL, Name: fmt.Sprintf("%s_%02d", dbname, dbn), Server: fmt.Sprintf("postgres:%s", port), User: dbname, Password: dbname}
	default:
		port := os.Getenv("MARIADB_PORT")
		if port == "" {
			port = "4001"
		}
		dbDSN = DSN{Driver: DriverMariaDB, Net: "tcp", Name: fmt.Sprintf("%s_%02d", dbname, dbn), Server: fmt.Sprintf("mariadb:%s", port), User: dbname, Password: dbname}
	}
	return dbDSN
}
