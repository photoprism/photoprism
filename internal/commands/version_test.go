package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var err error

		// Create test context with flags and arguments.
		args := []string{"version"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = VersionCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "test")
	})
}
