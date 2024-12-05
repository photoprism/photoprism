package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestPasswdCommand(t *testing.T) {
	t.Run("UserNotFound", func(t *testing.T) {
		var err error

		args := []string{"passwd", "--show", "mila"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = PasswdCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("DeletedUser", func(t *testing.T) {
		var err error

		args := []string{"passwd", "--show", "uqxqg7i1kperxvu8"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = PasswdCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		assert.Error(t, err)
		assert.Empty(t, output)
	})
	t.Run("DeletePassword", func(t *testing.T) {
		var err error

		args := []string{"passwd", "--rm", "no_local_auth"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = PasswdCommand.Run(ctx)
		})

		// Check command output for plausibility.
		assert.NoError(t, err)
		assert.Empty(t, output)
	})
}
