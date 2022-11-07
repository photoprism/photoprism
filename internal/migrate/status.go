package migrate

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/list"
)

// Status returns the current status of schema migrations.
func Status(db *gorm.DB, ids []string) (status Migrations, err error) {
	status = Migrations{}

	if db == nil {
		return status, fmt.Errorf("migrate: no database connection")
	}

	// Get SQL dialect name.
	name := db.Dialect().GetName()

	if name == "" {
		return status, fmt.Errorf("migrate: failed to determine sql dialect")
	}

	// Make sure a "migrations" table exists.
	once[name].Do(func() {
		err = db.AutoMigrate(&Migration{}).Error
	})

	if err != nil {
		return status, fmt.Errorf("migrate: %s (create migrations table)", err)
	}

	migrations, ok := Dialects[name]

	if !ok && len(migrations) == 0 {
		return status, fmt.Errorf("migrate: no migrations found for %s", name)
	}

	// Find previously executed migrations.
	executed := Existing(db, "")

	if prev := len(executed); prev == 0 {
		log.Infof("migrate: no previously executed migrations")
	} else {
		log.Debugf("migrate: found %s", english.Plural(len(executed), "previous migration", "previous migrations"))
	}

	for _, migration := range migrations {
		// Excluded?
		if list.Excludes(ids, migration.ID) {
			continue
		}

		// Already executed?
		if done, known := executed[migration.ID]; known {
			migration.Dialect = done.Dialect
			migration.Stage = done.Stage
			migration.Error = done.Error
			migration.Source = done.Source
			migration.StartedAt = done.StartedAt
			migration.FinishedAt = done.FinishedAt
			status = append(status, migration)
		} else {
			// Should not happen.
			status = append(status, migration)
		}
	}

	return status, nil
}
