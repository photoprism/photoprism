package commands

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestShowConfigOptionsCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	output := capture.Output(func() {
		err = ShowConfigOptionsCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, output, "PHOTOPRISM_IMPORT_PATH")
	assert.Contains(t, output, "--sidecar-path")
	assert.Contains(t, output, "sidecar `PATH` *optional*")
}
