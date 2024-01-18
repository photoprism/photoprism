package commands

import (
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUsersShowCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"show", "alice"})

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersShowCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "admin")
		assert.Contains(t, output, "alice@example.com")
	})
	t.Run("NoResult", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"show", "notexisting"})

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersShowCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
