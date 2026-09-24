package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestUsersRemoveCommand(t *testing.T) {
	t.Run("RemoveNotExistingUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersRemoveCommand, []string{"rm", "uqxqg7i1kperxxx0"})

		// Check command output for plausibility.
		// t.Logf(output)
		assertExitCode(t, err, 3)
		assert.Empty(t, output)
	})
	t.Run("RemoveDeletedUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersRemoveCommand, []string{"rm", "deleted"})

		// Check command output for plausibility.
		// t.Logf(output)
		assertExitCode(t, err, 3)
		assert.Empty(t, output)
	})
	t.Run("NoConfirmationProvided", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", "")

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=Keep Me", "--password=test1234", "--role=admin", "rmkeepme"})
		require.NoError(t, err)

		_, err = RunWithTestContext(UsersRemoveCommand, []string{"rm", "rmkeepme"})

		// Without a terminal the prompt cannot run, which is a usage error rather than a refusal.
		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 2, exit.ExitCode())
		assert.Contains(t, err.Error(), "--yes")

		_, err = RunWithTestContext(UsersRemoveCommand, []string{"rm", "--yes", "rmkeepme"})
		assert.NoError(t, err, "the account must still exist")
	})
	t.Run("Yes", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", "")

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=Remove Me", "--password=test1234", "--role=admin", "rmyes"})
		require.NoError(t, err)

		_, err = RunWithTestContext(UsersRemoveCommand, []string{"rm", "-y", "rmyes"})
		require.NoError(t, err)

		_, err = RunWithTestContext(UsersRemoveCommand, []string{"rm", "--yes", "rmyes"})
		assertExitCode(t, err, 3)
		assert.Contains(t, err.Error(), "already been deleted")
	})
	t.Run("ForceAlias", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", "")

		_, err := RunWithTestContext(UsersAddCommand, []string{"add", "--name=Force Me", "--password=test1234", "--role=admin", "rmforce"})
		require.NoError(t, err)

		_, err = RunWithTestContext(UsersRemoveCommand, []string{"rm", "--force", "rmforce"})
		require.NoError(t, err)

		_, err = RunWithTestContext(UsersRemoveCommand, []string{"rm", "--yes", "rmforce"})
		assertExitCode(t, err, 3)
		assert.Contains(t, err.Error(), "already been deleted")
	})
}
