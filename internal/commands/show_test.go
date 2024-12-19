package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowConfigCommand(t *testing.T) {
	// Run command with test context.
	output, err := RunWithTestContext(ShowConfigCommand, []string{})

	if err != nil {
		t.Fatal(err)
	}

	// Check the command output for plausibility.
	assert.Contains(t, output, "config-path")
	assert.Contains(t, output, "originals-path")
	assert.Contains(t, output, "import-path")
	assert.Contains(t, output, "import-dest")
	assert.Contains(t, output, "cache-path")
	assert.Contains(t, output, "assets-path")
	assert.Contains(t, output, "darktable-cli")
}

func TestShowTagsCommand(t *testing.T) {
	// Run command with test context.
	output, err := RunWithTestContext(ShowMetadataCommand, []string{})

	if err != nil {
		t.Fatal(err)
	}

	// Check the command output for plausibility.
	assert.Contains(t, output, "Exiftool")
	assert.Contains(t, output, "Adobe XMP")
	assert.Contains(t, output, "Dublin Core")
	assert.Contains(t, output, "Title")
	assert.Contains(t, output, "Description")
}

func TestShowFiltersCommand(t *testing.T) {
	// Run command with test context.
	output, err := RunWithTestContext(ShowSearchFiltersCommand, []string{})

	if err != nil {
		t.Fatal(err)
	}

	// Check the command output for plausibility.
	assert.Contains(t, output, "landscape")
	assert.Contains(t, output, "live")
	assert.Contains(t, output, "Examples")
	assert.Contains(t, output, "Filter")
	assert.Contains(t, output, "Notes")
}

func TestShowFormatsCommand(t *testing.T) {
	// Run command with test context.
	output, err := RunWithTestContext(ShowFileFormatsCommand, []string{})

	if err != nil {
		t.Fatal(err)
	}

	// Check the command output for plausibility.
	assert.Contains(t, output, "JPEG")
	assert.Contains(t, output, "MP4")
	assert.Contains(t, output, "Image")
	assert.Contains(t, output, "Format")
	assert.Contains(t, output, "Description")
}
