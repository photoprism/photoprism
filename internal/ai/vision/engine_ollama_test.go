package vision

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

func TestRegisterOllamaEngineDefaults(t *testing.T) {
	original := os.Getenv(ollama.APIKeyEnv)
	originalCaptionModel := CaptionModel.Clone()
	testCaptionModel := CaptionModel.Clone()
	testCaptionModel.Model = ""
	testCaptionModel.Service.Uri = ""
	testCaptionModel.Service.Think = ""
	cloudToken := "moo9yaiS4ShoKiojiathie2vuejiec2X.Mahl7ewaej4ebi7afq8f_vwe" //nolint:gosec

	t.Cleanup(func() {
		if original == "" {
			_ = os.Unsetenv(ollama.APIKeyEnv)
		} else {
			_ = os.Setenv(ollama.APIKeyEnv, original)
		}
		CaptionModel = originalCaptionModel
		registerOllamaEngineDefaults()
	})
	t.Run("SelfHosted", func(t *testing.T) {
		ensureEnvOnce = sync.Once{}
		CaptionModel = testCaptionModel.Clone()
		t.Setenv(ollama.APIKeyEnv, "")
		t.Setenv(ollama.BaseUrlEnv, ollama.DefaultBaseUrl)

		registerOllamaEngineDefaults()

		info, ok := EngineInfoFor(ollama.EngineName)
		if !ok {
			t.Fatalf("expected engine info for %s", ollama.EngineName)
		}

		if info.Uri != ollama.DefaultUri {
			t.Fatalf("expected default uri %s, got %s", ollama.DefaultUri, info.Uri)
		}

		if info.DefaultModel != ollama.DefaultModel {
			t.Fatalf("expected default model %s, got %s", ollama.DefaultModel, info.DefaultModel)
		}

		if CaptionModel.Model != ollama.DefaultModel {
			t.Fatalf("expected caption model %s, got %s", ollama.DefaultModel, CaptionModel.Model)
		}

		if CaptionModel.Service.Uri != ollama.DefaultUri {
			t.Fatalf("expected caption model uri %s, got %s", ollama.DefaultUri, CaptionModel.Service.Uri)
		}

		if CaptionModel.Service.Think != ollama.DefaultThink {
			t.Fatalf("expected caption model think %s, got %s", ollama.DefaultThink, CaptionModel.Service.Think)
		}
	})
	t.Run("Cloud", func(t *testing.T) {
		ensureEnvOnce = sync.Once{}
		CaptionModel = testCaptionModel.Clone()
		t.Setenv(ollama.BaseUrlEnv, ollama.CloudBaseUrl+"/")

		registerOllamaEngineDefaults()

		info, ok := EngineInfoFor(ollama.EngineName)
		if !ok {
			t.Fatalf("expected engine info for %s", ollama.EngineName)
		}

		if info.Uri != ollama.DefaultUri {
			t.Fatalf("expected default uri %s, got %s", ollama.DefaultUri, info.Uri)
		}

		if info.DefaultModel != ollama.CloudModel {
			t.Fatalf("expected cloud model %s, got %s", ollama.CloudModel, info.DefaultModel)
		}

		if CaptionModel.Model != ollama.CloudModel {
			t.Fatalf("expected caption model %s, got %s", ollama.CloudModel, CaptionModel.Model)
		}

		if CaptionModel.Service.Uri != ollama.DefaultUri {
			t.Fatalf("expected caption model uri %s, got %s", ollama.DefaultUri, CaptionModel.Service.Uri)
		}

		if CaptionModel.Service.Think != ollama.DefaultThink {
			t.Fatalf("expected caption model think %s, got %s", ollama.DefaultThink, CaptionModel.Service.Think)
		}
	})
	t.Run("ApiKeyAloneKeepsLocalDefaults", func(t *testing.T) {
		ensureEnvOnce = sync.Once{}
		CaptionModel = testCaptionModel.Clone()
		t.Setenv(ollama.APIKeyEnv, cloudToken)
		t.Setenv(ollama.BaseUrlEnv, ollama.DefaultBaseUrl)

		registerOllamaEngineDefaults()

		info, ok := EngineInfoFor(ollama.EngineName)
		if !ok {
			t.Fatalf("expected engine info for %s", ollama.EngineName)
		}

		if info.DefaultModel != ollama.DefaultModel {
			t.Fatalf("expected default model %s, got %s", ollama.DefaultModel, info.DefaultModel)
		}
	})
	t.Run("NewModels", func(t *testing.T) {
		ensureEnvOnce = sync.Once{}
		CaptionModel = testCaptionModel.Clone()

		t.Setenv(ollama.BaseUrlEnv, ollama.CloudBaseUrl)
		registerOllamaEngineDefaults()

		model := &Model{Type: ModelTypeCaption, Engine: ollama.EngineName}
		model.ApplyEngineDefaults()

		if model.Model != ollama.CloudModel {
			t.Fatalf("expected model %s, got %s", ollama.CloudModel, model.Model)
		}

		if model.Service.Uri != ollama.DefaultUri {
			t.Fatalf("expected service uri %s, got %s", ollama.DefaultUri, model.Service.Uri)
		}

		if model.Service.RequestFormat != ApiFormatOllama || model.Service.ResponseFormat != ApiFormatOllama {
			t.Fatalf("expected request/response format %s, got %s/%s", ApiFormatOllama, model.Service.RequestFormat, model.Service.ResponseFormat)
		}

		if model.Service.FileScheme != scheme.Base64 {
			t.Fatalf("expected file scheme %s, got %s", scheme.Base64, model.Service.FileScheme)
		}

		if model.Resolution != ollama.DefaultResolution {
			t.Fatalf("expected resolution %d, got %d", ollama.DefaultResolution, model.Resolution)
		}

		if model.Service.Think != ollama.DefaultThink {
			t.Fatalf("expected service think %s, got %s", ollama.DefaultThink, model.Service.Think)
		}
	})
}

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
	t.Run("CaptionFromThinkingField", func(t *testing.T) {
		req := &ApiRequest{}
		payload := ollama.Response{
			Response: "",
			Thinking: "A tabby cat with a white chest stares upward.",
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

		if resp.Result.Caption == nil {
			t.Fatal("expected caption result")
		}
		if resp.Result.Caption.Text != "A tabby cat with a white chest stares upward." {
			t.Fatalf("unexpected caption: %q", resp.Result.Caption.Text)
		}
	})
	t.Run("CaptionPrefersResponseOverThinking", func(t *testing.T) {
		req := &ApiRequest{}
		payload := ollama.Response{
			Response: "A tabby cat with a white chest stares upward.",
			Thinking: "Reasoning text that should not become the caption.",
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

		if resp.Result.Caption == nil {
			t.Fatal("expected caption result")
		}
		if resp.Result.Caption.Text != "A tabby cat with a white chest stares upward." {
			t.Fatalf("expected response field caption, got %q", resp.Result.Caption.Text)
		}
	})
	t.Run("StripsLeadingReasoningBlock", func(t *testing.T) {
		req := &ApiRequest{}
		payload := ollama.Response{
			Response: "<think>The user wants a concise caption.</think>A tabby cat stares upward.",
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

		if resp.Result.Caption == nil {
			t.Fatal("expected caption result")
		}
		if resp.Result.Caption.Text != "A tabby cat stares upward." {
			t.Fatalf("expected reasoning block to be stripped, got %q", resp.Result.Caption.Text)
		}
	})
}

func TestOllamaParserUnavailableStatus(t *testing.T) {
	t.Run("Gone", func(t *testing.T) {
		req := &ApiRequest{Model: ollama.CloudModel}
		payload := ollama.Response{}
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		parser := ollamaParser{}
		resp, err := parser.Parse(context.Background(), req, raw, http.StatusGone)
		if err != nil {
			t.Fatalf("parse should not error on upstream status, got %v", err)
		}

		if resp.Code != http.StatusGone {
			t.Fatalf("expected code %d, got %d", http.StatusGone, resp.Code)
		}
		if len(resp.Result.Labels) != 0 || resp.Result.Caption != nil {
			t.Fatalf("expected empty result, got %+v", resp.Result)
		}
	})
}

func TestStripReasoningBlock(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"LeadingBlock", "<think>reasoning here</think>Actual caption.", "Actual caption."},
		{"CaseInsensitive", "<THINK>reasoning</THINK>\n  Caption text.", "Caption text."},
		{"UntaggedUnchanged", "The user wants a concise description of the image.", "The user wants a concise description of the image."},
		{"UnterminatedUnchanged", "<think>reasoning without a close tag and no caption", "<think>reasoning without a close tag and no caption"},
		{"Empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripReasoningBlock(tc.in); got != tc.want {
				t.Fatalf("stripReasoningBlock(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestOllamaParserNormalizeModes(t *testing.T) {
	const payload = `{"labels":[{"name":"ferris wheel","confidence":0.92,"topicality":0.88}]}`

	parse := func(t *testing.T, mode NormalizeType, body []byte) []LabelResult {
		t.Helper()

		req := &ApiRequest{Model: "gemma4:latest", Normalize: mode}
		resp, err := ollamaParser{}.Parse(context.Background(), req, body, http.StatusOK)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		return resp.Result.Labels
	}

	response := func(t *testing.T, field, value string) []byte {
		t.Helper()

		body, err := json.Marshal(map[string]string{"model": "gemma4:latest", field: value})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		return body
	}

	cases := []struct {
		name string
		mode NormalizeType
		want string
	}{
		{name: "Default", mode: "", want: "Ferris"},
		{name: "SingleWord", mode: NormalizeWord, want: "Ferris"},
		{name: "Phrase", mode: NormalizePhrase, want: "Ferris Wheel"},
		{name: "False", mode: NormalizeFalse, want: "Ferris Wheel"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			labels := parse(t, tc.mode, response(t, "response", payload))

			if len(labels) != 1 || labels[0].Name != tc.want {
				t.Fatalf("expected a single %q label, got %+v", tc.want, labels)
			}
		})
	}
	t.Run("ThinkingFallbackPayload", func(t *testing.T) {
		labels := parse(t, NormalizePhrase, response(t, "thinking", payload))

		if len(labels) != 1 || labels[0].Name != "Ferris Wheel" {
			t.Fatalf("expected the mode to apply to the fallback payload, got %+v", labels)
		}
	})
}

func TestRegisterOllamaEngineDefaultsNormalize(t *testing.T) {
	t.Cleanup(func() {
		ensureEnvOnce = sync.Once{}
		registerOllamaEngineDefaults()
	})

	// A model that inherits the engine URI is classified by the endpoint it resolves to,
	// so the same configuration follows the base URL without an engine-wide default.
	cases := []struct {
		name    string
		baseUrl string
		want    NormalizeType
	}{
		{name: "SelfHosted", baseUrl: ollama.DefaultBaseUrl, want: NormalizeWord},
		{name: "Cloud", baseUrl: ollama.CloudBaseUrl, want: NormalizePhrase},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(ollama.BaseUrlEnv, tc.baseUrl)
			ensureEnvOnce = sync.Once{}
			registerOllamaEngineDefaults()

			model := &Model{Type: ModelTypeLabels, Engine: ollama.EngineName}
			model.ApplyEngineDefaults()

			if got := model.GetNormalize(); got != tc.want {
				t.Fatalf("expected %q for %s, got %q", tc.want, tc.baseUrl, got)
			}
		})
	}
}
