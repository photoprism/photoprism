package vision

import (
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// warnMessages returns the warning messages captured by the hook.
func warnMessages(entries []*logrus.Entry) (result []string) {
	for _, entry := range entries {
		if entry.Level == logrus.WarnLevel {
			result = append(result, entry.Message)
		}
	}

	return result
}

// TestModel_WarnOnFirstRun checks that configuration warnings logged while loading are logged once more
// when the model first runs, naming only the variable.
func TestModel_WarnOnFirstRun(t *testing.T) {
	images := Files{fs.Abs("./testdata/cat_224x224.jpg")}

	t.Run("UnsetVariable", func(t *testing.T) {
		resetUnresolvedUriWarnings(t)
		resetRefusedEnvWarned(t)
		hook := captureVisionLog(t)
		model := &Model{Type: ModelTypeLabels, Name: "nasnet", Service: Service{Uri: "${VISION_TEST_UNSET_URI}"}}

		prevConfig := Config
		t.Cleanup(func() { Config = prevConfig })
		Config = &ConfigValues{Models: Models{model}, Thresholds: DefaultThresholds}

		// Loading the config resolves the endpoint.
		model.Endpoint()
		model.Endpoint()
		require.Len(t, warnMessages(hook.AllEntries()), 1)

		for i := 0; i < 3; i++ {
			_, err := labelsInternal(images, media.SrcLocal, entity.SrcImage)
			require.Error(t, err)
		}

		warnings := warnMessages(hook.AllEntries())
		require.Len(t, warnings, 2)
		assert.Equal(t, warnings[0], warnings[1])
		assert.Contains(t, warnings[1], "does not resolve")
	})
	t.Run("RefusedVariable", func(t *testing.T) {
		resetUnresolvedUriWarnings(t)
		resetRefusedEnvWarned(t)
		t.Setenv("VISION_TEST_TOKEN", "s3cr3t-value")
		hook := captureVisionLog(t)
		model := &Model{Type: ModelTypeCaption, Name: "gemma3:4b", Service: Service{Uri: "https://vision.example.com/${VISION_TEST_TOKEN}"}}

		prevConfig := Config
		t.Cleanup(func() { Config = prevConfig })
		Config = &ConfigValues{Models: Models{model}, Thresholds: DefaultThresholds}

		model.Endpoint()
		require.Len(t, warnMessages(hook.AllEntries()), 2)

		for i := 0; i < 2; i++ {
			_, _, err := captionInternal(images, media.SrcLocal)
			require.Error(t, err)
		}

		warnings := warnMessages(hook.AllEntries())
		require.Len(t, warnings, 4)
		assert.Equal(t, warnings[:2], warnings[2:])
		assert.Contains(t, strings.Join(warnings, "\n"), "VISION_TEST_TOKEN")
		assert.NotContains(t, strings.Join(warnings, "\n"), "s3cr3t-value")
	})
	t.Run("ResolvedUri", func(t *testing.T) {
		resetUnresolvedUriWarnings(t)
		resetRefusedEnvWarned(t)
		hook := captureVisionLog(t)
		model := &Model{Type: ModelTypeLabels, Name: "nasnet", Service: Service{Uri: "https://vision.example.com/api"}}

		assert.NoError(t, model.unresolvedUriErr())
		assert.NoError(t, model.unresolvedUriErr())
		assert.Empty(t, warnMessages(hook.AllEntries()))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.NoError(t, (*Model)(nil).unresolvedUriErr())
	})
}

// TestModel_WarnOnFirstRunServiceModel checks that a refused Service.Model variable is logged once more on
// the first run, whether or not the service URI resolves.
func TestModel_WarnOnFirstRunServiceModel(t *testing.T) {
	images := Files{fs.Abs("./testdata/face_160x160.jpg")}

	for _, uri := range []string{"${VISION_TEST_MISSING_URI}", "http://127.0.0.1:1/api"} {
		t.Run(uri, func(t *testing.T) {
			resetUnresolvedUriWarnings(t)
			resetRefusedEnvWarned(t)
			t.Setenv("VISION_TEST_TOKEN", "s3cr3t-value")
			t.Setenv("OLLAMA_API_KEY", "")
			hook := captureVisionLog(t)
			model := &Model{Type: ModelTypeLabels, Name: "x", Engine: "ollama", Service: Service{Uri: uri, Model: "${VISION_TEST_TOKEN}"}}
			model.ApplyEngineDefaults()

			prevConfig := Config
			t.Cleanup(func() { Config = prevConfig })
			Config = &ConfigValues{Models: Models{model}, Thresholds: DefaultThresholds}

			model.GetModel()

			for i := 0; i < 3; i++ {
				_, err := labelsInternal(images, media.SrcLocal, entity.SrcImage)
				require.Error(t, err)
			}

			count := 0
			for _, msg := range warnMessages(hook.AllEntries()) {
				assert.NotContains(t, msg, "s3cr3t-value")
				if strings.Contains(msg, "Service.Model does not expand VISION_TEST_TOKEN") {
					count++
				}
			}

			assert.Equal(t, 2, count)
		})
	}
}

// TestModel_WarnKey checks that the key changes with the fields that identify the model's warnings.
func TestModel_WarnKey(t *testing.T) {
	model := &Model{Type: ModelTypeLabels, Name: "a", Model: "b", Service: Service{Uri: "c"}}
	assert.Equal(t, "labels/a/b/c", model.warnKey())
	other := &Model{Type: ModelTypeLabels, Name: "a", Model: "b", Service: Service{Uri: "d"}}
	assert.NotEqual(t, model.warnKey(), other.warnKey())
}

// TestModel_WarnOnFirstRunConcurrent checks that concurrent first runs log a warning only once.
func TestModel_WarnOnFirstRunConcurrent(t *testing.T) {
	resetUnresolvedUriWarnings(t)
	resetRefusedEnvWarned(t)
	hook := captureVisionLog(t)
	model := &Model{Type: ModelTypeLabels, Name: "x", Service: Service{Uri: "${VISION_TEST_MISSING_URI}"}}

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = model.unresolvedUriErr()
		}()
	}
	wg.Wait()

	assert.Len(t, warnMessages(hook.AllEntries()), 1)
}

