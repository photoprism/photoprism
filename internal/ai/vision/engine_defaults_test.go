package vision

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/pkg/fs"
)

// reregisterEngineDefaults sets the variables and registers the engine defaults again, and once more
// when the test ends, after the variables have been restored.
func reregisterEngineDefaults(t *testing.T, env map[string]string) {
	t.Helper()

	original := CaptionModel.Clone()

	// Registered before t.Setenv, so it runs after the variables are restored.
	t.Cleanup(func() {
		ensureEnvOnce = sync.Once{}
		CaptionModel = original
		registerOllamaEngineDefaults()
		registerOpenAIEngineDefaults()
	})

	for name, value := range env {
		t.Setenv(name, value)
	}

	ensureEnvOnce = sync.Once{}
	CaptionModel = &Model{Type: ModelTypeCaption, Engine: ollama.EngineName, Run: original.Run}
	registerOllamaEngineDefaults()
	registerOpenAIEngineDefaults()
}

// TestRegisterOpenAIEngineDefaults checks the default service URI and model of the OpenAI engine.
func TestRegisterOpenAIEngineDefaults(t *testing.T) {
	endpoint := func(t *testing.T) string {
		t.Helper()
		m := &Model{Type: ModelTypeCaption, Engine: openai.EngineName}
		m.ApplyEngineDefaults()
		uri, _ := m.Service.Endpoint()
		return uri
	}

	t.Run("BaseUrlUnset", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			openai.BaseUrlEnv: "",
		})

		assert.Equal(t, "https://api.openai.com/v1/responses", endpoint(t))
	})
	t.Run("BaseUrlSet", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			openai.BaseUrlEnv: "https://llm.example.com/v1/",
		})

		assert.Equal(t, "https://llm.example.com/v1/responses", endpoint(t))
	})
	t.Run("BaseUrlWithoutVersion", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			openai.BaseUrlEnv: "https://llm.example.com",
		})

		assert.Equal(t, "https://llm.example.com/responses", endpoint(t))
	})
	t.Run("ServiceUriWithVariables", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			openai.BaseUrlEnv: "https://llm.example.com/v1",
			openai.ModelEnv:   "qwen3-vl-8b",
		})

		s := &Service{Uri: "${OPENAI_BASE_URL}/responses", Model: "${OPENAI_MODEL}"}
		uri, _ := s.Endpoint()

		assert.Equal(t, "https://llm.example.com/v1/responses", uri)
		assert.Equal(t, "qwen3-vl-8b", s.GetModel())
	})
	t.Run("ModelSet", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			openai.ModelEnv: " gpt-5-nano ",
		})

		info, ok := EngineInfoFor(openai.EngineName)
		require.True(t, ok)
		assert.Equal(t, "gpt-5-nano", info.DefaultModel)
		assert.Equal(t, "gpt-5-nano", openaiDefaultModel())

		m := &Model{Type: ModelTypeCaption, Engine: openai.EngineName}
		m.ApplyEngineDefaults()
		model, _, _ := m.GetModel()
		assert.Equal(t, "gpt-5-nano", model)
	})
	t.Run("ModelUnset", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			openai.ModelEnv: "",
		})

		assert.Equal(t, openai.DefaultModel, openaiDefaultModel())

		m := &Model{Type: ModelTypeCaption, Engine: openai.EngineName, Service: Service{Model: "${OPENAI_MODEL}"}}
		m.ApplyEngineDefaults()
		model, _, _ := m.GetModel()
		assert.Equal(t, openai.DefaultModel, model)
	})
	t.Run("ConfiguredModelWins", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			openai.ModelEnv: "gpt-5-nano",
		})

		for _, m := range []*Model{
			{Type: ModelTypeCaption, Engine: openai.EngineName, Name: "gpt-4.1-mini"},
			{Type: ModelTypeCaption, Engine: openai.EngineName, Model: "gpt-4.1-mini"},
			{Type: ModelTypeCaption, Engine: openai.EngineName, Service: Service{Model: "gpt-4.1-mini"}},
		} {
			m.ApplyEngineDefaults()
			model, _, _ := m.GetModel()
			assert.Equal(t, "gpt-4.1-mini", model)
		}
	})
	t.Run("SharedServiceNotUsed", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			openai.BaseUrlEnv: "",
		})
		useSharedService(t, "https://vision.example.com/api/v1/vision", "shared-vision-key")

		m := &Model{Type: ModelTypeCaption, Engine: openai.EngineName}
		m.ApplyEngineDefaults()
		uri, _ := m.Endpoint()

		assert.Equal(t, "https://api.openai.com/v1/responses", uri)
	})
}

