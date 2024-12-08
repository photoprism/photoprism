package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestAuthRemoveCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		var err error

		args0 := []string{"show", "sessgh6123yt"}
		ctx0 := NewTestContext(args0)

		output0 := capture.Output(func() {
			err = AuthShowCommand.Run(ctx0, args0...)
		})

		//t.Logf(output0)
		assert.NoError(t, err)
		assert.NotEmpty(t, output0)

		// Create test context with flags and arguments.
		args := []string{"rm", "sessgh6123yt"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthRemoveCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		output1 := capture.Output(func() {
			err = AuthShowCommand.Run(ctx0, args0...)
		})

		//t.Logf(output1)
		assert.NoError(t, err)
		assert.NotEmpty(t, output1)
	})
}
