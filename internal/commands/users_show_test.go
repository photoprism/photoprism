package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersShowCommand(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersShowCommand, []string{"show", "alice"})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Alice")
		assert.Contains(t, output, "admin")
		assert.Contains(t, output, "alice@example.com")
	})
	t.Run("NoResult", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersShowCommand, []string{"show", "notexisting"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
