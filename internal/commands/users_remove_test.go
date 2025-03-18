package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersRemoveCommand(t *testing.T) {
	t.Run("RemoveNotExistingUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersRemoveCommand, []string{"rm", "uqxqg7i1kperxxx0"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("RemoveDeletedUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersRemoveCommand, []string{"rm", "deleted"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
