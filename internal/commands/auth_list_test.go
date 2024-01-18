package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/capture"
)

func TestAuthListCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"ls"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthListCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "bob")
		assert.Contains(t, output, "visitor")
	})
	t.Run("Alice", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"ls", "alice"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthListCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Session ID")
		assert.Contains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("CSV", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"ls", "--csv", "alice"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthListCommand.Run(ctx)
		})

		// Check command output for plausibility.
		t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Session ID;")
		assert.Contains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
}
