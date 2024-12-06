package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthListCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthListCommand, []string{"ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice ")
		assert.Contains(t, output, "bob ")
		assert.Contains(t, output, "visitor ")
	})
	t.Run("Alice", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthListCommand, []string{"ls", "alice"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Session ID")
		assert.Contains(t, output, "alice ")
		assert.NotContains(t, output, "bob ")
		assert.NotContains(t, output, "visitor ")
		assert.NotContains(t, output, "| Preview Token |")
	})
	t.Run("CSV", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthListCommand, []string{"ls", "--csv", "alice"})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Session ID;")
		assert.Contains(t, output, "alice;")
		assert.NotContains(t, output, "bob;")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("Tokens", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthListCommand, []string{"ls", "--tokens", "alice"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "|  Session ID  |")
		assert.Contains(t, output, "| Preview Token |")
		assert.Contains(t, output, "alice ")
		assert.NotContains(t, output, "bob ")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("NoResult", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthListCommand, []string{"ls", "--tokens", "notexisting"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)
	})
	t.Run("Error", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthListCommand, []string{"ls", "--xyz", "alice"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Empty(t, output)
		assert.Error(t, err)
	})
}
