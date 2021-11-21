package migrate

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// Auto automatically migrates the database provided.
func Auto(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("migrate: database connection required")
	}

	name := db.Dialect().GetName()

	if name == "" {
		return fmt.Errorf("migrate: database has no dialect name")
	}

	if err := db.AutoMigrate(&Migration{}).Error; err != nil {
		return fmt.Errorf("migrate: %s (create migrations table)", err)
	}

	if migrations, ok := Dialects[name]; ok && len(migrations) > 0 {
		migrations.Start(db)
		return nil
	} else {
		return fmt.Errorf("migrate: no migrations found for %s", name)
	}
}