// TestRegisterOllamaEngineDefaults_ModelEnv checks that OLLAMA_MODEL sets the default model of the Ollama engine.
func TestRegisterOllamaEngineDefaults_ModelEnv(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			ollama.BaseUrlEnv: ollama.DefaultBaseUrl,
			ollama.ModelEnv:   "qwen3-vl:8b",
		})

		info, ok := EngineInfoFor(ollama.EngineName)
		require.True(t, ok)
		assert.Equal(t, "qwen3-vl:8b", info.DefaultModel)
		assert.Equal(t, "qwen3-vl:8b", CaptionModel.Model)
	})
	t.Run("OverridesCloudPreset", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			ollama.BaseUrlEnv: ollama.CloudBaseUrl,
			ollama.ModelEnv:   "qwen3-vl:235b-cloud",
		})

		info, ok := EngineInfoFor(ollama.EngineName)
		require.True(t, ok)
		assert.Equal(t, "qwen3-vl:235b-cloud", info.DefaultModel)
	})
	t.Run("Unset", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			ollama.BaseUrlEnv: ollama.DefaultBaseUrl,
			ollama.ModelEnv:   "",
		})

		info, ok := EngineInfoFor(ollama.EngineName)
		require.True(t, ok)
		assert.Equal(t, ollama.DefaultModel, info.DefaultModel)

		m := &Model{Type: ModelTypeCaption, Engine: ollama.EngineName, Service: Service{Model: "${OLLAMA_MODEL}"}}
		m.ApplyEngineDefaults()
		model, _, _ := m.GetModel()
		assert.Equal(t, ollama.DefaultModel, model)
	})
	t.Run("StoragePathIgnored", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			ollama.BaseUrlEnv: ollama.DefaultBaseUrl,
			ollama.ModelEnv:   "",
			"OLLAMA_MODELS":   "/var/lib/ollama/models",
		})

		info, ok := EngineInfoFor(ollama.EngineName)
		require.True(t, ok)
		assert.Equal(t, ollama.DefaultModel, info.DefaultModel)
	})
	t.Run("ConfiguredModelWins", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			ollama.BaseUrlEnv: ollama.DefaultBaseUrl,
			ollama.ModelEnv:   "qwen3-vl:8b",
		})

		m := &Model{Type: ModelTypeLabels, Engine: ollama.EngineName, Name: "gemma3:27b"}
		m.ApplyEngineDefaults()
		model, _, _ := m.GetModel()
		assert.Equal(t, "gemma3:27b", model)
	})
}

// TestEnvModel checks that a model identifier from the environment is trimmed and sanitized.
func TestEnvModel(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		t.Setenv("VISION_TEST_MODEL", "  qwen3-vl:8b  ")
		assert.Equal(t, "qwen3-vl:8b", envModel("VISION_TEST_MODEL"))
	})
	t.Run("Blank", func(t *testing.T) {
		t.Setenv("VISION_TEST_MODEL", "   ")
		assert.Equal(t, "", envModel("VISION_TEST_MODEL"))
	})
	t.Run("Clipped", func(t *testing.T) {
		t.Setenv("VISION_TEST_MODEL", strings.Repeat("m", 100))
		assert.Len(t, envModel("VISION_TEST_MODEL"), 64)
	})
}

