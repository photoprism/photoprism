package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

// Run automatically migrates the schema of the database passed as argument.
func Run(db *gorm.DB, opt Options) (err error) {
	if db == nil || db.Dialector == nil {
		return fmt.Errorf("migrate: no database connection")
	}

	// Get SQL dialect name.
	name := db.Dialector.Name()

	if name == "" {
		return fmt.Errorf("migrate: failed to determine sql dialect")
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
		if len(migrations) > 0 {
			migrations.Start(db, opt)
		}
		return nil
	} else {
		return fmt.Errorf("migrate: no migrations found for %s", name)
	}
}
