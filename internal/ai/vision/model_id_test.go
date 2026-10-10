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
		logHook, hook := captureLogs(t)
		model, name, version := (&Model{Model: longId, Engine: openai.EngineName}).GetModel()
		assert.Equal(t, clippedId, model)
		assert.Equal(t, clippedId, name)
		assert.Equal(t, "", version)
		assert.Len(t, clipWarnings(hook), 1)
		assert.Empty(t, clipWarnings(logHook))
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
		logHook, hook := captureLogs(t)
		model, _, _ := (&Model{Name: longId, Engine: openai.EngineName}).GetModel()
		assert.Equal(t, clippedId, model)
		assert.Len(t, clipWarnings(hook), 1)
		assert.Empty(t, clipWarnings(logHook))
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

// TestModel_GetModelLength checks that a model name is limited to clean.LengthType characters, while its tag is
// kept as written.
func TestModel_GetModelLength(t *testing.T) {
	resetClippedModelIdWarnings(t)
	_, _ = captureLogs(t)
	long := strings.Repeat("m", 70)

	// Names within the limit are sent as they are, with the default tag.
	hf := "hf.co/mradermacher/Qwen2.5-VL-7B-Instruct-abliterated-GGUF"
	id, _, _ := (&Model{Type: ModelTypeCaption, Name: hf, Engine: "ollama"}).GetModel()
	assert.Equal(t, hf+":latest", id)

	id, name, version := (&Model{Type: ModelTypeCaption, Name: long, Engine: "ollama"}).GetModel()
	assert.Equal(t, strings.Repeat("m", clean.LengthType), name)
	assert.Equal(t, "latest", version)
	assert.Equal(t, name+":latest", id)

	id, _, _ = (&Model{Type: ModelTypeCaption, Name: long, Engine: openai.EngineName}).GetModel()
	assert.Len(t, id, clean.LengthType)

	// Tags written in the name are kept, and only the name is shortened.
	for raw, tag := range map[string]string{
		long + ":7b": "7b",
		"hf.co/some-org/" + strings.Repeat("x", 43) + ":Q4_K_M-instr": "Q4_K_M-instr",
		strings.Repeat("a", 60) + ":Q4_K_M":                           "Q4_K_M",
	} {
		for _, model := range []*Model{
			{Type: ModelTypeCaption, Name: raw, Engine: "ollama"},
			{Type: ModelTypeCaption, Name: "x", Engine: "ollama", Service: Service{Model: raw}},
			{Type: ModelTypeCaption, Name: raw, Service: Service{Uri: "http://ollama:11434/api/generate", RequestFormat: ApiFormatOllama}},
		} {
			id, name, version := model.GetModel()
			assert.Equal(t, tag, version, raw)
			assert.Equal(t, name+":"+tag, id)
			assert.LessOrEqual(t, len(name), clean.LengthType, id)
		}
	}
}

// TestModel_GetModelDefaultEngine checks that a long tagged name sent in the Vision API format keeps its tag.
func TestModel_GetModelDefaultEngine(t *testing.T) {
	resetClippedModelIdWarnings(t)
	_, _ = captureLogs(t)

	id, name, version := (&Model{Type: ModelTypeCaption, Name: strings.Repeat("m", 70) + ":7b"}).GetModel()
	assert.Equal(t, strings.Repeat("m", clean.LengthType), name)
	assert.Equal(t, name, id)
	assert.Equal(t, "7b", version)
}

// TestModel_GetModelClipWarning checks that one long Ollama model name is reported once, naming the name sent.
func TestModel_GetModelClipWarning(t *testing.T) {
	resetClippedModelIdWarnings(t)
	_, systemHook := captureLogs(t)

	_, name, _ := (&Model{Type: ModelTypeCaption, Name: strings.Repeat("m", 70) + ":7b", Engine: "ollama"}).GetModel()

	warnings := clipWarnings(systemHook)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], name)
	assert.NotContains(t, warnings[0], ":7b")
}

// TestModelIdText checks that identifiers are sanitized without being shortened.
func TestModelIdText(t *testing.T) {
	assert.Equal(t, "gemma3:4b", modelIdText(" gemma3:4b "))
	assert.Len(t, modelIdText(strings.Repeat("m", 70)), 70)
	assert.Equal(t, "", modelIdText(""))
}
