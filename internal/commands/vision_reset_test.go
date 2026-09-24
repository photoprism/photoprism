package commands

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestVisionResetCommand(t *testing.T) {
	t.Run("NoTerminal", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", "")

		// Without a terminal the prompt cannot run, which is a usage error rather than a refusal.
		_, err := RunWithTestContext(VisionResetCommand, []string{"reset", "--models=caption", "uid:ps6sg6be2lvl0yh0"})

		assertExitCode(t, err, 2)
		assert.Contains(t, err.Error(), "--yes")
	})
	t.Run("SharedYesFlag", func(t *testing.T) {
		found := false

		for _, f := range VisionResetCommand.Flags {
			if b, ok := f.(*cli.BoolFlag); ok && b.Name == "yes" {
				found = true
				assert.Equal(t, YesFlag().Usage, b.Usage)
				assert.Equal(t, YesFlag().Aliases, b.Aliases)
			}
		}

		assert.True(t, found)
	})
	t.Run("InvalidSourceBeforePrompt", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", "")

		_, err := RunWithTestContext(VisionResetCommand, []string{"reset", "--models=caption", "--source=invalid-source", "uid:ps6sg6be2lvl0yh0"})

		require.Error(t, err)
		assert.NotContains(t, err.Error(), "--yes", "an invalid source is reported before the prompt")
	})
	t.Run("ResetCaptionAndLabels", func(t *testing.T) {
		fixture := entity.PhotoFixtures.Get("VisionResetTarget")

		args := []string{
			"reset",
			"--models=caption,labels",
			"--source=ollama",
			"--yes",
			fmt.Sprintf("uid:%s", fixture.PhotoUID),
		}

		if output, err := RunWithTestContext(VisionResetCommand, args); err != nil {
			t.Fatalf("%T: %v", err, err)
		} else {
			assert.Empty(t, output)
		}
	})
}
