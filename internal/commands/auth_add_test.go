package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthAddCommand(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthAddCommand, []string{"add", "--scope=metrics", "--expires=5000", "--name=alice", "alice"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "App Password")
		assert.NotContains(t, output, "Access Token")
		assert.Contains(t, output, "metrics")
	})
	t.Run("NoUser", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthAddCommand, []string{"add", "--scope=test", "--expires=5000", "--name=xyz"})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.NotContains(t, output, "App Password")
		assert.Contains(t, output, "Access Token")
		assert.Contains(t, output, "test")
	})
	t.Run("UserNotFound", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthAddCommand, []string{"add", "--scope=test", "--expires=5000", "xxxxx"})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)

	})
	t.Run("NoClientName", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthAddCommand, []string{"add", "--scope=test", "--expires=5000", "alice"})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("NoScope", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthAddCommand, []string{"add", "--name=test", "--expires=5000", "alice"})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
