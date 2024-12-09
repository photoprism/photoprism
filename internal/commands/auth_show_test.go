package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthShowCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthShowCommand, []string{"show", "sess34q3hael"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "access_token")
		assert.Contains(t, output, "Client")
	})
	t.Run("NoResult", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthShowCommand, []string{"show", "sess34qxxxxx"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
