package testextras

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// ResetMariaDB will drop and recreate the named database, using the dbID as part of the name as fmt.Sprintf("%s_%02d", dbName, dbID)
// unless dbID = 0, in which case it's just the dbName.
// It assumes that the grant will be done to the user as dbName, and that the user already exists.
// This should ONLY be used in tests!
func ResetMariaDB(dbName string, dbID int) error {
	// Prevent this running against the default production database.
	if strings.ToLower(dbName) == "photoprism" {
		return fmt.Errorf("cannot use %s as database name for ResetMariaDB", dbName)
	}

	var dbn string
	if dbID > 0 {
		dbn = fmt.Sprintf("%s_%02d", dbName, dbID)
	} else {
		dbn = dbName
	}

	var db *gorm.DB
	var err error
	d := dsn.TestDSNPortFromEnv(dsn.DriverMariaDB, "migrate", 1)
	d.User = "root"
	d.Password = "photoprism"
	d.Name = "photoprism"
	dbDSN := d.ToString()
	db, err = gorm.Open(drivers[MySQL](dbDSN), gormConfig())
	if err != nil || db == nil {
		log.Fatal(err)
		return err
	}

	if err := db.Debug().Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbn)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if err := db.Debug().Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbn)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if err := db.Debug().Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s.*` TO %s@'%%'", dbn, dbName)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if err := db.Debug().Exec("FLUSH PRIVILEGES").Error; err != nil {
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

// ResetPostgresDB will drop and recreate the named database, using the dbID as part of the name as fmt.Sprintf("%s_%02d", dbName, dbID)
// unless dbID = 0, in which case it's just the dbName.
// It assumes that the grant will be done to the user as dbName, and that the user already exists.
// This should ONLY be used in tests!
func ResetPostgresDB(dbName string, dbID int) error {
	// Prevent this running against the default production database.
	if strings.ToLower(dbName) == "photoprism" {
		return fmt.Errorf("cannot use %s as database name for ResetPostgresDB", dbName)
	}

	var dbn string
	if dbID > 0 {
		dbn = fmt.Sprintf("%s_%02d", dbName, dbID)
	} else {
		dbn = dbName
	}

	var db *gorm.DB
	var err error
	d := dsn.TestDSNPortFromEnv(dsn.DriverPostgreSQL, "migrate", 1)
	d.User = "photoprism"
	d.Password = "photoprism"
	d.Name = "postgres"
	dbDSN := d.ToString()

	db, err = gorm.Open(drivers[Postgres](dbDSN), gormConfig())
	if err != nil || db == nil {
		log.Fatal(err)
		return err
	}

	if err := db.Debug().Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, dbn)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if err := db.Debug().Exec(fmt.Sprintf(`CREATE DATABASE "%s" WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.utf8'`, dbn)).Error; err != nil {
		log.Fatal(err)
		return err
	}

	if err := db.Debug().Exec(fmt.Sprintf(`ALTER DATABASE "%s" OWNER TO "%s"`, dbn, dbName)).Error; err != nil {
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
