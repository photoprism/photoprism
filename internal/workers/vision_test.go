package workers

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
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
	t.Run("PublishesSavedPhotos", func(t *testing.T) {
		conf := config.NewMinimalTestConfigWithDb("workers-vision-events", t.TempDir())
		uids := prepareVisionPhotos(t, conf)
		stubVisionCaption(t, nil)

		sub := subscribePhotoEvents(t, event.EntityUpdated)
		summary := captureEventsAtLog(t, sub, func(msg string) bool { return strings.HasPrefix(msg, "vision: updated 2 pictures") })

		require.NoError(t, NewVision(conf).Start("uid:"+strings.Join(uids, "|"), 10, []string{vision.ModelTypeCaption}, entity.SrcOllama, true, vision.RunManual))

		// The first save publishes at once; the second is held for the batch and published when the loop ends,
		// before face recognition and the index updates run.
		require.True(t, summary.seen)
		assert.Equal(t, 2, summary.count)

		events := receivePhotoEvents(t, sub)
		require.Len(t, events, 2)
		assert.ElementsMatch(t, uids, slices.Concat(events...))

		for _, uid := range uids {
			refreshed := entity.FindPhoto(entity.Photo{PhotoUID: uid})
			require.NotNil(t, refreshed)
			assert.Equal(t, "A generated caption", refreshed.PhotoCaption)
		}
	})
	t.Run("PublishesWhenCanceled", func(t *testing.T) {
		conf := config.NewMinimalTestConfigWithDb("workers-vision-cancel", t.TempDir())
		uids := prepareVisionPhotos(t, conf)
		calls := 0

		// Canceling while the second photo is processed makes Start return right after saving it.
		stubVisionCaption(t, func() {
			if calls++; calls == 2 {
				mutex.VisionWorker.Cancel()
			}
		})

		sub := subscribePhotoEvents(t, event.EntityUpdated)

		err := NewVision(conf).Start("uid:"+strings.Join(uids, "|"), 10, []string{vision.ModelTypeCaption}, entity.SrcOllama, true, vision.RunManual)
		require.EqualError(t, err, "vision: worker canceled")

		assert.ElementsMatch(t, uids, slices.Concat(receivePhotoEvents(t, sub)...))
	})
}

// prepareVisionPhotos places an original for two fixture photos and returns their UIDs.
func prepareVisionPhotos(t *testing.T, conf *config.Config) (uids []string) {
	t.Helper()

	for _, name := range []string{"VisionResetTarget", "Photo02"} {
		photo := entity.FindPhoto(entity.PhotoFixtures.Get(name))
		require.NotNil(t, photo)
		file, fileErr := photo.PrimaryFile()
		require.NoError(t, fileErr)
		fileName := filepath.Join(conf.OriginalsPath(), file.FileName)
		require.NoError(t, os.MkdirAll(filepath.Dir(fileName), fs.ModeDir))
		require.NoError(t, fs.Copy(filepath.Join("..", "photoprism", "testdata", "2015-02-04.jpg"), fileName, false))
		uids = append(uids, photo.PhotoUID)
	}

	return uids
}

// stubVisionCaption configures a caption model whose results are stubbed, calling before ahead of each result.
func stubVisionCaption(t *testing.T, before func()) {
	t.Helper()

	previous := vision.Config
	t.Cleanup(func() {
		vision.Config = previous
		vision.SetCaptionFunc(nil)
	})

	captionModel := &vision.Model{Type: vision.ModelTypeCaption, Engine: vision.ApiFormatOllama, Run: vision.RunManual}
	captionModel.ApplyEngineDefaults()
	vision.Config = &vision.ConfigValues{Models: vision.Models{captionModel}}
	vision.SetCaptionFunc(func(vision.Files, media.Src) (*vision.CaptionResult, *vision.Model, error) {
		if before != nil {
			before()
		}
		return &vision.CaptionResult{Text: "A generated caption", Source: captionModel.GetSource()}, captionModel, nil
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
