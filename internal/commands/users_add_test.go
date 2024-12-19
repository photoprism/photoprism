package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersAddCommand(t *testing.T) {
	t.Run("AddUserThatAlreadyExists", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=Alice", "--email=jane@test.de", "--password=test1234", "--role=admin", "alice"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)

	})
	t.Run("AddDeletedUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=deleted", "--password=test1234", "deleted"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)

	})
	t.Run("AddUsernameMissing", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=noname", "--password=test1234", "/##"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
