package migrate

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/dsn"
)

// Run automatically migrates the schema of the database passed as argument.
func Run(db *gorm.DB, opt Options) (err error) {
	if db == nil {
		return fmt.Errorf("migrate: no database connection")
	}

	// Get SQL dialect name.
	name := dsn.ParseDriver(db.Name())

	if name == dsn.DriverNone {
		return fmt.Errorf("migrate: failed to determine sql dialect")
	}

	if once[name] == nil {
		return fmt.Errorf("migrate: unsupported sql dialect %s", name)
	}

	// Make sure a "migrations" table exists.
	once[name].Do(func() {
		err = db.AutoMigrate(&Migration{})
	})

	if err != nil {
		return fmt.Errorf("migrate: %s (create migrations table)", err)
	}

	// Run migrations for dialect.
	if migrations, ok := Dialects[name]; ok {
		migrations.Start(db, opt)
		return nil
	} else {
		return fmt.Errorf("migrate: no migrations found for %s", name)
	}
}
