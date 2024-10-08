package testextras

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"gorm.io/gorm"
)

// Test DB Mutex structure to store the currently active mutex
type TestDBMutex struct {
	ID        uint      `gorm:"primaryKey;"`
	CreateAt  time.Time `sql:"index:idx_testdbmutex_create_at"`
	ProcessId int
}

// Attempts to acquire a database controlled mutex.  Using the table primary key to prevent more than 1 insert succeeding.
// Will retry 60 times with 10s interval, before returning false on failure to get mutex.
// The mutex uses the process id to ensure uniqueness between processes.
func LockDBMutex(db *gorm.DB, log event.Logger, caller string) bool {
	pid := os.Getpid()
	err := errors.New("so i am not nil")
	counter := 0
	for err != nil {
		record := TestDBMutex{ID: 1, CreateAt: time.Now().UTC(), ProcessId: pid}
		if err = db.Create(&record).Error; err != nil {
			counter += 1
			LogMessage(db, fmt.Sprintf("%v LockDBMutex Failed Attempt %v", caller, counter))
			if counter > 60 { // There is 10 minutes of wait time here.
				return false
			}
			time.Sleep(10 * time.Second)
		}
	}
	return true
}

// delete the mutex using the processes id.  This should be called with a defer to try and ensure that it always get cleared.
// But, if it's a really nasty internal error (eg. SIGFAULT) then go wont free the mutex and this will require manual intervention.
// The photoprism makefile tests drop the database, which will clear the mutex at the start of the testing.
func UnlockDBMutex(db *gorm.DB) {
	pid := os.Getpid()
	record := TestDBMutex{ProcessId: pid}
	db.Where("process_id = ?", pid).Delete(&record)
}

// Clears out a mutex lock and logs messages about it
func ReleaseDBMutex(db *gorm.DB, log event.Logger, caller string, code int) {
	LogMessage(db, fmt.Sprintf("%v UnlockDBMutex", caller))
	UnlockDBMutex(db)
	log.Info("database mutex released")
	LogMessage(db, fmt.Sprintf("%v ending with %v", caller, code))
}

// Opens a database connection, and then attempts to acquire a mutex for this process.
func AcquireDBMutex(log event.Logger, caller string) (dbc *DbConn, err error) {

	err = nil

	driver := os.Getenv("PHOTOPRISM_TEST_DRIVER")
	dsn := os.Getenv("PHOTOPRISM_TEST_DSN")

	// Set default test database driver.
	if driver == "test" || driver == "sqlite" || driver == "" || dsn == "" {
		driver = SQLite3
	}

	// Set default database DSN.
	if driver == SQLite3 {
		if dsn == "" {
			dsn = SQLiteMemoryDSN
		} else if dsn != SQLiteTestDB {
			// Continue.
		} else if err := os.Remove(dsn); err == nil {
			log.Debugf("sqlite: test file %s removed", clean.Log(dsn))
		}
	}

	// Create gorm.DB connection provider.
	dbc = &DbConn{
		Driver: driver,
		Dsn:    dsn,
	}

	SetDbProvider(dbc)
	log.Info("migrating test extras")
	MigrateTestExtras(dbc.Db())
	LogMessage(dbc.Db(), fmt.Sprintf("%v starting", caller))
	if LockDBMutex(dbc.Db(), log, caller) {
		LogMessage(dbc.Db(), fmt.Sprintf("%v LockDBMutex", caller))
		log.Info("database mutex acquired")

	} else {
		log.Error("Unable to get DBMutex")
		err = errors.New("unable to acquire DBMutex")
	}

	return dbc, err
}
