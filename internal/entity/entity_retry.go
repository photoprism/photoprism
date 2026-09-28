package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	deadlockRetryAttempts = 3
	deadlockRetryDelay    = 25 * time.Millisecond
)

// RetryDeadlock retries transactional writes with bounded backoff.
func RetryDeadlock(action string, fn func() error) (err error) {
	for attempt := range deadlockRetryAttempts {
		err = fn()
		if err == nil || !isDeadlockError(err) {
			return err
		}
		wait := deadlockRetryDelay * time.Duration(attempt+1)
		log.Warnf("sql: %s deadlock (attempt %d/%d): %s", action, attempt+1, deadlockRetryAttempts, err)
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
