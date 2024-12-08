package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestUsersAddCommand(t *testing.T) {
	t.Run("AddUserThatAlreadyExists", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"add", "--name=Alice", "--email=jane@test.de", "--password=test1234", "--role=admin", "alice"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersAddCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)

	})
	t.Run("AddDeletedUser", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"add", "--name=deleted", "--password=test1234", "deleted"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersAddCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)

	})
	t.Run("AddUsernameMissing", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"add", "--name=noname", "--password=test1234", "/##"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersAddCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
