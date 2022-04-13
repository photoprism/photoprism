package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/capture"
)

func TestShowConfigCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	output := capture.Output(func() {
		err = ShowConfigCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	// Expected config command output.
	assert.Contains(t, output, "config-file")
	assert.Contains(t, output, "darktable-cli")
	assert.Contains(t, output, "originals-path")
	assert.Contains(t, output, "import-path")
	assert.Contains(t, output, "cache-path")
	assert.Contains(t, output, "assets-path")
}

func TestShowFiltersCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	output := capture.Output(func() {
		err = ShowFiltersCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	// Expected config command output.
	assert.Contains(t, output, "landscape")
	assert.Contains(t, output, "live")
	assert.Contains(t, output, "Examples")
	assert.Contains(t, output, "Filter")
	assert.Contains(t, output, "Notes")
}

func TestShowFormatsCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	output := capture.Output(func() {
		err = ShowFormatsCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	// Expected config command output.
	assert.Contains(t, output, "JPEG")
	assert.Contains(t, output, "MP4")
	assert.Contains(t, output, "Image")
	assert.Contains(t, output, "Format")
	assert.Contains(t, output, "Description")
}
