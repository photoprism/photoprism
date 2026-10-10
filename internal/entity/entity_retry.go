package entity

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/mattn/go-sqlite3"
)

const (
	deadlockRetryAttempts = 3
	deadlockRetryDelay    = 25 * time.Millisecond
	lockRetryAttempts     = 2
)

// RetryDeadlock retries transactional writes with bounded backoff.
func RetryDeadlock(action string, fn func() error) error {
	return retryWrite(action, "deadlock", deadlockRetryAttempts, isDeadlockError, fn)
}

// RetryLock retries a write once if it could not acquire a lock, including deadlocks.
// A single retry bounds the time callers wait, as a lock wait timeout can take many seconds.
func RetryLock(action string, fn func() error) error {
	return retryWrite(action, "lock", lockRetryAttempts, isLockError, fn)
}

// retryWrite runs fn until it succeeds, returns an error that is not retryable, or the attempts are used up.
func retryWrite(action, reason string, attempts int, retryable func(error) bool, fn func() error) (err error) {
	attempts = max(attempts, 1)

	for attempt := range attempts {
		err = fn()
		if err == nil || !retryable(err) {
			return err
		}
		if attempt+1 == attempts {
			break
		}
		wait := deadlockRetryDelay * time.Duration(attempt+1)
		log.Warnf("sql: %s %s (attempt %d/%d): %s", action, reason, attempt+1, attempts, err)
		time.Sleep(wait)
	}
	return err
}

// isDeadlockError identifies retryable database lock errors.
func isDeadlockError(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1213 {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "deadlock")
}

// isLockError identifies writes that failed because a lock could not be acquired, based on the driver error:
// a deadlock or lock wait timeout on MariaDB, or a busy or locked SQLite database.
func isLockError(err error) bool {
	if err == nil {
		return false
	} else if isDeadlockError(err) {
		return true
	}
	// GORM collects multiple errors in a list that cannot be unwrapped.
	var gormErrs gorm.Errors
	if errors.As(err, &gormErrs) {
		return slices.ContainsFunc(gormErrs, isLockError)
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1205
	}
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked
	}
	return false
}
