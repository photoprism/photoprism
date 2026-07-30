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
	db, err = gorm.Open(drivers[MySQL](dbDSN), gormConfig())
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
		_ = sqlDb.Close()
	}

	return nil
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

	db, err = gorm.Open(drivers[Postgres](dbDSN), gormConfig())
	if err != nil || db == nil {
		log.Fatal(err)
		return err
	}

	if err := db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, dbName)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	// if err := db.Exec(fmt.Sprintf(`CREATE DATABASE "%s" WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.utf8'`, dbName)).Error; err != nil { // ToDo: Why is this using the template etc?
	if err := db.Exec(fmt.Sprintf(`CREATE DATABASE "%s" OWNER %s`, dbName, dbUser)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if sqlDb, err := db.DB(); err != nil {
		log.Fatal(err)
		return err
	} else {
		sqlDb.Close()
	}
	return nil
}
