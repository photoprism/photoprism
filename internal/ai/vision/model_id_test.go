package vision

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/pkg/clean"
)

// resetClippedModelIdWarnings clears the logged identifiers before and after a test.
func resetClippedModelIdWarnings(t *testing.T) {
	t.Helper()
	clippedModelIdWarned.Clear()
	t.Cleanup(clippedModelIdWarned.Clear)
}

// clipWarnings returns the system log warnings about shortened model identifiers.
func clipWarnings(hook *logtest.Hook) []string {
	var result []string

	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.WarnLevel && strings.Contains(entry.Message, "model identifier exceeds") {
			result = append(result, entry.Message)
		}
	}

	return result
}

func TestCleanModelId(t *testing.T) {
	maxId := "org/" + strings.Repeat("a", clean.LengthType-4)
	require.Len(t, maxId, clean.LengthType)

	t.Run("Success", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		_, hook := captureLogs(t)
		assert.Equal(t, "unsloth/Qwen3.5-9B-GGUF:Q4_K_M", cleanModelId(" unsloth/Qwen3.5-9B-GGUF:Q4_K_M "))
		assert.Empty(t, clipWarnings(hook))
	})
	t.Run("Empty", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		_, hook := captureLogs(t)
		assert.Equal(t, "", cleanModelId(""))
		assert.Equal(t, "", cleanModelId("   "))
		assert.Empty(t, clipWarnings(hook))
	})
	t.Run("MaxLength", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		_, hook := captureLogs(t)
		assert.Equal(t, maxId, cleanModelId(maxId))
		assert.Equal(t, maxId, cleanModelId(maxId+"  "))
		assert.Equal(t, maxId, cleanModelId(maxId+"ä"))
		assert.Empty(t, clipWarnings(hook))
	})
	t.Run("ClippedWarnsOnce", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		logHook, hook := captureLogs(t)
		assert.Equal(t, maxId, cleanModelId(maxId+":Q4_K_M"))
		assert.Equal(t, maxId, cleanModelId(maxId+":Q4_K_M"))
		assert.Equal(t, maxId, cleanModelId(maxId+":Q8_0"))
		warnings := clipWarnings(hook)
		require.Len(t, warnings, 1)
		assert.Equal(t, "vision: model identifier exceeds 64 characters, using "+clean.Log(maxId), warnings[0])
		assert.Empty(t, logHook.AllEntries())
	})
	t.Run("ClippedWarnsPerId", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		_, hook := captureLogs(t)
		otherId := "other/" + strings.Repeat("b", clean.LengthType)
		cleanModelId(maxId + ":Q4_K_M")
		cleanModelId(otherId)
		warnings := clipWarnings(hook)
		require.Len(t, warnings, 2)
		assert.Contains(t, warnings[0], clean.Log(maxId))
		assert.Contains(t, warnings[1], clean.Log(otherId[:clean.LengthType]))
	})
	t.Run("SanitizedLog", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		logHook, hook := captureLogs(t)
		id := cleanModelId("org/model\nsecond line " + strings.Repeat("c", clean.LengthType))
		require.Len(t, id, clean.LengthType)
		warnings := clipWarnings(hook)
		require.Len(t, warnings, 1)
		assert.NotContains(t, warnings[0], "\n")
		assert.Equal(t, "vision: model identifier exceeds 64 characters, using "+clean.Log(id), warnings[0])
		assert.Empty(t, logHook.AllEntries())
	})
}

func TestModel_GetModelClipped(t *testing.T) {
	longId := "org/" + strings.Repeat("a", clean.LengthType-4) + ":Q4_K_M"
	clippedId := longId[:clean.LengthType]

	t.Run("Model", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		_, hook := captureLogs(t)
		model, name, version := (&Model{Model: longId, Engine: openai.EngineName}).GetModel()
		assert.Equal(t, clippedId, model)
		assert.Equal(t, clippedId, name)
		assert.Equal(t, "", version)
		assert.Len(t, clipWarnings(hook), 1)
	})
	t.Run("ServiceModel", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		logHook, hook := captureLogs(t)
		t.Setenv("VISION_TEST_LONG_MODEL", longId)
		model, _, _ := (&Model{Engine: openai.EngineName, Service: Service{Model: "${VISION_TEST_LONG_MODEL}"}}).GetModel()
		assert.Equal(t, clippedId, model)
		assert.Len(t, clipWarnings(hook), 1)
		assert.Empty(t, clipWarnings(logHook))
	})
	t.Run("Name", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		_, hook := captureLogs(t)
		model, _, _ := (&Model{Name: longId, Engine: openai.EngineName}).GetModel()
		assert.Equal(t, clippedId, model)
		assert.Len(t, clipWarnings(hook), 1)
	})
	t.Run("UnusedNameNotReported", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		_, hook := captureLogs(t)
		model, _, _ := (&Model{Name: longId, Model: "gpt-5-mini", Engine: openai.EngineName}).GetModel()
		assert.Equal(t, "gpt-5-mini", model)
		assert.Empty(t, clipWarnings(hook))
	})
	t.Run("UnusedModelNotReported", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		_, hook := captureLogs(t)
		m := &Model{Model: longId, Engine: openai.EngineName, Service: Service{Model: "gpt-5-mini"}}
		model, _, _ := m.GetModel()
		assert.Equal(t, "gpt-5-mini", model)
		assert.Empty(t, clipWarnings(hook))
	})
	t.Run("DisabledServiceNotReported", func(t *testing.T) {
		resetClippedModelIdWarnings(t)
		_, hook := captureLogs(t)
		assert.Equal(t, "", (&Service{Model: longId, Disabled: true}).GetModel())
		assert.Empty(t, clipWarnings(hook))
	})
}
