package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/capture"
)

func TestAuthListCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	output := capture.Output(func() {
		err = AuthListCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	t.Logf(output)

	// Check the command output for plausibility.
	assert.Contains(t, output, "alice")
	assert.Contains(t, output, "bob")
	assert.Contains(t, output, "visitor")
}
