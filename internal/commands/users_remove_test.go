package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestUsersRemoveCommand(t *testing.T) {
	t.Run("RemoveNotExistingUser", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"rm", "uqxqg7i1kperxxx0"})

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersRemoveCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("RemoveDeletedUser", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"rm", "deleted"})

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersRemoveCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
