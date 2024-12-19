package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(VersionCommand, []string{"version"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "test")
	})
}
