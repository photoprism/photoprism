package commands

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// TestVisionRunCommand verifies that a dry run lists only the models that would run and names the
// reason for each requested one that would not.
func TestVisionRunCommand(t *testing.T) {
	t.Run("DryRunSkipsModel", func(t *testing.T) {
		// The command logs through this package's logger, the vision worker through the shared one.
		logger, ok := log.(*logrus.Logger)
		require.True(t, ok)
		hook := test.NewLocal(logger)
		shared, ok := event.Log.(*logrus.Logger)
		require.True(t, ok)
		sharedHook := test.NewLocal(shared)

		conf := get.Config()
		detectNSFW := conf.Options().DetectNSFW

		t.Cleanup(func() {
			conf.Options().DetectNSFW = detectNSFW
			hook.Reset()
			sharedHook.Reset()
		})

		conf.Options().DetectNSFW = false

		_, err := RunWithTestContext(VisionRunCommand, []string{"run", "-m", "nsfw,labels", "--dry-run"})
		require.NoError(t, err)

		var messages []string

		for _, entry := range append(hook.AllEntries(), sharedHook.AllEntries()...) {
			messages = append(messages, entry.Message)
		}

		assert.Contains(t, messages, "vision: skipping nsfw, because detect-nsfw is off")
		assert.Contains(t, messages, `dry-run: vision run would execute models [labels] with filter="" (count=100000, source=, force=false)`)
	})
}

// TestVisionRunCommandFlags verifies that the models usage names every model type a run accepts.
func TestVisionRunCommandFlags(t *testing.T) {
	var usage string

	for _, flag := range VisionRunCommand.Flags {
		if f, ok := flag.(*cli.StringFlag); ok && f.Name == "models" {
			usage = f.Usage
		}
	}

	for _, model := range []string{"caption", "labels", "nsfw", "face"} {
		assert.Contains(t, usage, model)
	}
}
