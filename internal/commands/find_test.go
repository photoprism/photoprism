package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(FindCommand, []string{"find", "--csv"})

		// Check command output for plausibility.
		//t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "File Name;Mime Type;")
	})
}
