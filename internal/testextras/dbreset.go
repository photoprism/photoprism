package testextras

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// ResetMariaDB will drop and recreate the named database, assigning permissions to dbUser as needed.
// This should ONLY be used in tests!
func ResetMariaDB(dbName string, dbUser string) error {
	// Prevent this running against the default production database.
	if strings.ToLower(dbName) == "photoprism" {
		return fmt.Errorf("cannot use %s as database name for ResetMariaDB", dbName)
	}

	var db *gorm.DB
	var err error
	d := dsn.TestDSNFromEnv(dsn.DriverMariaDB, dbName)
	d.User = "root"
	d.Password = "photoprism"
	d.Name = "photoprism"
	dbDSN := d.ToString()
	db, err = gorm.Open(dsn.GormDrivers[dsn.DriverMySQL](dbDSN), gormConfig())
	if err != nil || db == nil {
		log.Fatal(err)
		return err
	}

	if err := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if err := db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s.*` TO %s@'%%'", dbName, dbUser)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if err := db.Exec("FLUSH PRIVILEGES").Error; err != nil {
		log.Fatal(err)
		return err
	}

	if sqlDb, err := db.DB(); err != nil {
		log.Fatal(err)
		return err
	} else {
		return sqlDb.Close()
	}
}

// ResetPostgresDB will drop and recreate the named database, assigning permissions to dbUser as needed.
// This should ONLY be used in tests!
func ResetPostgresDB(dbName string, dbUser string) error {
	// Prevent this running against the default production database.
	if strings.ToLower(dbName) == "photoprism" {
		return fmt.Errorf("cannot use %s as database name for ResetPostgresDB", dbName)
	}

	var db *gorm.DB
	var err error
	d := dsn.TestDSNFromEnv(dsn.DriverPostgreSQL, dbName)
	d.User = "photoprism"
	d.Password = "photoprism"
	d.Name = "postgres"
	dbDSN := d.ToString()

	db, err = gorm.Open(dsn.GormDrivers[dsn.DriverPostgres](dbDSN), gormConfig())
	if err != nil || db == nil {
		log.Fatal(err)
		return err
	}

	if err := db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, dbName)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if err := db.Exec(fmt.Sprintf(`CREATE DATABASE "%s" OWNER %s`, dbName, dbUser)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if sqlDb, err := db.DB(); err != nil {
		log.Fatal(err)
		return err
	} else {
		return sqlDb.Close()
	}
}
