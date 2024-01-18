package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/capture"
)

func TestUsersListCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"ls"})

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersListCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "bob")
		assert.NotContains(t, output, "Monitoring")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("Friend", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"ls", "friend"})

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersListCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "| Last Login |")
		assert.Contains(t, output, "friend")
		assert.NotContains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
	t.Run("CSV", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"ls", "--csv", "friend"})

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersListCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "UID;Username;Role;Authentication;Super Admin;Web Login;")
		assert.Contains(t, output, "friend")
		assert.Contains(t, output, "uqxqg7i1kperxvu7")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "alice")
	})
	t.Run("NoResult", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"ls", "notexisting"})

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersListCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)
	})
	t.Run("InvalidFlag", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"ls", "--xyz", "friend"})

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersListCommand.Run(ctx)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("InvalidValue", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"ls", "-n=-1", "friend"})

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersListCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
