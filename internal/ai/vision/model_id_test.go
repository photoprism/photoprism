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

// resetModelIdWarnings clears the reported identifiers before and after a test.
func resetModelIdWarnings(t *testing.T) {
	t.Helper()
	modelIdWarned.Clear()
	t.Cleanup(modelIdWarned.Clear)
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
		resetModelIdWarnings(t)
		_, hook := captureLogs(t)
		assert.Equal(t, "unsloth/Qwen3.5-9B-GGUF:Q4_K_M", cleanModelId(" unsloth/Qwen3.5-9B-GGUF:Q4_K_M "))
		assert.Empty(t, clipWarnings(hook))
	})
	t.Run("Empty", func(t *testing.T) {
		resetModelIdWarnings(t)
		_, hook := captureLogs(t)
		assert.Equal(t, "", cleanModelId(""))
		assert.Equal(t, "", cleanModelId("   "))
		assert.Empty(t, clipWarnings(hook))
	})
	t.Run("MaxLength", func(t *testing.T) {
		resetModelIdWarnings(t)
		_, hook := captureLogs(t)
		assert.Equal(t, maxId, cleanModelId(maxId))
		assert.Equal(t, maxId, cleanModelId(maxId+"  "))
		assert.Equal(t, maxId, cleanModelId(maxId+"ä"))
		assert.Equal(t, maxId, cleanModelId(maxId+"\x00\x7f"))
		assert.Empty(t, clipWarnings(hook))
	})
	t.Run("InvalidCharsWarnOnce", func(t *testing.T) {
		resetModelIdWarnings(t)
		logHook, hook := captureLogs(t)
		assert.Equal(t, "gemma3:4b", cleanModelId("gemma3\x00:4b"))
		assert.Equal(t, "gemma3:4b", cleanModelId("gemma3:4b\x7f"))
		assert.Equal(t, "gemma3:4b", cleanModelId("gemma3:4b"))
		assert.Empty(t, clipWarnings(hook))

		var warnings []string

		for _, entry := range hook.AllEntries() {
			if entry.Level == logrus.WarnLevel {
				warnings = append(warnings, entry.Message)
			}
		}

		require.Len(t, warnings, 1)
		assert.Equal(t, "vision: model identifier contains unsupported characters, using gemma3:4b", warnings[0])
		assert.Empty(t, logHook.AllEntries())
	})
	t.Run("ClippedWarnsOnce", func(t *testing.T) {
		resetModelIdWarnings(t)
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
		resetModelIdWarnings(t)
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
		resetModelIdWarnings(t)
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
		resetModelIdWarnings(t)
		logHook, hook := captureLogs(t)
		model, name, version := (&Model{Model: longId, Engine: openai.EngineName}).GetModel()
		assert.Equal(t, clippedId, model)
		assert.Equal(t, clippedId, name)
		assert.Equal(t, "", version)
		assert.Len(t, clipWarnings(hook), 1)
		assert.Empty(t, clipWarnings(logHook))
	})
	t.Run("ServiceModel", func(t *testing.T) {
		resetModelIdWarnings(t)
		logHook, hook := captureLogs(t)
		t.Setenv("VISION_TEST_LONG_MODEL", longId)
		model, _, _ := (&Model{Engine: openai.EngineName, Service: Service{Model: "${VISION_TEST_LONG_MODEL}"}}).GetModel()
		assert.Equal(t, clippedId, model)
		assert.Len(t, clipWarnings(hook), 1)
		assert.Empty(t, clipWarnings(logHook))
	})
	t.Run("Name", func(t *testing.T) {
		resetModelIdWarnings(t)
		logHook, hook := captureLogs(t)
		model, _, _ := (&Model{Name: longId, Engine: openai.EngineName}).GetModel()
		assert.Equal(t, clippedId, model)
		assert.Len(t, clipWarnings(hook), 1)
		assert.Empty(t, clipWarnings(logHook))
	})
	t.Run("UnusedNameNotReported", func(t *testing.T) {
		resetModelIdWarnings(t)
		_, hook := captureLogs(t)
		model, _, _ := (&Model{Name: longId, Model: "gpt-5-mini", Engine: openai.EngineName}).GetModel()
		assert.Equal(t, "gpt-5-mini", model)
		assert.Empty(t, clipWarnings(hook))
	})
	t.Run("UnusedModelNotReported", func(t *testing.T) {
		resetModelIdWarnings(t)
		_, hook := captureLogs(t)
		m := &Model{Model: longId, Engine: openai.EngineName, Service: Service{Model: "gpt-5-mini"}}
		model, _, _ := m.GetModel()
		assert.Equal(t, "gpt-5-mini", model)
		assert.Empty(t, clipWarnings(hook))
	})
	t.Run("DisabledServiceNotReported", func(t *testing.T) {
		resetModelIdWarnings(t)
		_, hook := captureLogs(t)
		assert.Equal(t, "", (&Service{Model: longId, Disabled: true}).GetModel())
		assert.Empty(t, clipWarnings(hook))
	})
}

