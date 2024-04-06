package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestCientsRemoveCommand(t *testing.T) {
	t.Run("NoConfirmationProvided", func(t *testing.T) {
		var err error

		ctx0 := NewTestContext([]string{"show", "cs7pvt5h8rw9aaqj"})

		output0 := capture.Output(func() {
			err = ClientsShowCommand.Run(ctx0)
		})

		//t.Logf(output0)
		assert.NoError(t, err)
		assert.NotContains(t, output0, "not found")
		assert.Contains(t, output0, "client")

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"rm", "cs7pvt5h8rw9aaqj"})

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsRemoveCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		ctx2 := NewTestContext([]string{"show", "cs7pvt5h8rw9aaqj"})

		output2 := capture.Output(func() {
			err = ClientsShowCommand.Run(ctx2)
		})

		//t.Logf(output2)
		assert.NoError(t, err)
		assert.NotContains(t, output2, "not found")
		assert.Contains(t, output2, "client")
	})
	t.Run("RemoveClient", func(t *testing.T) {
		var err error

		ctx0 := NewTestContext([]string{"show", "cs7pvt5h8rw9aaqj"})

		output0 := capture.Output(func() {
			err = ClientsShowCommand.Run(ctx0)
		})

		//t.Logf(output0)
		assert.NoError(t, err)
		assert.NotContains(t, output0, "not found")
		assert.Contains(t, output0, "client")

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"rm", "--force", "cs7pvt5h8rw9aaqj"})

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsRemoveCommand.Run(ctx)
		})

		// Check command output for plausibility.
		assert.NoError(t, err)
		assert.Empty(t, output)

		ctx2 := NewTestContext([]string{"show", "cs7pvt5h8rw9aaqj"})

		output2 := capture.Output(func() {
			err = ClientsShowCommand.Run(ctx2)
		})

		assert.Error(t, err)
		assert.Empty(t, output2)
	})
	t.Run("NotFound", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"rm", "--force", "cs7pvt5h8rw9a000"})

		// Run command with test context.
		output := capture.Output(func() {
			err = ClientsRemoveCommand.Run(ctx)
		})

		// Check command output for plausibility.
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
