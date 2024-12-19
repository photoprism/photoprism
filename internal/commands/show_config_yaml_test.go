package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowConfigYamlCommand(t *testing.T) {
	// Run command with test context.
	output, err := RunWithTestContext(ShowConfigYamlCommand, []string{"config-yaml", "--md"})

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, output, "ImportPath")
	assert.Contains(t, output, "--sidecar-path")
}
