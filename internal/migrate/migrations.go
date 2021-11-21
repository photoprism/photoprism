package migrate

import "github.com/jinzhu/gorm"

// Migrations represents a sorted list of migrations.
type Migrations []Migration

// Start runs all migrations that haven't been executed yet.
func (m *Migrations) Start(db *gorm.DB) {
	for _, migration := range *m {
		migration.Execute(db)
	}
}
