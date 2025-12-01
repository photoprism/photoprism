package vision

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

func TestModelGetOptionsDefaultsOllamaLabels(t *testing.T) {
	ollamaModel := "redule26/huihui_ai_qwen2.5-vl-7b-abliterated:latest"

	model := &Model{
		Type:   ModelTypeLabels,
		Name:   ollamaModel,
		Engine: ollama.EngineName,
	}

	model.ApplyEngineDefaults()

	m, n, v := model.GetModel()

	assert.Equal(t, ollamaModel, m)
	assert.Equal(t, "redule26/huihui_ai_qwen2.5-vl-7b-abliterated", n)
	assert.Equal(t, "latest", v)

	opts := model.GetOptions()
	if opts == nil {
		t.Fatalf("expected options for labels model")
	}

	if opts.Temperature != DefaultTemperature {
		t.Errorf("unexpected temperature: got %v want %v", opts.Temperature, DefaultTemperature)
	}

	if opts.TopP != 0.9 {
		t.Errorf("unexpected top_p: got %v want 0.9", opts.TopP)
	}

	if len(opts.Stop) != 1 || opts.Stop[0] != "\n\n" {
		t.Fatalf("expected default stop sequence, got %#v", opts.Stop)
	}

	if opts != model.GetOptions() {
		t.Errorf("expected cached options pointer")
	}
}

func TestModel_GetModel(t *testing.T) {
	tests := []struct {
		name        string
		model       *Model
		wantModel   string
		wantName    string
		wantVersion string
	}{
		{
			name:        "Nil",
			wantModel:   "",
			wantName:    "",
			wantVersion: "",
		},
		{
			name: "OpenAINameOnly",
			model: &Model{
				Name:   "gpt-5-mini",
				Engine: openai.EngineName,
			},
			wantModel:   "gpt-5-mini",
			wantName:    "gpt-5-mini",
			wantVersion: "",
		},
		{
			name: "NonOpenAIAddsLatest",
			model: &Model{
				Name:   "gemma3",
				Engine: ollama.EngineName,
			},
			wantModel:   "gemma3:latest",
			wantName:    "gemma3",
			wantVersion: "latest",
		},
		{
			name: "ExplicitVersion",
			model: &Model{
				Name:    "gemma3",
				Version: "2",
				Engine:  ollama.EngineName,
			},
			wantModel:   "gemma3:2",
			wantName:    "gemma3",
			wantVersion: "2",
		},
		{
			name: "NameContainsVersion",
			model: &Model{
				Name:   "qwen2.5vl:7b",
				Engine: ollama.EngineName,
			},
			wantModel:   "qwen2.5vl:7b",
			wantName:    "qwen2.5vl",
			wantVersion: "7b",
		},
		{
			name: "ModelFieldFallback",
			model: &Model{
				Model:  "CUSTOM-MODEL",
				Engine: ollama.EngineName,
			},
			wantModel:   "custom-model:latest",
			wantName:    "custom-model",
			wantVersion: "latest",
		},
		{
			name: "ServiceOverrideWithVersion",
			model: &Model{
				Name:    "ignored",
				Engine:  ollama.EngineName,
				Service: Service{Model: "mixtral:8x7b"},
			},
			wantModel:   "mixtral:8x7b",
			wantName:    "mixtral",
			wantVersion: "8x7b",
		},
		{
			name: "ServiceOverrideOpenAI",
			model: &Model{
				Name:    "gpt-4.1",
				Engine:  openai.EngineName,
				Service: Service{Model: "gpt-5-mini"},
			},
			wantModel:   "gpt-5-mini",
			wantName:    "gpt-5-mini",
			wantVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, name, version := tt.model.GetModel()

			assert.Equal(t, tt.wantModel, model)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantVersion, version)
		})
	}
}

func TestModelGetOptionsRespectsCustomValues(t *testing.T) {
	model := &Model{
		Type:   ModelTypeLabels,
		Engine: ollama.EngineName,
		Options: &ApiRequestOptions{
			Temperature: 5,
			TopP:        0.95,
			Stop:        []string{"CUSTOM"},
		},
	}

	model.ApplyEngineDefaults()

	opts := model.GetOptions()
	if opts.Temperature != MaxTemperature {
		t.Errorf("temperature clamp failed: got %v want %v", opts.Temperature, MaxTemperature)
	}
	if opts.TopP != 0.95 {
		t.Errorf("top_p override lost: got %v", opts.TopP)
	}
	if len(opts.Stop) != 1 || opts.Stop[0] != "CUSTOM" {
		t.Errorf("stop override lost: %#v", opts.Stop)
	}
}

func TestModelGetOptionsFillsMissingFields(t *testing.T) {
	model := &Model{
		Type:    ModelTypeLabels,
		Engine:  ollama.EngineName,
		Options: &ApiRequestOptions{},
	}

	model.ApplyEngineDefaults()

	opts := model.GetOptions()
	if opts.TopP != 0.9 {
		t.Errorf("expected default top_p, got %v", opts.TopP)
	}
	if len(opts.Stop) != 1 || opts.Stop[0] != "\n\n" {
		t.Errorf("expected default stop sequence, got %#v", opts.Stop)
	}
}

func TestModelApplyEngineDefaultsSetsResolution(t *testing.T) {
	model := &Model{Type: ModelTypeLabels, Engine: ollama.EngineName}

	model.ApplyEngineDefaults()

	if model.Resolution != ollama.DefaultResolution {
		t.Fatalf("expected resolution %d, got %d", ollama.DefaultResolution, model.Resolution)
	}

	model.Resolution = 1024
	model.ApplyEngineDefaults()
	if model.Resolution != 1024 {
		t.Fatalf("expected custom resolution to be preserved, got %d", model.Resolution)
	}
}

