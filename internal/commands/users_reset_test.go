package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersResetCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		// Run command with test context.
		output0, err := RunWithTestContext(UsersListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output0, "alice")
		assert.Contains(t, output0, "bob")

		// Run command with test context.
		output, err := RunWithTestContext(UsersResetCommand, []string{"reset"})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output1, err := RunWithTestContext(UsersListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output1, "alice")
		assert.Contains(t, output1, "bob")
	})
}
