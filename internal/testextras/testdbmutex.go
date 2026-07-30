package testextras

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"runtime"
	"syscall"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// dbID holds the database identifier number for this instance
var dbID int

// TestDBChoice structure to assist finding the available databases
type TestDBChoice struct {
	ID uint `gorm:"primaryKey;"`
}

// TestDBMutex structure to store the currently active mutex
type TestDBMutex struct {
	ID        uint      `gorm:"primaryKey;"`
	CreateAt  time.Time `sql:"index:idx_testdbmutex_create_at"`
	ProcessId int
	Caller    string `gorm:"size:255"`
}

// LockDBMutex Attempts to acquire a database controlled mutex.  Using the table primary key to prevent more than 1 insert succeeding.
// Will retry 60 times with 10s interval, before returning false on failure to get mutex.
// The mutex uses the process id to ensure uniqueness between processes.
func LockDBMutex(db *gorm.DB, log event.Logger, caller string) (ok bool, dbNum int) {
	type Result struct {
		ID uint
	}
	var result Result
	var results []TestDBMutex

	pid := os.Getpid()
	err := errors.New("so i am not nil")
	counter := 0
	ok = false
	for err != nil {
		if len(caller) > 255 {
			caller = caller[:255]
		}

		if err = db.Model(&TestDBChoice{}).Select("test_db_choices.id").Joins("left join test_db_mutexes on test_db_choices.id = test_db_mutexes.id").Where("test_db_mutexes.id is null").Order("test_db_choices.id ASC").First(&result).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				LogMessage(db, fmt.Sprintf("%v LockDBMutex No Database Available %v", caller, counter))
				counter++
				time.Sleep(10 * time.Second)

				// Check if any of the stored process id's are no longer active...
				if dberr := db.Model(TestDBMutex{}).Where("id is not null").Find(&results).Error; dberr != nil {
					LogMessage(db, fmt.Sprintf("%v LockDBMutex Find Failed Attempt %v with %s", caller, counter, dberr.Error()))
					return ok, dbNum
				} else {
					for _, existing := range results {
						if proc, oserr := os.FindProcess(existing.ProcessId); oserr == nil {
							running := false
							if runtime.GOOS == "windows" {
								running = true
								LogMessage(db, fmt.Sprintf("Process %d is running on windows", existing.ProcessId))
							} else {
								if procerr := proc.Signal(syscall.Signal(0)); procerr == nil {
									running = true
									LogMessage(db, fmt.Sprintf("Process %d is running on *nix", existing.ProcessId))
								} else if procerr == os.ErrProcessDone {
									running = false
								} else {
									LogMessage(db, fmt.Sprintf("Unable to Signal %d due to %s", existing.ProcessId, procerr.Error()))
								}
							}
							if !running {
								if dberr := db.Where("process_id = ?", existing.ProcessId).Delete(existing); dberr.Error != nil {
									LogMessage(db, fmt.Sprintf("Unable to delete not running %d due to %s", existing.ProcessId, dberr.Error))
								} else {
									LogMessage(db, fmt.Sprintf("Cleaned up not running process id %d from DBMutex", existing.ProcessId))
								}
							}
						} else {
							LogMessage(db, fmt.Sprintf("Unable to FindProcess %d due to %s", existing.ProcessId, oserr.Error()))
						}
					}
				}
			} else {
				LogMessage(db, fmt.Sprintf("%v LockDBMutex Failed Attempt %v with %s", caller, counter, err.Error()))
				return ok, dbNum
			}
		} else {
			record := TestDBMutex{ID: result.ID, CreateAt: time.Now().UTC(), ProcessId: pid, Caller: caller}
			if err = db.Create(&record).Error; err != nil {
				// Assumption is that this will be a unique index error, because someone else got it before us...
				LogMessage(db, fmt.Sprintf("%v LockDBMutex Failed Attempt %v with %s", caller, counter, err.Error()))
			} else {
				ok = true
				dbNum = int(result.ID)
			}
		}
	}
	dbID = dbNum
	return ok, dbNum
}

// UnlockDBMutex deletes the mutex using the processes id.  This should be called with a defer to try and ensure that it always get cleared.
// But, if it's a really nasty internal error (eg. SIGFAULT) then go wont free the mutex and this will require manual intervention.
// The photoprism makefile tests drop the database, which will clear the mutex at the start of the testing.
func UnlockDBMutex(db *gorm.DB) {
	pid := os.Getpid()
	record := TestDBMutex{ProcessId: pid}
	db.Where("process_id = ?", pid).Delete(&record)
}

// ReleaseDBMutex Clears out a mutex lock and logs messages about it
func ReleaseDBMutex(db *gorm.DB, log event.Logger, caller string, code int) {
	LogMessage(db, fmt.Sprintf("%v UnlockDBMutex", caller))
	UnlockDBMutex(db)
	log.Info("database mutex released")
	LogMessage(db, fmt.Sprintf("%v ending with %v", caller, code))
}

// AcquireDBMutex Opens a database connection, and then attempts to acquire a mutex for this process.
func AcquireDBMutex(log event.Logger, caller string) (dbc *DbConn, dbn int, err error) {

	err = nil

	driver, dsname := dsn.PhotoPrismTestToDriverDSN()

	// Set default test database driver.
	if driver == "test" || driver == "sqlite" || driver == "" || dsname == "" {
		driver = SQLite3
	}

	// Set default database DSN.
	if driver == SQLite3 {
		if dsname == "" {
			dsname = SQLiteMutexDSN
			// Try to create the path, ignoring errors
			_ = os.MkdirAll("/go/src/github.com/photoprism/photoprism/storage/testdata", fs.ModePerm)
		} else if dsname != SQLiteTestDB {
			// Continue.
		} else if err := os.Remove(dsname); err == nil {
			log.Debugf("sqlite: test file %s removed", clean.Log(dsname))
		}
	}

	// Create gorm.DB connection provider.
	dbc = &DbConn{
		Driver: driver,
		Dsn:    dsname,
	}

	SetDbProvider(dbc)
	log.Info("migrating test extras")
	MigrateTestExtras(dbc.Db())
	LogMessage(dbc.Db(), fmt.Sprintf("%v starting", caller))
	if ok, n := LockDBMutex(dbc.Db(), log, caller); ok {
		LogMessage(dbc.Db(), fmt.Sprintf("%v LockDBMutex database %d acquired", caller, n))
		log.Info("database mutex acquired")
		dbn = n

		// switch driver {
		// case Postgres:
		// 	if err = ResetPostgresDB(dsn.Parse(dsname).Name, dbn); err != nil {
		// 		log.Errorf("Unable to get reset database with %v", err)
		// 	}
		// case MySQL:
		// 	if err = ResetMariaDB(dsn.Parse(dsname).Name, dbn); err != nil {
		// 		log.Errorf("Unable to get reset database with %v", err)
		// 	}
		// }
	} else {
		log.Error("Unable to get DBMutex")
		err = errors.New("unable to acquire DBMutex")
	}

	return dbc, dbn, err
}

// GetDBMutexID returns the database id that has been assigned to this process
func GetDBMutexID() int {
	return dbID
}
