package entity

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
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
