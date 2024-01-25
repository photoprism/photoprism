package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestAuthRemoveCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		var err error

		ctx0 := NewTestContext([]string{"show", "sessgh6123yt"})

		output0 := capture.Output(func() {
			err = AuthShowCommand.Run(ctx0)
		})

		//t.Logf(output0)
		assert.NoError(t, err)
		assert.NotEmpty(t, output0)

		// Create test context with flags and arguments.
		ctx := NewTestContext([]string{"rm", "sessgh6123yt"})

		// Run command with test context.
		output := capture.Output(func() {
			err = AuthRemoveCommand.Run(ctx)
		})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		output1 := capture.Output(func() {
			err = AuthShowCommand.Run(ctx0)
		})

		//t.Logf(output1)
		assert.NoError(t, err)
		assert.NotEmpty(t, output1)
	})
}
