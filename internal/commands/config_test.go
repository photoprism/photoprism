package commands

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestConfigCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	output := util.CaptureOutput(func() {
		err = ConfigCommand.Run(ctx)
	})

	assert.Contains(t, output, "NAME                  VALUE")
	assert.Contains(t, output, "config-file")
	assert.Contains(t, output, "darktable-cli")
	assert.Contains(t, output, "originals-path")
	assert.Contains(t, output, "import-path")
	assert.Contains(t, output, "export-path")
	assert.Contains(t, output, "cache-path")
	assert.Contains(t, output, "assets-path")

	assert.Equal(t, output, output)
	assert.Nil(t, err)
}
