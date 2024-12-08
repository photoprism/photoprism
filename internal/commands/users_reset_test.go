package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestUsersResetCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args0 := []string{"ls"}
		ctx0 := NewTestContext(args0)

		// Run command with test context.
		output0 := capture.Output(func() {
			err = UsersListCommand.Run(ctx0, args0...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output0, "alice")
		assert.Contains(t, output0, "bob")

		// Create test context with flags and arguments.
		args := []string{"reset"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersResetCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output1 := capture.Output(func() {
			err = UsersListCommand.Run(ctx0, args0...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output1, "alice")
		assert.Contains(t, output1, "bob")
	})
}
