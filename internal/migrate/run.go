package migrate

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// Run automatically migrates the schema of the database passed as argument.
func Run(db *gorm.DB, opt Options) (err error) {
	if db == nil {
		return fmt.Errorf("migrate: no database connection")
	}

	// Get SQL dialect name.
	name := db.Dialect().GetName()

	if name == "" {
		return fmt.Errorf("migrate: failed to determine sql dialect")
	}

	// Make sure a "migrations" table exists.
	once[name].Do(func() {
		err = db.AutoMigrate(&Migration{}).Error
	})

	if err != nil {
		return fmt.Errorf("migrate: %s (create migrations table)", err)
	}

	// Run migrations for dialect.
	if migrations, ok := Dialects[name]; ok && len(migrations) > 0 {
		migrations.Start(db, opt)
		return nil
	} else {
		return fmt.Errorf("migrate: no migrations found for %s", name)
	}
}
