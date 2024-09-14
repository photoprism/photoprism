package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestShowConfigOptionsCommand(t *testing.T) {
	var err error

	ctx := NewTestContext([]string{"config-options", "--md"})

	output := capture.Stdout(func() {
		err = ShowConfigOptionsCommand.Run(ctx)
	})

	assert.NoError(t, err)
	assert.Contains(t, output, "PHOTOPRISM_IMPORT_PATH")
}
