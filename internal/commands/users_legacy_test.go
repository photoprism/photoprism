package commands

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestUsersLegacyCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(UsersLegacyCommand, []string{""})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		// remove spaces as this test will fail if there are records in the table due to dynamic sizing of headings
		var result strings.Builder
		result.Grow(len(output))
		for _, char := range output {
			if !unicode.IsSpace(char) {
				result.WriteRune(char)
			}
		}
		assert.Contains(t, result.String(), "│ID│UID│Name│User│Email│Admin│CreatedAt│")
	})
}