// captureAuditLog replaces the audit logger with one whose entries are captured by the returned hook.
func captureAuditLog(t *testing.T) *test.Hook {
	t.Helper()

	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	previous := event.AuditLog
	event.AuditLog = logger
	t.Cleanup(func() { event.AuditLog = previous })

	return hook
}

// TestWarnModel checks that model configuration warnings go to the audit log.
func TestWarnModel(t *testing.T) {
	t.Run("AuditLog", func(t *testing.T) {
		hook := captureVisionLog(t)
		audit := captureAuditLog(t)

		warnModel("%s model %s needs a service uri", "nsfw", "x")

		assert.Empty(t, warnMessages(hook.AllEntries()))
		require.Len(t, warnMessages(audit.AllEntries()), 1)
		assert.Equal(t, "audit: vision › nsfw model x needs a service uri", warnMessages(audit.AllEntries())[0])
	})
}

// TestModel_WarnOnFirstRunAuditLog checks that the warnings of a model, including those repeated on its
// first run, go only to the audit log.
func TestModel_WarnOnFirstRunAuditLog(t *testing.T) {
	images := Files{fs.Abs("./testdata/cat_224x224.jpg")}

	resetUnresolvedUriWarnings(t)
	resetRefusedEnvWarned(t)
	t.Setenv("VISION_TEST_TOKEN", "s3cr3t-value")
	hook := captureVisionLog(t)
	audit := captureAuditLog(t)
	model := &Model{Type: ModelTypeNsfw, Name: "nsfw-test", Service: Service{Uri: "https://vision.example.com/${VISION_TEST_TOKEN}"}}

	prevConfig := Config
	t.Cleanup(func() { Config = prevConfig })
	Config = &ConfigValues{Models: Models{model}, Thresholds: DefaultThresholds}

	model.Endpoint()

	for i := 0; i < 2; i++ {
		_, err := nsfwInternalContext(images, media.SrcLocal, nsfwThresholdIndex)
		require.Error(t, err)
	}

	assert.Empty(t, warnMessages(hook.AllEntries()))
	warnings := warnMessages(audit.AllEntries())
	require.Len(t, warnings, 4)
	assert.Contains(t, strings.Join(warnings, "\n"), "VISION_TEST_TOKEN")
	assert.Contains(t, strings.Join(warnings, "\n"), "does not resolve")
	assert.NotContains(t, strings.Join(warnings, "\n"), "s3cr3t-value")
}

// TestModel_NsfwModelAuditLog checks that an NSFW model initialization warning goes to the audit log.
func TestModel_NsfwModelAuditLog(t *testing.T) {
	hook := captureVisionLog(t)
	audit := captureAuditLog(t)
	model := &Model{Type: ModelTypeNsfw, Name: "custom-nsfw", Reduction: nsfw.ReductionSoftmaxUnsafe}

	assert.Nil(t, model.NsfwModel())
	assert.Empty(t, warnMessages(hook.AllEntries()))
	require.Len(t, warnMessages(audit.AllEntries()), 1)
	assert.Contains(t, warnMessages(audit.AllEntries())[0], "unsafe class index is required")
}
