package openai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadTestResponse(t *testing.T, name string) *Response {
	t.Helper()

	filePath := filepath.Join("testdata", name)

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", filePath, err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal %s: %v", filePath, err)
	}

	return &resp
}

func TestParseErrorMessage(t *testing.T) {
	t.Run("returns message when present", func(t *testing.T) {
		raw := []byte(`{"error":{"message":"Invalid schema"}}`)
		msg := ParseErrorMessage(raw)
		if msg != "Invalid schema" {
			t.Fatalf("expected message, got %q", msg)
		}
	})

	t.Run("returns empty string when error is missing", func(t *testing.T) {
		raw := []byte(`{"output":[]}`)
		if msg := ParseErrorMessage(raw); msg != "" {
			t.Fatalf("expected empty message, got %q", msg)
		}
	})
}

func TestResponseFirstTextCaption(t *testing.T) {
	resp := loadTestResponse(t, "caption-response.json")

	if jsonPayload := resp.FirstJSON(); len(jsonPayload) != 0 {
		t.Fatalf("expected no JSON payload, got: %s", jsonPayload)
	}

	text := resp.FirstText()
	expected := "A bee gathers nectar from the vibrant red poppy’s center."
	if text != expected {
		t.Fatalf("unexpected caption text: %q", text)
	}
}

func TestResponseFirstTextLabels(t *testing.T) {
	resp := loadTestResponse(t, "labels-response.json")

	if jsonPayload := resp.FirstJSON(); len(jsonPayload) != 0 {
		t.Fatalf("expected no JSON payload, got: %s", jsonPayload)
	}

	text := resp.FirstText()
	if len(text) == 0 {
		t.Fatal("expected structured JSON string in text payload")
	}
	if text[0] != '{' {
		t.Fatalf("expected JSON object in text payload, got %q", text)
	}
}

func TestResponseFirstJSONFromStructuredPayload(t *testing.T) {
	resp := &Response{
		ID:    "resp_structured",
		Model: "gpt-5-mini",
		Output: []ResponseOutput{
			{
				Role: "assistant",
				Content: []ResponseContent{
					{
						Type: "output_json",
						JSON: json.RawMessage(`{"labels":[{"name":"sunset"}]}`),
					},
				},
			},
		},
	}

	jsonPayload := resp.FirstJSON()
	if len(jsonPayload) == 0 {
		t.Fatal("expected JSON payload, got empty result")
	}

	var decoded struct {
		Labels []map[string]string `json:"labels"`
	}
	if err := json.Unmarshal(jsonPayload, &decoded); err != nil {
		t.Fatalf("failed to decode JSON payload: %v", err)
	}

	if len(decoded.Labels) != 1 || decoded.Labels[0]["name"] != "sunset" {
		t.Fatalf("unexpected JSON payload: %+v", decoded.Labels)
	}
}

func TestSchemaLabelsReturnsValidJSON(t *testing.T) {
	raw := SchemaLabels(false)

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("schema should be valid JSON: %v", err)
	}

	if decoded["type"] != "object" {
		t.Fatalf("expected type object, got %v", decoded["type"])
	}
}
