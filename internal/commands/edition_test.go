package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestEditionCommand(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"edition"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = EditionCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "ce")
	})
}
