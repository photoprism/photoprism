package commands

import (
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShowConfigYamlCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	output := capture.Output(func() {
		err = ShowConfigYamlCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, output, "ImportPath")
	assert.Contains(t, output, "--sidecar-path")
}
