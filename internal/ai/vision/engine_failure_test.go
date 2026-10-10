package vision

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
)

// TestServiceFailureKey checks that failures are recorded per engine and model.
func TestServiceFailureKey(t *testing.T) {
	assert.Equal(t, "openai\x00gpt-5-mini", serviceFailureKey(openai.EngineName, "gpt-5-mini"))
	assert.NotEqual(t, serviceFailureKey(openai.EngineName, "gemma3"), serviceFailureKey(ollama.EngineName, "gemma3"))
}

// TestWarnServiceFailure checks that a failure is logged once per engine, model, and status.
func TestWarnServiceFailure(t *testing.T) {
	t.Run("RepeatedAtDebugLevel", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, systemHook := captureLogs(t)

		warnServiceFailure(openai.EngineName, ApiFormatOpenAI, "gpt-5-mini", http.StatusTooManyRequests)
		warnServiceFailure(openai.EngineName, ApiFormatOpenAI, "gpt-5-mini", http.StatusTooManyRequests)

		require.Len(t, logHook.AllEntries(), 2)
		assert.Equal(t, logrus.WarnLevel, logHook.AllEntries()[0].Level)
		assert.Equal(t, "vision: openai request for model gpt-5-mini failed (status 429)", logHook.AllEntries()[0].Message)
		assert.Equal(t, logrus.DebugLevel, logHook.AllEntries()[1].Level)
		assert.Equal(t, "vision: openai request for model gpt-5-mini failed again (status 429)", logHook.AllEntries()[1].Message)
		assert.Empty(t, systemHook.AllEntries())
	})
	t.Run("Unavailable", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		warnServiceFailure(openai.EngineName, ApiFormatOpenAI, "gpt-4-vision", http.StatusNotFound)

		require.NotNil(t, logHook.LastEntry())
		assert.Equal(t, "vision: openai model gpt-4-vision is unavailable (status 404), check the model name and the service uri", logHook.LastEntry().Message)

		warnServiceFailure(ollama.EngineName, ApiFormatOllama, "gemma3:27b", http.StatusNotFound)
		assert.Equal(t, "vision: ollama model gemma3:27b is unavailable (status 404), it may have been retired or renamed", logHook.LastEntry().Message)

		warnServiceFailure(openai.EngineName, ApiFormatOpenAI, "gpt-4-vision", http.StatusGone)
		assert.Equal(t, "vision: openai model gpt-4-vision is unavailable (status 410), it may have been retired or renamed", logHook.LastEntry().Message)
	})
	t.Run("PerEngine", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		warnServiceFailure(openai.EngineName, ApiFormatOpenAI, "gemma3", http.StatusInternalServerError)
		warnServiceFailure(ollama.EngineName, ApiFormatOllama, "gemma3", http.StatusInternalServerError)

		assert.Len(t, logHook.AllEntries(), 2)
		assert.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
	})
}

// TestClearServiceFailure checks that a cleared failure is logged again.
func TestClearServiceFailure(t *testing.T) {
	resetServiceFailures(t)
	logHook, _ := captureLogs(t)

	warnServiceFailure(openai.EngineName, ApiFormatOpenAI, "gpt-5-mini", http.StatusInternalServerError)
	clearServiceFailure(openai.EngineName, "gpt-5-mini")
	warnServiceFailure(openai.EngineName, ApiFormatOpenAI, "gpt-5-mini", http.StatusInternalServerError)

	require.Len(t, logHook.AllEntries(), 2)
	assert.Equal(t, logrus.WarnLevel, logHook.AllEntries()[1].Level)
}

// TestServiceStatusError checks the error and warning for each status class.
func TestServiceStatusError(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resetServiceFailures(t)
		warnServiceFailure(openai.EngineName, ApiFormatOpenAI, "gpt-5-mini", http.StatusInternalServerError)

		assert.NoError(t, serviceStatusError(openai.EngineName, ApiFormatOpenAI, "gpt-5-mini", http.StatusOK))

		_, loaded := serviceFailures.Load(serviceFailureKey(openai.EngineName, "gpt-5-mini"))
		assert.False(t, loaded)
	})
	t.Run("Redirect", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		warnServiceFailure(openai.EngineName, ApiFormatOpenAI, "gpt-5-mini", http.StatusInternalServerError)
		logHook.Reset()

		assert.EqualError(t, serviceStatusError(openai.EngineName, ApiFormatOpenAI, "gpt-5-mini", http.StatusFound), "openai service request failed (status 302)")
		assert.Empty(t, logHook.AllEntries())

		prev, loaded := serviceFailures.Load(serviceFailureKey(openai.EngineName, "gpt-5-mini"))
		assert.True(t, loaded, "a redirect must keep the failure state")
		assert.Equal(t, http.StatusInternalServerError, prev)
	})
	t.Run("Failure", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		assert.EqualError(t, serviceStatusError(openai.EngineName, ApiFormatOpenAI, "gpt-5-mini", http.StatusBadRequest), "openai service request failed (status 400)")
		require.Len(t, logHook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
	})
}

