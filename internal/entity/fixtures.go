package entity

import (
	"github.com/jinzhu/gorm"
)

// CreateTestFixtures inserts all known entities into the database for testing.
func CreateTestFixtures(db *gorm.DB) {
	CreateLabelFixtures(db)
}
