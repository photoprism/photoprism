package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswdCommand(t *testing.T) {
	t.Run("UserNotFound", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(PasswdCommand, []string{"passwd", "--show", "mila"})

		// Check command output for plausibility.
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("DeletedUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(PasswdCommand, []string{"passwd", "--show", "uqxqg7i1kperxvu8"})

		// Check command output for plausibility.
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("DeletePassword", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(PasswdCommand, []string{"passwd", "--rm", "no_local_auth"})

		// Check command output for plausibility.
		assert.NoError(t, err)
		assert.Empty(t, output)
	})
}
