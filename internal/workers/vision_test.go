package workers

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

// captureVisionLog replaces the package logger for the duration of a test.
func captureVisionLog(t *testing.T) *test.Hook {
	t.Helper()

	logger, hook := test.NewNullLogger()
	prev := log
	log = logger
	t.Cleanup(func() { log = prev })

	return hook
}

// captureSystemLog replaces the system logger for the duration of a test.
func captureSystemLog(t *testing.T) *test.Hook {
	t.Helper()

	logger, hook := test.NewNullLogger()
	prev := event.SystemLog
	event.SystemLog = logger
	t.Cleanup(func() { event.SystemLog = prev })

	return hook
}

// visionLogMessages returns the messages a test logger recorded.
func visionLogMessages(hook *test.Hook) (messages []string) {
	for _, entry := range hook.AllEntries() {
		messages = append(messages, entry.Message)
	}

	return messages
}

func TestVision_Start(t *testing.T) {
	t.Run("SkippedModel", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().DetectNSFW = false
		hook := captureVisionLog(t)
		system := captureSystemLog(t)

		require.NoError(t, NewVision(conf).Start("", 1, []string{vision.ModelTypeNsfw}, entity.SrcAuto, false, vision.RunManual))

		messages := visionLogMessages(hook)
		assert.Contains(t, visionLogMessages(system), "vision: skipping nsfw, because detect-nsfw is off")
		assert.NotContains(t, messages, "vision: skipping nsfw, because detect-nsfw is off")
		assert.NotContains(t, messages, "vision: no models were specified")
	})
	t.Run("NoModels", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		hook := captureVisionLog(t)

		require.NoError(t, NewVision(conf).Start("", 1, nil, entity.SrcAuto, false, vision.RunManual))
		assert.Contains(t, visionLogMessages(hook), "vision: no models were specified")
	})
}

func TestVision_RunnableModels(t *testing.T) {
	conf := config.NewMinimalTestConfig(t.TempDir())
	conf.Options().DetectNSFW = false
	worker := NewVision(conf)
	t.Run("Skipped", func(t *testing.T) {
		hook := captureVisionLog(t)
		system := captureSystemLog(t)
		assert.Equal(t, []string{vision.ModelTypeLabels}, worker.RunnableModels([]string{vision.ModelTypeNsfw, vision.ModelTypeLabels}, vision.RunManual))
		assert.Empty(t, visionLogMessages(hook))
		assert.Equal(t, []string{"vision: skipping nsfw, because detect-nsfw is off"}, visionLogMessages(system))
	})
	t.Run("None", func(t *testing.T) {
		hook := captureVisionLog(t)
		assert.Empty(t, worker.RunnableModels(nil, vision.RunManual))
		assert.Empty(t, visionLogMessages(hook))
	})
}
