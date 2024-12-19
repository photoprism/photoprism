package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEditionCommand(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(EditionCommand, []string{"edition"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "ce")
	})
}
