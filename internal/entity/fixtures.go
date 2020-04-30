package entity

import (
	"github.com/jinzhu/gorm"
)

// CreateTestFixtures inserts all known entities into the database for testing.
func CreateTestFixtures(db *gorm.DB) {
	CreateLabelFixtures(db)
	CreateCameraFixtures(db)
	CreateCountryFixtures(db)
	CreatePhotoFixtures(db)
	CreateAlbumFixtures(db)
}
