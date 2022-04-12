package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/capture"
)

func TestConfigCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	output := capture.Output(func() {
		err = ShowConfigCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	// Expected config command output.
	assert.Contains(t, output, "NAME                      VALUE")
	assert.Contains(t, output, "config-file")
	assert.Contains(t, output, "darktable-cli")
	assert.Contains(t, output, "originals-path")
	assert.Contains(t, output, "import-path")
	assert.Contains(t, output, "cache-path")
	assert.Contains(t, output, "assets-path")
}
