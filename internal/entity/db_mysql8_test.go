package entity

import (
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

func TestMySQL8(t *testing.T) {
	dbDriver := MySQL
	dbDsn := os.Getenv("PHOTOPRISM_TEST_DSN_MYSQL8")

	db, err := gorm.Open(dbDriver, dbDsn)

	if err != nil || db == nil {
		for i := 1; i <= 5; i++ {
			db, err = gorm.Open(dbDriver, dbDsn)

			if db != nil && err == nil {
				break
			}

			time.Sleep(5 * time.Second)
		}

		if err != nil || db == nil {
			t.Fatal(err)
		}
	}

	defer db.Close()

	db.LogMode(false)

	DeprecatedTables.Drop(db)
	Entities.Drop(db)

	// First migration.
	Entities.Migrate(db, false)
	Entities.WaitForMigration(db)

	// Second migration.
	Entities.Migrate(db, false)
	Entities.WaitForMigration(db)

	// Third migration with force flag.
	Entities.Migrate(db, true)
	Entities.WaitForMigration(db)
}
