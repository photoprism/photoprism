package entity

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRetryDeadlock checks bounded recovery and preservation of non-retryable errors.
func TestRetryDeadlock(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		attempts := 0
		require.NoError(t, RetryDeadlock("test", func() error { attempts++; return nil }))
		assert.Equal(t, 1, attempts)
	})
	t.Run("Transient", func(t *testing.T) {
		attempts := 0
		require.NoError(t, RetryDeadlock("test", func() error {
			attempts++
			if attempts < 3 {
				return &mysql.MySQLError{Number: 1213, Message: "retry control"}
			}
			return nil
		}))
		assert.Equal(t, 3, attempts)
	})
	t.Run("Exhausted", func(t *testing.T) {
		attempts := 0
		expected := &mysql.MySQLError{Number: 1213, Message: "retry control"}
		err := RetryDeadlock("test", func() error { attempts++; return expected })
		assert.Same(t, expected, err)
		assert.Equal(t, 3, attempts)
	})
	t.Run("OtherError", func(t *testing.T) {
		attempts := 0
		expected := errors.New("write control")
		err := RetryDeadlock("test", func() error { attempts++; return expected })
		assert.Same(t, expected, err)
		assert.Equal(t, 1, attempts)
	})
}

// TestIsDeadlockError checks driver, wrapped, and compatible textual errors.
func TestIsDeadlockError(t *testing.T) {
	for _, tc := range []struct {
		name  string
		err   error
		retry bool
	}{
		{"Nil", nil, false},
		{"Driver", &mysql.MySQLError{Number: 1213, Message: "retry control"}, true},
		{"Wrapped", fmt.Errorf("wrapped: %w", &mysql.MySQLError{Number: 1213, Message: "retry control"}), true},
		{"Text", errors.New("DEADLOCK found"), true},
		{"OtherDriver", &mysql.MySQLError{Number: 1062, Message: "duplicate"}, false},
		{"Other", errors.New("write control"), false},
	} {
		t.Run(tc.name, func(t *testing.T) { assert.Equal(t, tc.retry, isDeadlockError(tc.err)) })
	}
}

// TestRetryLock checks that writes are retried after lock errors only.
func TestRetryLock(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		attempts := 0
		require.NoError(t, RetryLock("test", func() error { attempts++; return nil }))
		assert.Equal(t, 1, attempts)
	})
	t.Run("Transient", func(t *testing.T) {
		attempts := 0
		require.NoError(t, RetryLock("test", func() error {
			attempts++
			if attempts < 2 {
				return sqlite3.Error{Code: sqlite3.ErrBusy}
			}
			return nil
		}))
		assert.Equal(t, 2, attempts)
	})
	t.Run("Exhausted", func(t *testing.T) {
		attempts := 0
		expected := &mysql.MySQLError{Number: 1205, Message: "Lock wait timeout exceeded"}
		err := RetryLock("test", func() error { attempts++; return expected })
		assert.Same(t, expected, err)
		assert.Equal(t, 2, attempts)
	})
	t.Run("OtherError", func(t *testing.T) {
		attempts := 0
		expected := errors.New("write control")
		err := RetryLock("test", func() error { attempts++; return expected })
		assert.Same(t, expected, err)
		assert.Equal(t, 1, attempts)
	})
}

// TestIsLockError checks driver, wrapped, and textual lock errors.
func TestIsLockError(t *testing.T) {
	for _, tc := range []struct {
		name  string
		err   error
		retry bool
	}{
		{"Nil", nil, false},
		{"Deadlock", &mysql.MySQLError{Number: 1213, Message: "retry control"}, true},
		{"LockWaitTimeout", &mysql.MySQLError{Number: 1205, Message: "retry control"}, true},
		{"Wrapped", fmt.Errorf("wrapped: %w", &mysql.MySQLError{Number: 1205, Message: "retry control"}), true},
		{"SQLiteBusy", sqlite3.Error{Code: sqlite3.ErrBusy}, true},
		{"SQLiteLocked", sqlite3.Error{Code: sqlite3.ErrLocked}, true},
		{"WrappedSQLite", fmt.Errorf("wrapped: %w", sqlite3.Error{Code: sqlite3.ErrBusy}), true},
		{"OtherSQLite", sqlite3.Error{Code: sqlite3.ErrConstraint}, false},
		{"GormErrors", gorm.Errors{errors.New("write control"), sqlite3.Error{Code: sqlite3.ErrBusy}}, true},
		{"OtherGormErrors", gorm.Errors{errors.New("write control")}, false},
		{"DeadlockText", errors.New("DEADLOCK found"), true},
		{"LockText", errors.New("database is locked"), false},
		{"OtherDriver", &mysql.MySQLError{Number: 1062, Message: "duplicate"}, false},
		{"Other", errors.New("write control"), false},
	} {
		t.Run(tc.name, func(t *testing.T) { assert.Equal(t, tc.retry, isLockError(tc.err)) })
	}
}

// TestRetryWrite checks the attempt limit and which errors are retried.
func TestRetryWrite(t *testing.T) {
	retryable := func(err error) bool { return err.Error() == "retry control" }

	t.Run("Success", func(t *testing.T) {
		attempts := 0
		require.NoError(t, retryWrite("test", "control", 4, retryable, func() error { attempts++; return nil }))
		assert.Equal(t, 1, attempts)
	})
	t.Run("Exhausted", func(t *testing.T) {
		attempts := 0
		err := retryWrite("test", "control", 4, retryable, func() error { attempts++; return errors.New("retry control") })
		assert.EqualError(t, err, "retry control")
		assert.Equal(t, 4, attempts)
	})
	t.Run("OtherError", func(t *testing.T) {
		attempts := 0
		err := retryWrite("test", "control", 4, retryable, func() error { attempts++; return errors.New("write control") })
		assert.EqualError(t, err, "write control")
		assert.Equal(t, 1, attempts)
	})
	t.Run("NoAttempts", func(t *testing.T) {
		attempts := 0
		err := retryWrite("test", "control", 0, retryable, func() error { attempts++; return errors.New("retry control") })
		assert.EqualError(t, err, "retry control")
		assert.Equal(t, 1, attempts)
	})
}
