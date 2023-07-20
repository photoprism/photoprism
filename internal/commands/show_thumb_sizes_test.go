package commands

import (
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShowThumbSizesCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	output := capture.Output(func() {
		err = ShowThumbSizesCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, output, "fit_2048")
	assert.Contains(t, output, "Mosaic View")
	assert.Contains(t, output, "Color Detection")
}
