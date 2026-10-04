package vision

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
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
		resetOllamaFailures(t)
		logHook, _ := captureLogs(t)
		raw, err := json.Marshal(ollama.Response{})
		require.NoError(t, err)

		resp, err := ollamaParser{}.Parse(context.Background(), &ApiRequest{Model: ollama.CloudModel}, raw, http.StatusGone)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "ollama service request failed (status 410)")
		require.NotNil(t, logHook.LastEntry())
		assert.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
		assert.Contains(t, logHook.LastEntry().Message, "is unavailable (status 410)")
	})
	t.Run("ServerErrorCaption", func(t *testing.T) {
		resetOllamaFailures(t)
		logHook, _ := captureLogs(t)
		raw, err := json.Marshal(ollama.Response{Model: "qwen2.5vl:latest", Response: "A caption from a failed request."})
		require.NoError(t, err)

		resp, err := ollamaParser{}.Parse(context.Background(), &ApiRequest{Model: "qwen2.5vl:latest"}, raw, http.StatusInternalServerError)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "ollama service request failed (status 500)")
		require.NotNil(t, logHook.LastEntry())
		assert.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
		assert.Equal(t, "vision: ollama request for model qwen2.5vl:latest failed (status 500)", logHook.LastEntry().Message)
	})
	t.Run("Redirect", func(t *testing.T) {
		logHook, _ := captureLogs(t)
		raw, err := json.Marshal(ollama.Response{Model: "qwen2.5vl:latest", Response: "A caption."})
		require.NoError(t, err)

		resp, err := ollamaParser{}.Parse(context.Background(), &ApiRequest{Model: "qwen2.5vl:latest"}, raw, http.StatusMultipleChoices)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "ollama service request failed (status 300)")
		assert.Empty(t, logHook.AllEntries())
	})
}

// resetOllamaFailures clears the logged request failures before and after a test.
func resetOllamaFailures(t *testing.T) {
	t.Helper()
	ollamaFailures.Clear()
	t.Cleanup(ollamaFailures.Clear)
}

// ollamaFailureWarnings returns the logged warnings about failed Ollama requests.
func ollamaFailureWarnings(hook *logtest.Hook) []string {
	var result []string

	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.WarnLevel && strings.HasPrefix(entry.Message, "vision: ollama ") {
			result = append(result, entry.Message)
		}
	}

	return result
}

func TestWarnOllamaFailure(t *testing.T) {
	failed := func(t *testing.T, model string, status int) {
		t.Helper()
		_, err := ollamaParser{}.Parse(context.Background(), &ApiRequest{Model: model}, []byte("{}"), status)
		require.Error(t, err)
	}

	t.Run("RepeatedAtDebugLevel", func(t *testing.T) {
		resetOllamaFailures(t)
		logHook, _ := captureLogs(t)
		failed(t, "gemma3:27b", http.StatusNotFound)
		failed(t, "gemma3:27b", http.StatusNotFound)
		failed(t, "gemma3:27b", http.StatusNotFound)
		assert.Equal(t, []string{"vision: ollama model gemma3:27b is unavailable (status 404), it may have been retired or renamed"}, ollamaFailureWarnings(logHook))
		require.NotNil(t, logHook.LastEntry())
		assert.Equal(t, logrus.DebugLevel, logHook.LastEntry().Level)
		assert.Equal(t, "vision: ollama request for model gemma3:27b failed again (status 404)", logHook.LastEntry().Message)
	})
	t.Run("StatusChange", func(t *testing.T) {
		resetOllamaFailures(t)
		logHook, _ := captureLogs(t)
		failed(t, "gemma3:27b", http.StatusServiceUnavailable)
		failed(t, "gemma3:27b", http.StatusNotFound)
		failed(t, "gemma3:27b", http.StatusNotFound)
		assert.Len(t, ollamaFailureWarnings(logHook), 2)
	})
	t.Run("PerModel", func(t *testing.T) {
		resetOllamaFailures(t)
		logHook, _ := captureLogs(t)
		failed(t, "gemma3:27b", http.StatusInternalServerError)
		failed(t, "qwen3-vl:8b", http.StatusInternalServerError)
		assert.Len(t, ollamaFailureWarnings(logHook), 2)
	})
	t.Run("ClearedBySuccess", func(t *testing.T) {
		resetOllamaFailures(t)
		logHook, _ := captureLogs(t)
		failed(t, "gemma3:27b", http.StatusInternalServerError)
		raw, err := json.Marshal(ollama.Response{Model: "gemma3:27b", Response: "A caption."})
		require.NoError(t, err)
		_, err = ollamaParser{}.Parse(context.Background(), &ApiRequest{Model: "gemma3:27b"}, raw, http.StatusOK)
		require.NoError(t, err)
		failed(t, "gemma3:27b", http.StatusInternalServerError)
		assert.Len(t, ollamaFailureWarnings(logHook), 2)
	})
	t.Run("ClearedByInvalidBody", func(t *testing.T) {
		resetOllamaFailures(t)
		logHook, _ := captureLogs(t)
		failed(t, "gemma3:27b", http.StatusInternalServerError)
		_, err := ollamaParser{}.Parse(context.Background(), &ApiRequest{Model: "gemma3:27b"}, []byte("not json"), http.StatusOK)
		require.Error(t, err)
		failed(t, "gemma3:27b", http.StatusInternalServerError)
		assert.Len(t, ollamaFailureWarnings(logHook), 2)
	})
	t.Run("RedirectKeepsState", func(t *testing.T) {
		resetOllamaFailures(t)
		logHook, _ := captureLogs(t)
		failed(t, "gemma3:27b", http.StatusInternalServerError)
		failed(t, "gemma3:27b", http.StatusMultipleChoices)
		failed(t, "gemma3:27b", http.StatusInternalServerError)
		assert.Len(t, ollamaFailureWarnings(logHook), 1)
	})
	t.Run("RedirectNotLogged", func(t *testing.T) {
		resetOllamaFailures(t)
		logHook, _ := captureLogs(t)
		failed(t, "gemma3:27b", http.StatusMultipleChoices)
		assert.Empty(t, logHook.AllEntries())
		_, loaded := ollamaFailures.Load("gemma3:27b")
		assert.False(t, loaded)
	})
}

