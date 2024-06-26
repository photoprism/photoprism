package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/capture"
)

func TestShowConfigOptionsCommand(t *testing.T) {
	var err error

	ctx := NewTestContext(nil)

	output := capture.Output(func() {
		err = ShowConfigOptionsCommand.Run(ctx)
	})

	assert.NoError(t, err)
	assert.Contains(t, output, "PHOTOPRISM_IMPORT_PATH")
	assert.Contains(t, output, "--sidecar-path")
	assert.Contains(t, output, "sidecar `PATH` *optional*")
}
