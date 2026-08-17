package vision

import (
	"sync"
	"testing"

	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

// useSelfHostedOllamaDefaults pins the Ollama engine alias to its self-hosted defaults, which
// otherwise follow OLLAMA_BASE_URL in the ambient environment.
func useSelfHostedOllamaDefaults(t *testing.T) {
	// Cleanup runs last-registered-first, so this is registered before t.Setenv to rebuild
	// the alias after the environment has been restored.
	t.Cleanup(func() {
		ensureEnvOnce = sync.Once{}
		registerOllamaEngineDefaults()
	})

	t.Setenv(ollama.BaseUrlEnv, ollama.DefaultBaseUrl)
	ensureEnvOnce = sync.Once{}
	registerOllamaEngineDefaults()
}

func TestParseNormalizeType(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  NormalizeType
	}{
		{name: "EmptyIsAuto", in: "", out: NormalizeAuto},
		{name: "DefaultAlias", in: "default", out: NormalizeAuto},
		{name: "SingleWord", in: "single-word", out: NormalizeWord},
		{name: "WordAlias", in: "word", out: NormalizeWord},
		{name: "Phrase", in: "phrase", out: NormalizePhrase},
		{name: "MultiWordAlias", in: "multi-word", out: NormalizePhrase},
		{name: "False", in: "false", out: NormalizeFalse},
		{name: "OffAlias", in: "off", out: NormalizeFalse},
		{name: "NoneAlias", in: "none", out: NormalizeFalse},
		{name: "WhitespaceTrim", in: "  phrase  ", out: NormalizePhrase},
		{name: "Uppercase", in: "SINGLE-WORD", out: NormalizeWord},
		{name: "UnknownFallsBack", in: "something", out: NormalizeAuto},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseNormalizeType(tc.in); got != tc.out {
				t.Fatalf("ParseNormalizeType(%q) = %q, want %q", tc.in, got, tc.out)
			}
		})
	}
}

func TestIsNormalizeType(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  bool
	}{
		{name: "Empty", in: "", out: true},
		{name: "Known", in: "phrase", out: true},
		{name: "Alias", in: "off", out: true},
		{name: "Uppercase", in: "False", out: true},
		{name: "Unknown", in: "phrasy", out: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsNormalizeType(tc.in); got != tc.out {
				t.Fatalf("IsNormalizeType(%q) = %v, want %v", tc.in, got, tc.out)
			}
		})
	}
}

func TestReportNormalizeType(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{name: "EmptyIsAuto", in: "", out: "auto"},
		{name: "UnknownIsAuto", in: "bogus", out: "auto"},
		{name: "Phrase", in: "phrase", out: "phrase"},
		{name: "AliasIsCanonical", in: "off", out: "false"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ReportNormalizeType(tc.in); got != tc.out {
				t.Fatalf("ReportNormalizeType(%q) = %q, want %q", tc.in, got, tc.out)
			}
		})
	}
}

func TestModel_GetNormalize(t *testing.T) {
	useSelfHostedOllamaDefaults(t)

	// A local classifier vocabulary is multi-word by construction, so an engine may
	// declare a default that differs from the one the LLM engines register.
	RegisterEngineAlias("test-phrase-engine", EngineInfo{
		RequestFormat:    ApiFormatVision,
		ResponseFormat:   ApiFormatVision,
		FileScheme:       scheme.Data,
		DefaultNormalize: NormalizePhrase,
	})

	cases := []struct {
		name  string
		model *Model
		want  NormalizeType
	}{
		{name: "Nil", model: nil, want: NormalizeWord},
		{name: "Unset", model: &Model{}, want: NormalizeWord},
		{name: "Explicit", model: &Model{Normalize: "phrase"}, want: NormalizePhrase},
		{name: "Alias", model: &Model{Normalize: "off"}, want: NormalizeFalse},
		{name: "Invalid", model: &Model{Normalize: "bogus"}, want: NormalizeWord},
		{name: "OllamaDefault", model: &Model{Engine: "ollama"}, want: NormalizeWord},
		{name: "OpenAIHosted", model: &Model{Engine: "openai", Name: "gpt-5-mini"}, want: NormalizePhrase},
		{name: "OpenAICompatibleLocal", model: &Model{Engine: "openai", Name: "Qwen2.5-VL-7B-Instruct"}, want: NormalizeWord},
		{name: "EngineDefault", model: &Model{Engine: "test-phrase-engine"}, want: NormalizePhrase},
		{name: "ExplicitBeatsEngine", model: &Model{Engine: "test-phrase-engine", Normalize: "single-word"}, want: NormalizeWord},
		{name: "UnknownEngine", model: &Model{Engine: "nope"}, want: NormalizeWord},
		{name: "CloudTag", model: &Model{Engine: "ollama", Model: "minimax-m3:cloud"}, want: NormalizePhrase},
		{name: "CloudTagViaName", model: &Model{Engine: "ollama", Name: "kimi-k3", Version: "cloud"}, want: NormalizePhrase},
		{name: "CloudTagExplicitOverride", model: &Model{Engine: "ollama", Model: "minimax-m3:cloud", Normalize: "single-word"}, want: NormalizeWord},
		{name: "SelfHostedTag", model: &Model{Engine: "ollama", Model: "gemma4:latest"}, want: NormalizeWord},
		{name: "CloudEndpoint", model: &Model{Engine: "ollama", Model: "qwen3-vl:235b-instruct",
			Service: Service{Uri: "https://ollama.com/api/generate", Method: "POST"}}, want: NormalizePhrase},
		{name: "LocalEndpointStaysWord", model: &Model{Engine: "ollama", Model: "gemma3:latest",
			Service: Service{Uri: "http://192.0.2.10:11434/api/generate", Method: "POST"}}, want: NormalizeWord},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.model.GetNormalize(); got != tc.want {
				t.Fatalf("(*Model).GetNormalize() = %q, want %q", got, tc.want)
			}
		})
	}
}
