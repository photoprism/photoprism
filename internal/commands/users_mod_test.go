package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersModCommand(t *testing.T) {
	t.Run("ModNotExistingUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=New", "--email=new@test.de", "uqxqg7i1kperxxx0"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("ModDeletedUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersModCommand, []string{"mod", "--name=New", "--email=new@test.de", "deleted"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