// TestModel_GetModelLength checks that a model name is limited to clean.LengthType characters, while its tag is
// kept as written.
func TestModel_GetModelLength(t *testing.T) {
	resetModelIdWarnings(t)
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
	resetModelIdWarnings(t)
	_, _ = captureLogs(t)

	id, name, version := (&Model{Type: ModelTypeCaption, Name: strings.Repeat("m", 70) + ":7b"}).GetModel()
	assert.Equal(t, strings.Repeat("m", clean.LengthType), name)
	assert.Equal(t, name, id)
	assert.Equal(t, "7b", version)
}

// TestModel_GetModelClipWarning checks that one long Ollama model name is reported once, naming the name sent.
func TestModel_GetModelClipWarning(t *testing.T) {
	resetModelIdWarnings(t)
	_, systemHook := captureLogs(t)

	_, name, _ := (&Model{Type: ModelTypeCaption, Name: strings.Repeat("m", 70) + ":7b", Engine: "ollama"}).GetModel()

	warnings := clipWarnings(systemHook)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], name)
	assert.NotContains(t, warnings[0], ":7b")
}

// TestHasModelId checks that a value counts as a model identifier only if something is left once sanitized.
func TestHasModelId(t *testing.T) {
	assert.True(t, hasModelId(" gemma3:4b "))
	assert.True(t, hasModelId(strings.Repeat("m", 70)))
	assert.False(t, hasModelId(""))
	assert.False(t, hasModelId(" \t\n"))
	assert.False(t, hasModelId("\x00\x7f"))
}

// TestModel_GetModelJunkParts checks that a name or tag holding nothing once sanitized is not split off or sent.
func TestModel_GetModelJunkParts(t *testing.T) {
	resetModelIdWarnings(t)
	_, _ = captureLogs(t)

	id, _, _ := (&Model{Type: ModelTypeCaption, Name: "x", Engine: "ollama", Service: Service{Model: "\x00:bar"}}).GetModel()
	assert.Equal(t, ":bar:latest", id)

	id, name, version := (&Model{Type: ModelTypeCaption, Name: "\x00", Engine: "ollama"}).GetModel()
	assert.Equal(t, "", id)
	assert.Equal(t, "", name)
	assert.Equal(t, "", version)

	m := &Model{Type: ModelTypeCaption, Model: "\x00", Engine: "ollama"}
	m.ApplyEngineDefaults()
	assert.NotEqual(t, "\x00", m.Model)
	assert.True(t, hasModelId(m.Model))
}

// TestModel_GetModelTagWarning checks that a long or invalid tag is reported, and that both issues of one
// identifier are reported.
func TestModel_GetModelTagWarning(t *testing.T) {
	resetModelIdWarnings(t)
	_, systemHook := captureLogs(t)

	_, _, version := (&Model{Type: ModelTypeCaption, Name: "gemma3:" + strings.Repeat("t", 70), Engine: "ollama"}).GetModel()
	assert.Len(t, version, clean.LengthType)
	require.Len(t, clipWarnings(systemHook), 1)
	assert.Contains(t, clipWarnings(systemHook)[0], version)

	cleanModelId("\x00" + strings.Repeat("m", 70))

	var chars []string

	for _, entry := range systemHook.AllEntries() {
		if strings.Contains(entry.Message, "unsupported characters") {
			chars = append(chars, entry.Message)
		}
	}

	assert.Len(t, clipWarnings(systemHook), 2)
	assert.Len(t, chars, 1)
}

// TestModel_GetModelControlChars checks that control characters are removed from the name and the tag, and
// that a source holding nothing else counts as unset.
func TestModel_GetModelControlChars(t *testing.T) {
	resetModelIdWarnings(t)
	_, _ = captureLogs(t)

	id, name, version := (&Model{Type: ModelTypeCaption, Name: "gemma3\x00:4b\x7f", Engine: "ollama"}).GetModel()
	assert.Equal(t, "gemma3:4b", id)
	assert.Equal(t, "gemma3", name)
	assert.Equal(t, "4b", version)

	id, _, _ = (&Model{Type: ModelTypeCaption, Name: "gemma3:4b", Model: "\x00", Engine: "ollama"}).GetModel()
	assert.Equal(t, "gemma3:4b", id)

	id, _, _ = (&Model{Type: ModelTypeCaption, Name: "gemma3:4b", Engine: "ollama", Service: Service{Model: "\x00\t"}}).GetModel()
	assert.Equal(t, "gemma3:4b", id)
	assert.Equal(t, "", (&Service{Model: "\x00"}).GetModel())
	assert.Equal(t, "llava:7b", (&Service{Model: " llava:7b "}).GetModel())
}
