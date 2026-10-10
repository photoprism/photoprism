package vision

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/pkg/media"
)

func TestGenerateCaption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	} else if _, err := net.DialTimeout("tcp", "photoprism-vision:5000", 10*time.Second); err != nil {
		t.Skip("skipping test because photoprism-vision is not running.")
	}

	t.Run("Success", func(t *testing.T) {
		expectedText := "An image of sound waves"

		result, model, err := GenerateCaption(Files{"https://dl.photoprism.app/img/artwork/colorwaves-400.jpg"}, media.SrcRemote)

		assert.NoError(t, err)
		assert.NotNil(t, model)
		assert.IsType(t, CaptionResult{}, result)
		assert.LessOrEqual(t, float32(0.0), result.Confidence)

		t.Logf("caption: %#v", result)

		assert.Equal(t, expectedText, result.Text)
	})
	t.Run("Invalid", func(t *testing.T) {
		result, model, err := GenerateCaption(nil, media.SrcLocal)

		assert.Error(t, err)
		assert.Nil(t, model)
		assert.IsType(t, CaptionResult{}, result)
		assert.Equal(t, "", result.Text)
		assert.Equal(t, float32(0.0), result.Confidence)
	})
}

// TestGenerateCaptionServiceError checks that the error text of a failed service request is only written to the system log.
func TestGenerateCaptionServiceError(t *testing.T) {
	const marker = "remote-error-marker"

	resetServiceFailures(t)

	prevConfig := Config
	t.Cleanup(func() { Config = prevConfig })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"` + marker + `"}}`))
	}))
	defer server.Close()

	model := &Model{Type: ModelTypeCaption, Name: "gpt-5-mini", Engine: openai.EngineName, Service: Service{Uri: server.URL, Key: "test-key"}}
	model.ApplyEngineDefaults()
	Config = &ConfigValues{Models: Models{model}, Thresholds: DefaultThresholds}

	logHook, systemHook := captureLogs(t)

	result, _, err := GenerateCaption(Files{samplesPath + "/cat_224.jpeg"}, media.SrcLocal)
	assert.Nil(t, result)
	assert.EqualError(t, err, "openai service request failed (status 400)")

	for _, entry := range logHook.AllEntries() {
		assert.NotContains(t, entry.Message, marker, entry.Level.String())
	}

	require.Len(t, systemHook.AllEntries(), 1)
	assert.Equal(t, logrus.ErrorLevel, systemHook.LastEntry().Level)
	assert.Contains(t, systemHook.LastEntry().Message, marker)
}
