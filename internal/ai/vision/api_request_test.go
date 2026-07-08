package vision

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestApiRequestWriteLogRedactsBase64(t *testing.T) {
	logger, ok := log.(*logrus.Logger)
	if !ok {
		t.Fatalf("unexpected logger type %T", log)
	}

	originalLevel := logger.GetLevel()
	originalOutput := logger.Out

	buffer := &bytes.Buffer{}
	logger.SetLevel(logrus.TraceLevel)
	logger.SetOutput(buffer)

	defer func() {
		logger.SetOutput(originalOutput)
		logger.SetLevel(originalLevel)
	}()

	req := &ApiRequest{
		Url: "data:image/jpeg;base64," + strings.Repeat("C", 40),
		Images: Files{
			"data:image/png;base64," + strings.Repeat("A", 40),
			strings.Repeat("B", 48),
			"https://example.test/image.jpg",
		},
	}

	req.WriteLog()

	output := buffer.String()

	if output == "" {
		t.Fatalf("expected trace log output")
	}

	if strings.Contains(output, strings.Repeat("A", 24)) {
		t.Errorf("log contains unredacted data URL image payload: %s", output)
	}

	if strings.Contains(output, strings.Repeat("B", 24)) {
		t.Errorf("log contains unredacted base64 image payload: %s", output)
	}

	if strings.Contains(output, strings.Repeat("C", 24)) {
		t.Errorf("log contains unredacted data URL in url field: %s", output)
	}

	imagePreview := "data:image/png;base64," + strings.Repeat("A", logDataPreviewLength) + logDataTruncatedSuffix
	if !strings.Contains(output, imagePreview) {
		t.Errorf("missing truncated image data preview, got: %s", output)
	}

	base64Preview := strings.Repeat("B", logDataPreviewLength) + logDataTruncatedSuffix
	if !strings.Contains(output, base64Preview) {
		t.Errorf("missing truncated base64 preview, got: %s", output)
	}

	urlPreview := "data:image/jpeg;base64," + strings.Repeat("C", logDataPreviewLength) + logDataTruncatedSuffix
	if !strings.Contains(output, urlPreview) {
		t.Errorf("missing truncated url preview, got: %s", output)
	}

	if !strings.Contains(output, "https://example.test/image.jpg") {
		t.Errorf("expected https url to remain unchanged: %s", output)
	}
}

func TestApiRequestJSONThinkOmitempty(t *testing.T) {
	t.Run("OmitWhenEmpty", func(t *testing.T) {
		req := &ApiRequest{
			Model:          "qwen3-vl:4b",
			ResponseFormat: ApiFormatOllama,
		}

		data, err := req.JSON()
		if err != nil {
			t.Fatalf("json marshal failed: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("json unmarshal failed: %v", err)
		}

		if _, ok := payload["think"]; ok {
			t.Fatalf("expected think field to be omitted, payload: %s", string(data))
		}
	})
	t.Run("IncludeWhenSet", func(t *testing.T) {
		req := &ApiRequest{
			Model:          "gpt-oss:20b",
			Think:          "low",
			ResponseFormat: ApiFormatOllama,
		}

		data, err := req.JSON()
		if err != nil {
			t.Fatalf("json marshal failed: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("json unmarshal failed: %v", err)
		}

		if got, ok := payload["think"].(string); !ok || got != "low" {
			t.Fatalf("expected think=low, got %#v", payload["think"])
		}
	})
	t.Run("StringFalseSerializedAsBool", func(t *testing.T) {
		req := &ApiRequest{
			Model:          "qwen3-vl:4b",
			Think:          "false",
			ResponseFormat: ApiFormatOllama,
		}

		data, err := req.JSON()
		if err != nil {
			t.Fatalf("json marshal failed: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("json unmarshal failed: %v", err)
		}

		if got, ok := payload["think"].(bool); !ok || got {
			t.Fatalf("expected think=false bool, got %#v", payload["think"])
		}
	})
	t.Run("StringTrueSerializedAsBool", func(t *testing.T) {
		req := &ApiRequest{
			Model:          "qwen3-vl:4b",
			Think:          "true",
			ResponseFormat: ApiFormatOllama,
		}

		data, err := req.JSON()
		if err != nil {
			t.Fatalf("json marshal failed: %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(data, &payload); err != nil {
			t.Fatalf("json unmarshal failed: %v", err)
		}

		if got, ok := payload["think"].(bool); !ok || !got {
			t.Fatalf("expected think=true bool, got %#v", payload["think"])
		}
	})
}
