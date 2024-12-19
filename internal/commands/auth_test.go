package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthCommands(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthCommands, []string{"auth", "ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "bob")
		assert.Contains(t, output, "visitor")
	})
	t.Run("ListAlice", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthCommands, []string{"auth", "ls", "alice"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
}
