package commands

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/capture"
)

func TestClientsListCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"ls"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsListCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "bob")
		assert.Contains(t, output, "Monitoring")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("Monitoring", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"ls", "monitoring"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsListCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "|  Scope  |")
		assert.Contains(t, output, "Monitoring")
		assert.NotContains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("CSV", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"ls", "--csv", "monitoring"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsListCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Client ID;Name;Authentication Method;User;Role;Scope;Enabled;Access Token Lifetime;Created At")
		assert.Contains(t, output, "Monitoring")
		assert.Contains(t, output, "metrics")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "alice")
	})
	t.Run("NoResult", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"ls", "notexisting"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsListCommand.Run(ctx, args...)
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
		args := []string{"ls", "--xyz", "monitoring"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsListCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
