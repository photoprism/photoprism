package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCientsRemoveCommand(t *testing.T) {
	t.Run("NoConfirmationProvided", func(t *testing.T) {
		// Run command with test context.
		output0, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		//t.Logf(output0)
		assert.NoError(t, err)
		assert.NotContains(t, output0, "not found")
		assert.Contains(t, output0, "client")

		// Run command with test context.
		output, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "cs7pvt5h8rw9aaqj"})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output2, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		//t.Logf(output2)
		assert.NoError(t, err)
		assert.NotContains(t, output2, "not found")
		assert.Contains(t, output2, "client")
	})
	t.Run("RemoveClient", func(t *testing.T) {
		// Run command with test context.
		output0, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		//t.Logf(output0)
		assert.NoError(t, err)
		assert.NotContains(t, output0, "not found")
		assert.Contains(t, output0, "client")

		// Run command with test context.
		output, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--force", "cs7pvt5h8rw9aaqj"})

		// Check command output for plausibility.
		assert.NoError(t, err)
		assert.Empty(t, output)

		// Run command with test context.
		output2, err := RunWithTestContext(ClientsShowCommand, []string{"show", "cs7pvt5h8rw9aaqj"})

		assert.Error(t, err)
		assert.Empty(t, output2)
	})
	t.Run("NotFound", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(ClientsRemoveCommand, []string{"rm", "--force", "cs7pvt5h8rw9a000"})

		// Check command output for plausibility.
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
