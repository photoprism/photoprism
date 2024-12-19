package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowVideoSizesCommand(t *testing.T) {
	// Run command with test context.
	output, err := RunWithTestContext(ShowVideoSizesCommand, []string{})

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, output, "3840")
	assert.Contains(t, output, "7680")
	assert.Contains(t, output, "4K Ultra HD")
}
