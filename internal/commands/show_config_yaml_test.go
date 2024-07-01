package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestShowConfigYamlCommand(t *testing.T) {
	var err error

	ctx := NewTestContext([]string{"config-yaml", "--md"})

	output := capture.Stdout(func() {
		err = ShowConfigYamlCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, output, "ImportPath")
	assert.Contains(t, output, "--sidecar-path")
}
