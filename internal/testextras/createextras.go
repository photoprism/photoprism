package testextras

import (
	"runtime"
	"time"

	"gorm.io/gorm"
)

// MigrateTestExtras ensures that the tables needed for test extras are in place, and populated as needed.
func MigrateTestExtras(db *gorm.DB) {
	var err error
	for migrateRetry := range 10 {
		if err = db.AutoMigrate(&TestLog{}); err != nil {
			log.Warnf("migratetestextras: automigrate testlog encountered %+v on loop %d", err, migrateRetry)
			if migrateRetry < 10 {
				time.Sleep(time.Second * 5)
			}
		} else {
			break
		}
	}
	if err != nil {
		panic(err)
	}
	for migrateRetry := range 10 {
		if err = db.AutoMigrate(&TestDBMutex{}); err != nil {
			log.Warnf("migratetestextras: automigrate testdbmutex encountered %+v on loop %d", err, migrateRetry)
			if migrateRetry < 10 {
				time.Sleep(time.Second * 5)
			}
		} else {
			break
		}
	}
	if err != nil {
		panic(err)
	}
	for migrateRetry := range 10 {
		if err = db.AutoMigrate(&TestDBChoice{}); err != nil {
			log.Warnf("migratetestextras: automigrate testdbchoice encountered %+v on loop %d", err, migrateRetry)
			if migrateRetry < 10 {
				time.Sleep(time.Second * 5)
			}
		} else {
			break
		}
	}
	if err != nil {
		panic(err)
	}
	// Populate the choice table
	dbCount := uint(8)
	if cpus := runtime.NumCPU(); cpus > 0 {
		dbCount = uint(cpus)
	}
	for c := uint(1); c <= dbCount; c++ {
		var result TestDBChoice
		db.FirstOrCreate(&result, TestDBChoice{ID: c})
	}
}
