package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestClientsModCommand(t *testing.T) {
	t.Run("ModNotExistingClient", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"mod", "--name=New", "--scope=test", "cs5cpu17n6gjxxxx"})

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsModCommand.Run(ctx)
		})

		// Check command output for plausibility.
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("DisableEnableAuth", func(t *testing.T) {
		var err error

		ctx0 := NewTestContext([]string{"show", "cs7pvt5h8rw9aaqj"})

		// Run command with test context.
		output0 := capture.Output(func() {
			err = ClientsShowCommand.Run(ctx0)
		})

		// Check command output for plausibility.
		t.Logf(output0)
		assert.NoError(t, err)
		assert.Contains(t, output0, "AuthEnabled  | true")
		assert.Contains(t, output0, "oauth2")

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"mod", "--disable", "cs7pvt5h8rw9aaqj"})

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsModCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output1 := capture.Output(func() {
			err = ClientsShowCommand.Run(ctx0)
		})

		// Check command output for plausibility.
		//t.Logf(output1)
		assert.NoError(t, err)
		assert.Contains(t, output1, "AuthEnabled  | false")

		// Create test context with flags and arguments.
		ctx1 := NewTestContext([]string{"mod", "--enable", "cs7pvt5h8rw9aaqj"})

		// Run command with test context.
		output2 := capture.Output(func() {
			err = ClientsModCommand.Run(ctx1)
		})

		// Check command output for plausibility.
		assert.NoError(t, err)
		assert.Empty(t, output2)

		// Run command with test context.
		output3 := capture.Output(func() {
			err = ClientsShowCommand.Run(ctx0)
		})

		// Check command output for plausibility.
		//t.Logf(output3)
		assert.NoError(t, err)
		assert.Contains(t, output3, "AuthEnabled  | true")
	})
	t.Run("RegenerateSecret", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"mod", "--regenerate", "cs7pvt5h8rw9aaqj"})

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsModCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "Client Secret")
	})
}