// TestOllamaParserInvalidLabels checks that invalid label JSON from the model is only quoted at debug level.
func TestOllamaParserInvalidLabels(t *testing.T) {
	resetOllamaInvalidLabels(t)
	logHook, systemHook := captureLogs(t)

	digits := strings.Repeat("9", 60)
	raw, err := json.Marshal(ollama.Response{Model: "qwen2.5vl:latest", Response: `{"labels":[{"name":"cat","priority":` + digits + `}]}`})
	require.NoError(t, err)

	resp, err := ollamaParser{}.Parse(context.Background(), &ApiRequest{Model: "qwen2.5vl:latest", Format: FormatJSON}, raw, http.StatusOK)
	require.NoError(t, err)
	assert.Empty(t, resp.Result.Labels)

	var warn, debug int

	for _, entry := range logHook.AllEntries() {
		switch entry.Level {
		case logrus.WarnLevel:
			warn++
			assert.Equal(t, "vision: ollama returned invalid labels for model qwen2.5vl:latest", entry.Message)
		case logrus.DebugLevel:
			if strings.Contains(entry.Message, "(parse ollama labels)") {
				debug++
				assert.Contains(t, entry.Message, digits)
			}
		default:
			assert.NotContains(t, entry.Message, digits, entry.Level.String())
		}
	}

	assert.Equal(t, 1, warn)
	assert.Equal(t, 1, debug)
	assert.Empty(t, systemHook.AllEntries())
}

// resetOllamaInvalidLabels clears the logged invalid-label models before and after a test.
func resetOllamaInvalidLabels(t *testing.T) {
	t.Helper()
	ollamaInvalidLabels.Clear()
	t.Cleanup(ollamaInvalidLabels.Clear)
}

func TestOllamaParserInvalidLabelsOnce(t *testing.T) {
	parse := func(t *testing.T, model, text string) {
		t.Helper()
		raw, err := json.Marshal(ollama.Response{Model: model, Response: text})
		require.NoError(t, err)
		_, err = ollamaParser{}.Parse(context.Background(), &ApiRequest{Model: model, Format: FormatJSON}, raw, http.StatusOK)
		require.NoError(t, err)
	}
	invalid := `{"labels":[{"name":"cat","priority":"high"}]}`
	valid := `{"labels":[{"name":"cat","confidence":0.9,"topicality":0.9}]}`

	t.Run("Repeated", func(t *testing.T) {
		resetOllamaInvalidLabels(t)
		logHook, _ := captureLogs(t)
		parse(t, "gemma3:27b", invalid)
		parse(t, "gemma3:27b", invalid)
		parse(t, "qwen3-vl:8b", invalid)
		assert.Len(t, ollamaFailureWarnings(logHook), 2)
	})
	t.Run("ClearedByValidLabels", func(t *testing.T) {
		resetOllamaInvalidLabels(t)
		logHook, _ := captureLogs(t)
		parse(t, "gemma3:27b", invalid)
		parse(t, "gemma3:27b", valid)
		parse(t, "gemma3:27b", invalid)
		assert.Len(t, ollamaFailureWarnings(logHook), 2)
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
	// so the same configuration follows the base URL without an engine-wide default. The name
	// carries no cloud tag, which is what makes the endpoint the deciding factor.
	cases := []struct {
		name    string
		baseUrl string
		want    NormalizeType
	}{
		{name: "SelfHostedNameSelfHostedEndpoint", baseUrl: ollama.DefaultBaseUrl, want: NormalizeWord},
		{name: "SelfHostedNameCloudEndpoint", baseUrl: ollama.CloudBaseUrl, want: NormalizePhrase},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(ollama.BaseUrlEnv, tc.baseUrl)
			ensureEnvOnce = sync.Once{}
			registerOllamaEngineDefaults()

			model := &Model{Type: ModelTypeLabels, Engine: ollama.EngineName, Model: "gemma4:latest"}
			model.ApplyEngineDefaults()

			if got := model.GetNormalize(); got != tc.want {
				t.Fatalf("expected %q for %s, got %q", tc.want, tc.baseUrl, got)
			}
		})
	}
}

func TestOllamaBuilderBuildModelId(t *testing.T) {
	images := Files{fs.Abs("./testdata/face_160x160.jpg")}

	for _, tc := range []struct {
		name string
		want string
	}{
		{"gemma3:27b", "gemma3:27b"},
		{"gemma3", "gemma3:latest"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			model := &Model{Type: ModelTypeLabels, Name: tc.name, Service: Service{RequestFormat: ApiFormatOllama}}
			req, err := ollamaBuilder{}.Build(context.Background(), model, images, media.SrcLocal)
			require.NoError(t, err)
			assert.Equal(t, tc.want, req.Model)
			assert.Equal(t, "", req.Version)
		})
	}
}
