package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestUsersModCommand(t *testing.T) {
	t.Run("ModNotExistingUser", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"mod", "--name=New", "--email=new@test.de", "uqxqg7i1kperxxx0"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersModCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("ModDeletedUser", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"mod", "--name=New", "--email=new@test.de", "deleted"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersModCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}
