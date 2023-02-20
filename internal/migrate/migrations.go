package migrate

import (
	"fmt"
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
func Existing(db *gorm.DB, stage string) MigrationMap {
	var err error

	if db == nil {
		return make(MigrationMap)
	}

	// Get SQL dialect name.
	name := db.Dialect().GetName()

	if name == "" {
		return make(MigrationMap)
	}

	// Make sure a "migrations" table exists.
	once[name].Do(func() {
		err = db.AutoMigrate(&Migration{}).Error
	})

	if err != nil {
		return make(MigrationMap)
	}

	found := Migrations{}

	stmt := db

	if stage == StageMain {
		stmt = stmt.Where("stage = ? OR stage = '' OR stage IS NULL", stage)
	} else if stage != "" {
		stmt = stmt.Where("stage = ?", stage)
	}

	if err = stmt.Find(&found).Error; err != nil {
		log.Warnf("migrate: %s (find existing)", err)
		return make(MigrationMap)
	}

	result := make(MigrationMap, len(found))

	for _, m := range found {
		result[m.ID] = m
	}

	return result
}

// Start runs all migrations that haven't been executed yet.
func (m *Migrations) Start(db *gorm.DB, opt Options) {
	if db == nil {
		return
	}

	// Find previously executed migrations.
	executed := Existing(db, opt.StageName())

	// Log information about existing migrations.
	if prev := len(executed); prev > 0 {
		stage := fmt.Sprintf("previously executed %s stage", opt.StageName())
		log.Tracef("migrate: found %s", english.Plural(len(executed), stage+" migration", stage+" migrations"))
	}

	// Run migrations.
	for _, migration := range *m {
		if migration.Skip(opt) {
			continue
		}

		start := time.Now()
		migration.StartedAt = start.UTC().Truncate(time.Second)

		// Excluded?
		if list.Excludes(opt.Migrations, migration.ID) {
			log.Tracef("migrate: %s skipped", migration.ID)
			continue
		}

		// Already executed?
		if done, ok := executed[migration.ID]; ok {
			// Repeat?
			if !done.Repeat(opt.RunFailed) && !list.Contains(opt.Migrations, migration.ID) {
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