func TestModelApplyEngineDefaultsSetsServiceDefaults(t *testing.T) {
	t.Run("OpenAIEngine", func(t *testing.T) {
		model := &Model{
			Type:   ModelTypeCaption,
			Engine: openai.EngineName,
		}

		model.ApplyEngineDefaults()

		assert.Equal(t, "https://api.openai.com/v1/responses", model.Service.Uri)
		assert.Equal(t, ApiFormatOpenAI, model.Service.RequestFormat)
		assert.Equal(t, ApiFormatOpenAI, model.Service.ResponseFormat)
		assert.Equal(t, scheme.Data, model.Service.FileScheme)
	})
	t.Run("PreserveExistingService", func(t *testing.T) {
		model := &Model{
			Type:   ModelTypeCaption,
			Engine: openai.EngineName,
			Service: Service{
				Uri:           "https://custom.example",
				FileScheme:    scheme.Base64,
				RequestFormat: ApiFormatOpenAI,
			},
		}

		model.ApplyEngineDefaults()

		assert.Equal(t, "https://custom.example", model.Service.Uri)
		assert.Equal(t, scheme.Base64, model.Service.FileScheme)
	})
}

func TestModelEndpointKeyOpenAIFallbacks(t *testing.T) {
	t.Run("EnvFile", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "openai.key")
		if err := os.WriteFile(path, []byte("from-file\n"), 0o600); err != nil {
			t.Fatalf("write key file: %v", err)
		}

		// Reset ensureEnvOnce.
		ensureEnvOnce = sync.Once{}

		t.Setenv("OPENAI_API_KEY", "")
		t.Setenv("OPENAI_API_KEY_FILE", path)

		model := &Model{Type: ModelTypeCaption, Engine: openai.EngineName}
		model.ApplyEngineDefaults()

		if got := model.EndpointKey(); got != "from-file" {
			t.Fatalf("expected file key, got %q", got)
		}
	})
	t.Run("CustomPlaceholder", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "env-secret")

		model := &Model{Type: ModelTypeCaption, Engine: openai.EngineName}
		model.ApplyEngineDefaults()
		if got := model.EndpointKey(); got != "env-secret" {
			t.Fatalf("expected env secret, got %q", got)
		}

		model.Service.Key = "${CUSTOM_KEY}"
		t.Setenv("CUSTOM_KEY", "custom-secret")
		if got := model.EndpointKey(); got != "custom-secret" {
			t.Fatalf("expected custom secret, got %q", got)
		}
	})
	t.Run("GlobalFallback", func(t *testing.T) {
		prev := ServiceKey
		ServiceKey = "${GLOBAL_KEY}"
		defer func() { ServiceKey = prev }()

		t.Setenv("GLOBAL_KEY", "global-secret")

		model := &Model{}
		if got := model.EndpointKey(); got != "global-secret" {
			t.Fatalf("expected global secret, got %q", got)
		}
	})
}

func TestModelGetSource(t *testing.T) {
	t.Run("NilModel", func(t *testing.T) {
		var model *Model
		if src := model.GetSource(); src != entity.SrcAuto {
			t.Fatalf("expected SrcAuto for nil model, got %s", src)
		}
	})
	t.Run("EngineAlias", func(t *testing.T) {
		model := &Model{Engine: ollama.EngineName}
		if src := model.GetSource(); src != entity.SrcOllama {
			t.Fatalf("expected SrcOllama, got %s", src)
		}
	})
	t.Run("RequestFormat", func(t *testing.T) {
		model := &Model{Service: Service{RequestFormat: ApiFormatOpenAI}}
		if src := model.GetSource(); src != entity.SrcOpenAI {
			t.Fatalf("expected SrcOpenAI, got %s", src)
		}
	})
	t.Run("DefaultImage", func(t *testing.T) {
		model := &Model{}
		if src := model.GetSource(); src != entity.SrcImage {
			t.Fatalf("expected SrcImage fallback, got %s", src)
		}
	})
}

func TestModelApplyService(t *testing.T) {
	t.Run("OpenAIHeaders", func(t *testing.T) {
		req := &ApiRequest{}
		model := &Model{
			Engine:  openai.EngineName,
			Service: Service{Org: "org-123", Project: "proj-abc"},
		}

		model.ApplyService(req)

		assert.Equal(t, "org-123", req.Org)
		assert.Equal(t, "proj-abc", req.Project)
	})
	t.Run("OtherEngineNoop", func(t *testing.T) {
		req := &ApiRequest{Org: "keep", Project: "keep"}
		model := &Model{Engine: ollama.EngineName, Service: Service{Org: "new", Project: "new"}}

		model.ApplyService(req)

		assert.Equal(t, "keep", req.Org)
		assert.Equal(t, "keep", req.Project)
	})
}

func TestModel_IsDefault(t *testing.T) {
	nasnetCopy := *NasnetModel //nolint:govet // copy for test inspection only
	nasnetCopy.Default = false

	cases := []struct {
		name  string
		model *Model
		want  bool
	}{
		{
			name:  "DefaultFlag",
			model: &Model{Default: true},
			want:  true,
		},
		{
			name:  "NasnetCopy",
			model: &nasnetCopy,
			want:  true,
		},
		{
			name: "CustomTensorFlow",
			model: &Model{
				Type:       ModelTypeLabels,
				Name:       "custom",
				TensorFlow: &tensorflow.ModelInfo{},
			},
			want: false,
		},
		{
			name: "RemoteService",
			model: &Model{
				Type:   ModelTypeCaption,
				Name:   "custom-caption",
				Engine: ollama.EngineName,
			},
			want: false,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if got := tc.model.IsDefault(); got != tc.want {
				t.Fatalf("IsDefault() = %v, want %v", got, tc.want)
			}
		})
	}
}