// TestEnvModelTagged checks that the identifier is sanitized without being shortened.
func TestEnvModelTagged(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		t.Setenv("VISION_TEST_MODEL", "  qwen3-vl:8b  ")
		assert.Equal(t, "qwen3-vl:8b", envModelTagged("VISION_TEST_MODEL"))
	})
	t.Run("Blank", func(t *testing.T) {
		t.Setenv("VISION_TEST_MODEL", "   ")
		assert.Equal(t, "", envModelTagged("VISION_TEST_MODEL"))
	})
	t.Run("Long", func(t *testing.T) {
		t.Setenv("VISION_TEST_MODEL", strings.Repeat("m", 100))
		assert.Len(t, envModelTagged("VISION_TEST_MODEL"), 100)
	})
}

// TestOllamaEnvModelTag checks that a long OLLAMA_MODEL keeps its tag when it is sent.
func TestOllamaEnvModelTag(t *testing.T) {
	resetClippedModelIdWarnings(t)
	_, _ = captureLogs(t)
	t.Setenv(ollama.ModelEnv, "hf.co/some-org/"+strings.Repeat("x", 43)+":Q4_K_M-instr")
	registerOllamaEngineDefaults()
	t.Cleanup(registerOllamaEngineDefaults)

	model := &Model{Type: ModelTypeCaption, Engine: ollama.EngineName}
	model.ApplyEngineDefaults()

	id, _, version := model.GetModel()
	assert.Equal(t, "Q4_K_M-instr", version)
	assert.Equal(t, "hf.co/some-org/"+strings.Repeat("x", 43)+":Q4_K_M-instr", id)
}

// TestOpenaiRequestModel checks the model named for a request.
func TestOpenaiRequestModel(t *testing.T) {
	t.Run("Configured", func(t *testing.T) {
		assert.Equal(t, "gpt-4.1-mini", openaiRequestModel(&ApiRequest{Model: " gpt-4.1-mini "}))
	})
	t.Run("Default", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{openai.ModelEnv: ""})
		assert.Equal(t, openai.DefaultModel, openaiRequestModel(&ApiRequest{}))
		assert.Equal(t, openai.DefaultModel, openaiRequestModel(nil))
	})
	t.Run("Environment", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{openai.ModelEnv: "gpt-5-nano"})
		assert.Equal(t, "gpt-5-nano", openaiRequestModel(&ApiRequest{}))
	})
}

// TestOpenaiDefaultModel checks that the default model of the OpenAI engine follows OPENAI_MODEL.
func TestOpenaiDefaultModel(t *testing.T) {
	t.Run("NotRegistered", func(t *testing.T) {
		info, ok := EngineInfoFor(openai.EngineName)
		require.True(t, ok)

		engineMu.Lock()
		delete(engineAliasIndex, openai.EngineName)
		engineMu.Unlock()

		t.Cleanup(func() { RegisterEngineAlias(openai.EngineName, info) })

		assert.Equal(t, openai.DefaultModel, openaiDefaultModel())
	})
	t.Run("Builtin", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			openai.ModelEnv: "",
		})
		assert.Equal(t, openai.DefaultModel, openaiDefaultModel())
	})
	t.Run("Environment", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{
			openai.ModelEnv: "gpt-5-nano",
		})
		assert.Equal(t, "gpt-5-nano", openaiDefaultModel())

		payload, err := (&ApiRequest{ResponseFormat: ApiFormatOpenAI, Images: []string{"data:image/jpeg;base64,AA=="}}).JSON()
		require.NoError(t, err)
		assert.Contains(t, string(payload), `"model":"gpt-5-nano"`)
	})
}

