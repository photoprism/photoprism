package commands

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/capture"
)

func TestAuthListCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"ls"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthListCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice ")
		assert.Contains(t, output, "bob ")
		assert.Contains(t, output, "visitor ")
	})
	t.Run("Alice", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"ls", "alice"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthListCommand.Run(ctx, args...)
		})

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
		var err error

		// Create test context with flags and arguments.
		args := []string{"ls", "--csv", "alice"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthListCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Session ID;")
		assert.Contains(t, output, "alice;")
		assert.NotContains(t, output, "bob;")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("Tokens", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"ls", "--tokens", "alice"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthListCommand.Run(ctx, args...)
		})

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
		var err error

		// Create test context with flags and arguments.
		args := []string{"ls", "--tokens", "notexisting"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthListCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)
	})
	t.Run("Error", func(t *testing.T) {
		var err error

		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("error: %s \nstack: %s", r, debug.Stack())
				assert.Error(t, err)
			}
		}()

		// Create test context with flags and arguments.
		args := []string{"ls", "--xyz", "alice"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthListCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
