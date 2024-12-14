package testextras

import (
	"gorm.io/gorm"
)

func MigrateTestExtras(db *gorm.DB) {
	if err := db.AutoMigrate(&TestLog{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&TestDBMutex{}); err != nil {
		panic(err)
	}
}