// TestLogOpenAIBaseUrl checks that a base URL other than the default is written to the system log only.
func TestLogOpenAIBaseUrl(t *testing.T) {
	t.Run("Custom", func(t *testing.T) {
		t.Setenv(openai.BaseUrlEnv, "https://user:secret@llm.example.com/v1")
		logHook, systemHook := captureLogs(t)

		logOpenAIBaseUrl()

		assert.Empty(t, logHook.AllEntries())
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Contains(t, systemHook.LastEntry().Message, "openai engine uses base url")
		assert.Contains(t, systemHook.LastEntry().Message, "llm.example.com/v1")
		assert.NotContains(t, systemHook.LastEntry().Message, "secret")
	})
	t.Run("Query", func(t *testing.T) {
		t.Setenv(openai.BaseUrlEnv, "https://llm.example.com/v1?token=s3cr3t-value&api-version=preview")
		logHook, systemHook := captureLogs(t)

		logOpenAIBaseUrl()

		assert.Empty(t, logHook.AllEntries())
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Contains(t, systemHook.LastEntry().Message, "llm.example.com/v1")
		assert.NotContains(t, systemHook.LastEntry().Message, "s3cr3t-value")
	})
	t.Run("Invalid", func(t *testing.T) {
		t.Setenv(openai.BaseUrlEnv, "https://llm.example.com/%zz")
		logHook, systemHook := captureLogs(t)

		logOpenAIBaseUrl()

		assert.Empty(t, logHook.AllEntries())
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, "vision: openai engine uses an invalid base url", systemHook.LastEntry().Message)
	})
	t.Run("Default", func(t *testing.T) {
		t.Setenv(openai.BaseUrlEnv, openai.DefaultBaseUrl)
		logHook, systemHook := captureLogs(t)

		logOpenAIBaseUrl()

		assert.Empty(t, logHook.AllEntries())
		assert.Empty(t, systemHook.AllEntries())
	})
	t.Run("ConfigLoad", func(t *testing.T) {
		reregisterEngineDefaults(t, map[string]string{openai.BaseUrlEnv: "https://llm.example.com/v1"})

		load := func(t *testing.T, yaml string) []string {
			t.Helper()
			configFile := filepath.Join(t.TempDir(), "vision.yml")
			require.NoError(t, os.WriteFile(configFile, []byte(yaml), fs.ModeConfigFile))
			_, systemHook := captureLogs(t)
			require.NoError(t, NewConfig().Load(configFile))

			var messages []string

			for _, entry := range systemHook.AllEntries() {
				if strings.Contains(entry.Message, "openai engine uses") {
					messages = append(messages, entry.Message)
				}
			}

			return messages
		}

		assert.Equal(t, []string{"vision: openai engine uses base url https://llm.example.com/v1"},
			load(t, "Models:\n- Type: caption\n  Engine: openai\n"))
		assert.Empty(t, load(t, "Models:\n- Type: caption\n  Engine: openai\n  Service:\n    Uri: https://llm.example.com/v1/responses\n"))
		assert.Empty(t, load(t, "Models:\n- Type: caption\n  Engine: ollama\n"))
	})
}

// TestModel_UsesOpenAIDefaultUri checks which models send their requests to the OpenAI base URL.
func TestModel_UsesOpenAIDefaultUri(t *testing.T) {
	enabled := &Model{Type: ModelTypeCaption, Engine: openai.EngineName}
	enabled.ApplyEngineDefaults()
	assert.True(t, enabled.usesOpenAIDefaultUri())

	disabled := enabled.Clone()
	disabled.Disabled = true
	assert.False(t, disabled.usesOpenAIDefaultUri())

	own := &Model{Type: ModelTypeCaption, Engine: openai.EngineName, Service: Service{Uri: "https://llm.example.com/v1/responses"}}
	own.ApplyEngineDefaults()
	assert.False(t, own.usesOpenAIDefaultUri())

	ollamaModel := &Model{Type: ModelTypeCaption, Engine: ollama.EngineName}
	ollamaModel.ApplyEngineDefaults()
	assert.False(t, ollamaModel.usesOpenAIDefaultUri())

	assert.False(t, (*Model)(nil).usesOpenAIDefaultUri())
}
