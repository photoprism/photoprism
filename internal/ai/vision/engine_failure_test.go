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

		warnServiceFailure(openai.EngineName, "gpt-5-mini", http.StatusTooManyRequests)
		warnServiceFailure(openai.EngineName, "gpt-5-mini", http.StatusTooManyRequests)

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

		warnServiceFailure(openai.EngineName, "gpt-4-vision", http.StatusNotFound)

		require.NotNil(t, logHook.LastEntry())
		assert.Equal(t, "vision: openai model gpt-4-vision is unavailable (status 404), check the model name and the service uri", logHook.LastEntry().Message)

		warnServiceFailure(ollama.EngineName, "gemma3:27b", http.StatusNotFound)
		assert.Equal(t, "vision: ollama model gemma3:27b is unavailable (status 404), it may have been retired or renamed", logHook.LastEntry().Message)

		warnServiceFailure(openai.EngineName, "gpt-4-vision", http.StatusGone)
		assert.Equal(t, "vision: openai model gpt-4-vision is unavailable (status 410), it may have been retired or renamed", logHook.LastEntry().Message)
	})
	t.Run("PerEngine", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		warnServiceFailure(openai.EngineName, "gemma3", http.StatusInternalServerError)
		warnServiceFailure(ollama.EngineName, "gemma3", http.StatusInternalServerError)

		assert.Len(t, logHook.AllEntries(), 2)
		assert.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
	})
}

// TestClearServiceFailure checks that a cleared failure is logged again.
func TestClearServiceFailure(t *testing.T) {
	resetServiceFailures(t)
	logHook, _ := captureLogs(t)

	warnServiceFailure(openai.EngineName, "gpt-5-mini", http.StatusInternalServerError)
	clearServiceFailure(openai.EngineName, "gpt-5-mini")
	warnServiceFailure(openai.EngineName, "gpt-5-mini", http.StatusInternalServerError)

	require.Len(t, logHook.AllEntries(), 2)
	assert.Equal(t, logrus.WarnLevel, logHook.AllEntries()[1].Level)
}

// TestServiceStatusError checks the error and warning for each status class.
func TestServiceStatusError(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resetServiceFailures(t)
		warnServiceFailure(openai.EngineName, "gpt-5-mini", http.StatusInternalServerError)

		assert.NoError(t, serviceStatusError(openai.EngineName, ApiFormatOpenAI, "gpt-5-mini", http.StatusOK))

		_, loaded := serviceFailures.Load(serviceFailureKey(openai.EngineName, "gpt-5-mini"))
		assert.False(t, loaded)
	})
	t.Run("Redirect", func(t *testing.T) {
		resetServiceFailures(t)
		logHook, _ := captureLogs(t)

		warnServiceFailure(openai.EngineName, "gpt-5-mini", http.StatusInternalServerError)
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
