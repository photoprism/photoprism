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
	assert.Contains(t, output, "PHOTOPRISM_HTTP_HEADER_TIMEOUT")
	assert.Contains(t, output, "PHOTOPRISM_HTTP_HEADER_BYTES")
	assert.Contains(t, output, "PHOTOPRISM_HTTP_IDLE_TIMEOUT")
}
