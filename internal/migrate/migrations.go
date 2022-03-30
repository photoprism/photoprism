package migrate

import (
	"database/sql"
	"time"

	"github.com/photoprism/photoprism/pkg/list"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"
)

// Migrations represents a sorted list of migrations.
type Migrations []Migration

// MigrationMap represents a map of migrations.
type MigrationMap map[string]Migration

// Existing finds and returns previously executed database schema migrations.
func Existing(db *gorm.DB) MigrationMap {
	result := make(MigrationMap)

	stmt := db.Model(Migration{})
	stmt = stmt.Select("id, dialect, error, source, started_at, finished_at")

	rows, err := stmt.Rows()

	if err != nil {
		log.Warnf("migrate: %s (find existing)", err)
		return result
	}

	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	for rows.Next() {
		m := Migration{}

		if err = rows.Scan(&m.ID, &m.Dialect, &m.Error, &m.Source, &m.StartedAt, &m.FinishedAt); err != nil {
			log.Warnf("migrate: %s (scan existing)", err)
			return result
		}

		result[m.ID] = m
	}

	return result
}

// Start runs all migrations that haven't been executed yet.
func (m *Migrations) Start(db *gorm.DB, runFailed bool, ids []string) {
	// Find previously executed migrations.
	executed := Existing(db)

	if prev := len(executed); prev == 0 {
		log.Infof("migrate: no previously executed migrations")
	} else {
		log.Debugf("migrate: found %s", english.Plural(len(executed), "previous migration", "previous migrations"))
	}

	for _, migration := range *m {
		start := time.Now()
		migration.StartedAt = start.UTC().Round(time.Second)

		// Excluded?
		if list.Excludes(ids, migration.ID) {
			log.Debugf("migrate: %s skipped", migration.ID)
			continue
		}

		// Already executed?
		if done, ok := executed[migration.ID]; ok {
			// Try to run failed migrations again?
			if (!runFailed || done.Error == "") && !list.Contains(ids, migration.ID) {
				log.Debugf("migrate: %s skipped", migration.ID)
				continue
			}
		} else if err := db.Create(migration).Error; err != nil {
			// Should not happen.
			log.Warnf("migrate: creating %s failed with %s [%s]", migration.ID, err, time.Since(start))
			continue
		}

		// Run migration.
		if err := migration.Execute(db); err != nil {
			migration.Fail(err, db)
			log.Errorf("migrate: executing %s failed with %s [%s]", migration.ID, err, time.Since(start))
		} else if err = migration.Finish(db); err != nil {
			log.Warnf("migrate: updating %s failed with %s [%s]", migration.ID, err, time.Since(start))
		} else {
			log.Infof("migrate: %s successful [%s]", migration.ID, time.Since(start))
		}
	}
}