// TestOpenAIParserFailure checks that failed OpenAI requests warn once per model and status like Ollama.
func TestOpenAIParserFailure(t *testing.T) {
	failed := func(t *testing.T, model string, status int) {
		t.Helper()
		_, err := openaiParser{}.Parse(context.Background(), &ApiRequest{Model: model}, []byte("{}"), status)
		require.Error(t, err)
	}

	t.Run("WarnDebugRearmed", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		failed(t, "gpt-5-mini", http.StatusInternalServerError)
		failed(t, "gpt-5-mini", http.StatusInternalServerError)

		raw, err := json.Marshal(map[string]any{"output": []any{}})
		require.NoError(t, err)
		_, err = openaiParser{}.Parse(context.Background(), &ApiRequest{Model: "gpt-5-mini"}, raw, http.StatusOK)
		require.NoError(t, err)

		failed(t, "gpt-5-mini", http.StatusInternalServerError)

		var levels []logrus.Level

		for _, entry := range logHook.AllEntries() {
			if entry.Message == "vision: openai request for model gpt-5-mini failed (status 500)" ||
				entry.Message == "vision: openai request for model gpt-5-mini failed again (status 500)" {
				levels = append(levels, entry.Level)
			}
		}

		assert.Equal(t, []logrus.Level{logrus.WarnLevel, logrus.DebugLevel, logrus.WarnLevel}, levels)
	})
	t.Run("DefaultModelNamed", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		failed(t, "", http.StatusNotFound)

		require.NotNil(t, logHook.LastEntry())
		assert.Contains(t, logHook.LastEntry().Message, "openai model "+openaiDefaultModel()+" is unavailable")
	})
}

// TestParse_EngineName checks that a failed request is reported with the engine of the model, not the
// response format.
func TestParse_EngineName(t *testing.T) {
	t.Run("OllamaWithOpenAIFormat", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		_, err := openaiParser{}.Parse(context.Background(), &ApiRequest{Model: "qwen3-vl:8b", Engine: ollama.EngineName}, nil, http.StatusInternalServerError)
		require.Error(t, err)
		require.NotNil(t, logHook.LastEntry())
		assert.Equal(t, "vision: ollama request for model qwen3-vl:8b failed (status 500)", logHook.LastEntry().Message)
	})
	t.Run("OpenAIWithOllamaFormat", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		_, err := ollamaParser{}.Parse(context.Background(), &ApiRequest{Model: "gpt-5-mini", Engine: openai.EngineName}, nil, http.StatusInternalServerError)
		require.Error(t, err)
		require.NotNil(t, logHook.LastEntry())
		assert.Equal(t, "vision: openai request for model gpt-5-mini failed (status 500)", logHook.LastEntry().Message)
	})
	t.Run("Default", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		_, err := ollamaParser{}.Parse(context.Background(), &ApiRequest{Model: "gemma3:4b"}, nil, http.StatusInternalServerError)
		require.Error(t, err)
		require.NotNil(t, logHook.LastEntry())
		assert.Equal(t, "vision: ollama request for model gemma3:4b failed (status 500)", logHook.LastEntry().Message)
	})
}

// TestApiRequest_EngineName checks the engine name of a request and its default.
func TestApiRequest_EngineName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "ollama", (&ApiRequest{Engine: "ollama"}).engineName("openai"))
	})
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "openai", (&ApiRequest{}).engineName("openai"))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "openai", (*ApiRequest)(nil).engineName("openai"))
	})
}

// TestModel_ApplyServiceEngine checks that a request records the configured or inferred engine of the model.
func TestModel_ApplyServiceEngine(t *testing.T) {
	t.Run("Configured", func(t *testing.T) {
		model := &Model{Type: ModelTypeCaption, Name: "qwen3-vl:8b", Engine: ollama.EngineName, Service: Service{Uri: "http://ollama:11434/v1/responses", RequestFormat: ApiFormatOpenAI}}
		req := &ApiRequest{}
		model.ApplyService(req)
		assert.Equal(t, ollama.EngineName, req.Engine)
	})
	t.Run("Inferred", func(t *testing.T) {
		for format, engine := range map[ApiFormat]string{ApiFormatOllama: ollama.EngineName, ApiFormatOpenAI: openai.EngineName} {
			model := &Model{Type: ModelTypeCaption, Name: "qwen3-vl:8b", Service: Service{Uri: "http://llm.example.com/api", RequestFormat: format}}
			req := &ApiRequest{}
			model.ApplyService(req)
			assert.Equal(t, engine, req.Engine, format)
			assert.Equal(t, model.requestEngine(), req.Engine, format)
		}
	})
}

// TestWarnServiceFailure_FormatHint checks that the hint for a missing model follows the request format.
func TestWarnServiceFailure_FormatHint(t *testing.T) {
	resetServiceFailures(t)
	logHook, _ := captureLogs(t)

	warnServiceFailure(ollama.EngineName, ApiFormatOpenAI, "qwen3-vl:8b", http.StatusNotFound)
	assert.Equal(t, "vision: ollama model qwen3-vl:8b is unavailable (status 404), check the model name and the service uri", logHook.LastEntry().Message)

	warnServiceFailure(openai.EngineName, ApiFormatOllama, "gpt-5-mini", http.StatusNotFound)
	assert.Equal(t, "vision: openai model gpt-5-mini is unavailable (status 404), it may have been retired or renamed", logHook.LastEntry().Message)

	warnServiceFailure("bad\nengine", ApiFormatOllama, "gemma3", http.StatusInternalServerError)
	assert.NotContains(t, logHook.LastEntry().Message, "\n")
}

// TestApiRequest_EngineJson checks that the engine of a request is neither written to nor read from JSON.
func TestApiRequest_EngineJson(t *testing.T) {
	data, err := json.Marshal(&ApiRequest{Model: "gemma3:4b", Engine: ollama.EngineName})
	require.NoError(t, err)
	assert.NotContains(t, string(data), ollama.EngineName+"\"")

	var req ApiRequest
	require.NoError(t, json.Unmarshal([]byte(`{"model":"gemma3:4b","engine":"ollama","Engine":"ollama"}`), &req))
	assert.Equal(t, "", req.Engine)
}
