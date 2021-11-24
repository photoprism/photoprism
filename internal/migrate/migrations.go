package migrate

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Migrations represents a sorted list of migrations.
type Migrations []Migration

// Start runs all migrations that haven't been executed yet.
func (m *Migrations) Start(db *gorm.DB) {
	for _, migration := range *m {
		start := time.Now()

		migration.StartedAt = start.UTC().Round(time.Second)

		// Continue if already executed.
		if err := db.Create(migration).Error; err != nil {
			continue
		}

		if err := migration.Execute(db); err != nil {
			migration.Fail(err, db)
			log.Errorf("migration %s failed: %s [%s]", migration.ID, err, time.Since(start))
		} else if err = migration.Finish(db); err != nil {
			log.Warnf("migration %s failed: %s [%s]", migration.ID, err, time.Since(start))
		} else {
			log.Infof("migration %s successful [%s]", migration.ID, time.Since(start))
		}
	}
}
