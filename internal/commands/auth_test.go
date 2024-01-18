package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/capture"
)

func TestAuthCommands(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"auth", "ls"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthCommands.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "bob")
		assert.Contains(t, output, "visitor")
	})
	t.Run("ListAlice", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"auth", "ls", "alice"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthCommands.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
}
