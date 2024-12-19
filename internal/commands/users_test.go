package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersCommand(t *testing.T) {
	t.Run("AddModifyAndRemoveJohn", func(t *testing.T) {
		// Add John
		// Run command with test context.
		output, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=John", "--email=john@test.de", "--password=test1234", "--role=admin", "john"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output2, err := RunWithTestContext(UsersShowCommand, []string{"show", "john"})

		//t.Logf(output2)
		assert.NoError(t, err)
		assert.Contains(t, output2, "John")
		assert.Contains(t, output2, "admin")
		assert.Contains(t, output2, "john@test.de")

		//Modify John
		// Run command with test context.
		output3, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=Johnny", "--email=johnnny@test.de", "--password=test12345", "john"})

		// Check command output for plausibility.
		// t.Logf(output3)
		assert.NoError(t, err)
		assert.Empty(t, output3)

		// Run command with test context.
		output4, err := RunWithTestContext(UsersShowCommand, []string{"show", "john"})

		//t.Logf(output4)
		assert.NoError(t, err)
		assert.Contains(t, output4, "Johnny")
		assert.Contains(t, output4, "admin")
		assert.Contains(t, output4, "johnnny@test.de")
		assert.Contains(t, output4, "| DeletedAt     | <nil>")

		//Remove John
		// Run command with test context.
		output5, err := RunWithTestContext(UsersRemoveCommand, []string{"rm", "--force", "john"})

		// Check command output for plausibility.
		// t.Logf(output5)
		assert.NoError(t, err)
		assert.Empty(t, output5)

		// Run command with test context.
		output6, err := RunWithTestContext(UsersShowCommand, []string{"show", "john"})

		//t.Logf(output6)
		assert.NoError(t, err)
		assert.Contains(t, output6, "Johnny")
		assert.Contains(t, output6, "admin")
		assert.Contains(t, output6, "johnnny@test.de")
		assert.Contains(t, output6, "| DeletedAt     | time.Date")
		assert.NotContains(t, output6, "| DeletedAt     | <nil>")
	})
}
