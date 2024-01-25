package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestAuthAddCommand(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"add", "--scope=metrics", "--expires=5000", "--name=alice", "alice"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthAddCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "App Password")
		assert.NotContains(t, output, "Access Token")
		assert.Contains(t, output, "metrics")
	})
	t.Run("NoUser", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"add", "--scope=test", "--expires=5000", "--name=xyz"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthAddCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.NotContains(t, output, "App Password")
		assert.Contains(t, output, "Access Token")
		assert.Contains(t, output, "test")
	})
	t.Run("UserNotFound", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"add", "--scope=test", "--expires=5000", "xxxxx"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthAddCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)

	})
	t.Run("NoClientName", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"add", "--scope=test", "--expires=5000", "alice"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthAddCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("NoScope", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"add", "--name=test", "--expires=5000", "alice"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthAddCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
