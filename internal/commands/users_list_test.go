package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersListCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersListCommand, []string{"ls", "--login", "--created", "--deleted", "-n", "100", "--md"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "bob")
		assert.NotContains(t, output, "Monitoring")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("LastLogin", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersListCommand, []string{"ls", "-l", "friend"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, " Last Login ")
		assert.Contains(t, output, "friend")
		assert.NotContains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("Created", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersListCommand, []string{"ls", "-a", "friend"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, " Created At ")
		assert.Contains(t, output, "friend")
		assert.NotContains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("Deleted", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersListCommand, []string{"ls", "-r", "friend"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, " Deleted At ")
		assert.Contains(t, output, "friend")
		assert.NotContains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("CSV", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersListCommand, []string{"ls", "--csv", "friend"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "UID;Username;Role;Authentication;Super Admin;Web Login;")
		assert.Contains(t, output, "friend")
		assert.Contains(t, output, "uqxqg7i1kperxvu7")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "alice")
	})
	t.Run("NoResult", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersListCommand, []string{"ls", "notexisting"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)
	})
	t.Run("OneResult", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersListCommand, []string{"ls", "--count=1", "friend"})

		// Check result.
		assert.NoError(t, err)
		assert.NotEmpty(t, output)
	})
	t.Run("InvalidFlag", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersListCommand, []string{"ls", "--xyz", "friend"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Empty(t, output)
		assert.Error(t, err)
	})
}
