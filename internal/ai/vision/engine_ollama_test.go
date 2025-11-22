package vision

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
)

func TestOllamaDefaultConfidenceApplied(t *testing.T) {
	req := &ApiRequest{Format: FormatJSON}
	payload := ollama.Response{
		Result: ollama.ResultPayload{
			Labels: []ollama.LabelPayload{{Name: "forest path", Confidence: 0, Topicality: 0}},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	parser := ollamaParser{}
	resp, err := parser.Parse(context.Background(), req, raw, 200)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(resp.Result.Labels) != 1 {
		t.Fatalf("expected one label, got %d", len(resp.Result.Labels))
	}

	if resp.Result.Labels[0].Confidence != ollama.LabelConfidenceDefault {
		t.Fatalf("expected default confidence %.2f, got %.2f", ollama.LabelConfidenceDefault, resp.Result.Labels[0].Confidence)
	}
	if resp.Result.Labels[0].Topicality != ollama.LabelConfidenceDefault {
		t.Fatalf("expected topicality to default to confidence, got %.2f", resp.Result.Labels[0].Topicality)
	}
}

func TestOllamaParserFallbacks(t *testing.T) {
	t.Run("ThinkingFieldJSON", func(t *testing.T) {
		req := &ApiRequest{Format: FormatJSON}
		payload := ollama.Response{
			Thinking: `{"labels":[{"name":"cat","confidence":0.9,"topicality":0.8}]}`,
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		parser := ollamaParser{}
		resp, err := parser.Parse(context.Background(), req, raw, 200)
		if err != nil {
			t.Fatalf("parse failed: %v", err)
		}

		if len(resp.Result.Labels) != 1 || resp.Result.Labels[0].Name != "Cat" {
			t.Fatalf("expected cat label, got %+v", resp.Result.Labels)
		}
	})
	t.Run("JsonPrefixedResponse", func(t *testing.T) {
		req := &ApiRequest{} // no explicit format
		payload := ollama.Response{
			Response: `{"labels":[{"name":"cat","confidence":0.91,"topicality":0.81}]}`,
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		parser := ollamaParser{}
		resp, err := parser.Parse(context.Background(), req, raw, 200)
		if err != nil {
			t.Fatalf("parse failed: %v", err)
		}

		if len(resp.Result.Labels) != 1 || resp.Result.Labels[0].Name != "Cat" {
			t.Fatalf("expected cat label, got %+v", resp.Result.Labels)
		}
	})
}
