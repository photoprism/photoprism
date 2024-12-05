package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestShowConfigOptionsCommand(t *testing.T) {
	var err error

	args := []string{"config-options", "--md"}
	ctx := NewTestContext(args)
	output := capture.Stdout(func() {
		err = ShowConfigOptionsCommand.Run(ctx, args...)
	})

	assert.NoError(t, err)
	assert.Contains(t, output, "PHOTOPRISM_IMPORT_PATH")
}
