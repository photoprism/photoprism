package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowConfigOptionsCommand(t *testing.T) {
	// Run command with test context.
	output, err := RunWithTestContext(ShowConfigOptionsCommand, []string{"config-options", "--md"})

	assert.NoError(t, err)
	assert.Contains(t, output, "PHOTOPRISM_IMPORT_PATH")
}
