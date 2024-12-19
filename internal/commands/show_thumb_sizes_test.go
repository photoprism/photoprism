package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowThumbSizesCommand(t *testing.T) {
	// Run command with test context.
	output, err := RunWithTestContext(ShowThumbSizesCommand, []string{})

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, output, "fit_1920")
	assert.Contains(t, output, "Mosaic View")
	assert.Contains(t, output, "Color Detection")
}
