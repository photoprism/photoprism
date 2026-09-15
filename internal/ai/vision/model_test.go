package vision

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

func TestReadSchemaFile(t *testing.T) {
	t.Run("ReadsRegularFile", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "schema.json")
		if err := os.WriteFile(path, []byte(`{"type":"object"}`), 0o600); err != nil {
			t.Fatalf("write schema file: %v", err)
		}

		got, err := readSchemaFile(path)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != `{"type":"object"}` {
			t.Fatalf("unexpected schema content: %s", got)
		}
	})
	t.Run("RejectsDirectory", func(t *testing.T) {
		if _, err := readSchemaFile(t.TempDir()); err == nil {
			t.Fatal("expected error for directory path")
		}
	})
}

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
			wantModel:   "CUSTOM-MODEL:latest",
			wantName:    "CUSTOM-MODEL",
			wantVersion: "latest",
		},
		{
			name: "OpenAIPreservesHuggingFaceCase",
			model: &Model{
				Engine:  openai.EngineName,
				Service: Service{Model: "QuantTrio/Qwen3-VL-30B-A3B-Instruct-AWQ"},
			},
			wantModel:   "QuantTrio/Qwen3-VL-30B-A3B-Instruct-AWQ",
			wantName:    "QuantTrio/Qwen3-VL-30B-A3B-Instruct-AWQ",
			wantVersion: "",
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
		Options: &ModelOptions{
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
		Options: &ModelOptions{},
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
		assert.Equal(t, openai.APIKeyPlaceholder, model.Service.Key)
	})
	t.Run("OllamaEngineDefaults", func(t *testing.T) {
		model := &Model{
			Type:   ModelTypeLabels,
			Engine: ollama.EngineName,
		}

		model.ApplyEngineDefaults()

		assert.Equal(t, ApiFormatOllama, model.Service.RequestFormat)
		assert.Equal(t, ApiFormatOllama, model.Service.ResponseFormat)
		assert.Equal(t, scheme.Base64, model.Service.FileScheme)
		assert.Equal(t, ollama.APIKeyPlaceholder, model.Service.Key)
		assert.Equal(t, ollama.DefaultThink, model.Service.Think)
	})
	t.Run("OllamaPreservesExplicitThink", func(t *testing.T) {
		model := &Model{
			Type:    ModelTypeLabels,
			Engine:  ollama.EngineName,
			Service: Service{Think: "true"},
		}

		model.ApplyEngineDefaults()

		assert.Equal(t, "true", model.Service.Think)
	})
	t.Run("PreserveExistingService", func(t *testing.T) {
		model := &Model{
			Type:   ModelTypeCaption,
			Engine: openai.EngineName,
			Service: Service{
				Uri:           "https://custom.example",
				FileScheme:    scheme.Base64,
				RequestFormat: ApiFormatOpenAI,
				Key:           "custom-key",
			},
		}

		model.ApplyEngineDefaults()

		assert.Equal(t, "https://custom.example", model.Service.Uri)
		assert.Equal(t, scheme.Base64, model.Service.FileScheme)
		assert.Equal(t, "custom-key", model.Service.Key)
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

func TestModelEndpointKeyOllamaFallbacks(t *testing.T) {
	t.Run("EnvFile", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "ollama.key")
		if err := os.WriteFile(path, []byte("ollama-from-file\n"), 0o600); err != nil {
			t.Fatalf("write key file: %v", err)
		}

		ensureEnvOnce = sync.Once{}

		t.Setenv("OLLAMA_API_KEY", "")
		t.Setenv("OLLAMA_API_KEY_FILE", path)

		model := &Model{Type: ModelTypeCaption, Engine: ollama.EngineName}
		model.ApplyEngineDefaults()

		if got := model.EndpointKey(); got != "ollama-from-file" {
			t.Fatalf("expected file key, got %q", got)
		}
	})
	t.Run("EnvVariable", func(t *testing.T) {
		t.Setenv("OLLAMA_API_KEY", "ollama-env")

		model := &Model{Type: ModelTypeCaption, Engine: ollama.EngineName}
		model.ApplyEngineDefaults()

		if got := model.EndpointKey(); got != "ollama-env" {
			t.Fatalf("expected env key, got %q", got)
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
			Service: Service{Org: "org-123", Project: "proj-abc", Tier: "flex", Think: "medium"},
		}

		model.ApplyService(req)

		assert.Equal(t, "org-123", req.Org)
		assert.Equal(t, "proj-abc", req.Project)
		assert.Equal(t, "flex", req.Tier)
		assert.Equal(t, "medium", req.Think)
	})
	t.Run("OtherEngineIgnoresOpenAIHeadersButAppliesThink", func(t *testing.T) {
		req := &ApiRequest{Org: "keep", Project: "keep", Tier: "keep"}
		model := &Model{Engine: ollama.EngineName, Service: Service{Org: "new", Project: "new", Tier: "new", Think: "false"}}

		model.ApplyService(req)

		assert.Equal(t, "keep", req.Org)
		assert.Equal(t, "keep", req.Project)
		assert.Equal(t, "keep", req.Tier)
		assert.Equal(t, "false", req.Think)
	})
}

func TestModel_IsDefault(t *testing.T) {
	nasnetCopy := NasnetModel.Clone() //nolint:govet // copy for test inspection only
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
			model: nasnetCopy,
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

		t.Run(tc.name, func(t *testing.T) {
			if got := tc.model.IsDefault(); got != tc.want {
				t.Fatalf("IsDefault() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestModel_EngineNameONNX verifies local ONNX models report their runtime.
func TestModel_EngineNameONNX(t *testing.T) {
	model := &Model{Type: ModelTypeLabels, ONNX: &onnx.ModelInfo{}}
	assert.Equal(t, EngineONNX, model.EngineName())
}

// TestModel_ClassifyModelMissingRegisteredDisables verifies named models never fall back.
func TestModel_ClassifyModelMissingRegisteredDisables(t *testing.T) {
	previousModelsPath := ModelsPath
	ModelsPath = t.TempDir()
	t.Cleanup(func() { ModelsPath = previousModelsPath })

	model := NewLabelModel(classify.ModelRepViTM10)
	require.NotNil(t, model)
	assert.Nil(t, model.ClassifyModel())
	assert.True(t, model.Disabled)
}

func TestModel_FaceModel(t *testing.T) {
	restore := face.ConfiguredModel()

	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})

	t.Run("EmbeddingsDisabled", func(t *testing.T) {
		// FACE_MODEL=none must win over the model configured in vision.yml, otherwise
		// the TensorFlow fallback keeps generating embeddings that were turned off.
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelNone}))
		assert.Nil(t, (&Model{Name: "facenet", Type: ModelTypeFace}).FaceModel())
	})
	t.Run("EmbeddingsBlocked", func(t *testing.T) {
		// A library the configured model cannot read is migrated rather than added to, so
		// nothing generates embeddings until it is.
		t.Cleanup(face.UnblockEmbeddings)
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
			Name:  face.ModelFaceNet,
			Model: face.FindEmbeddingModel(face.ModelFaceNet),
		}))
		face.BlockEmbeddings("12 marker(s) use sface, but this instance is configured for facenet")

		assert.Nil(t, (&Model{Name: "facenet", Type: ModelTypeFace}).FaceModel())
	})
	t.Run("ActiveEmbedder", func(t *testing.T) {
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
			Name:  face.ModelFaceNet,
			Model: face.FindEmbeddingModel(face.ModelFaceNet),
		}))

		embedder := &stubEmbedder{dims: 128}
		prev := face.UseEmbedder(embedder)

		t.Cleanup(func() { face.UseEmbedder(prev) })

		assert.Equal(t, embedder, (&Model{Name: "facenet", Type: ModelTypeFace}).FaceModel())
	})
	t.Run("CustomModelDeprecated", func(t *testing.T) {
		// FACE_MODEL decides which model produces embeddings, so a custom face entry has
		// to say it is on the way out rather than look like a supported way to configure
		// one. Selecting it is what the operator would otherwise never be told about.
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
			Name:  face.ModelFaceNet,
			Model: face.FindEmbeddingModel(face.ModelFaceNet),
		}))

		prev := face.UseEmbedder(nil)
		t.Cleanup(func() { face.UseEmbedder(prev) })

		logger, ok := log.(*logrus.Logger)
		require.True(t, ok)

		originalOutput := logger.Out
		buffer := &bytes.Buffer{}
		logger.SetOutput(buffer)
		t.Cleanup(func() { logger.SetOutput(originalOutput) })

		(&Model{Name: "custom-face-net", Type: ModelTypeFace}).FaceModel()

		assert.Contains(t, buffer.String(), "custom-face-net")
		assert.Contains(t, buffer.String(), "deprecated")
		assert.Contains(t, buffer.String(), "FACE_MODEL")
	})
	t.Run("NilModel", func(t *testing.T) {
		assert.Nil(t, (*Model)(nil).FaceModel())
	})
}

func TestModel_IsCloud(t *testing.T) {
	cases := []struct {
		name  string
		model *Model
		want  bool
	}{
		{name: "Nil", model: nil, want: false},
		{name: "Empty", model: &Model{}, want: false},
		{name: "CloudTag", model: &Model{Engine: "ollama", Model: "minimax-m3:cloud"}, want: true},
		{name: "CloudVersion", model: &Model{Engine: "ollama", Name: "kimi-k3", Version: "cloud"}, want: true},
		{name: "SelfHosted", model: &Model{Engine: "ollama", Model: "gemma4:latest"}, want: false},
		{name: "NoVersion", model: &Model{Engine: "ollama", Name: "gemma4"}, want: false},
		{name: "OpenAIGPT", model: &Model{Engine: "openai", Name: "gpt-5-mini"}, want: true},
		{name: "OpenAIReasoning", model: &Model{Engine: "openai", Name: "o4-mini"}, want: true},
		{name: "OpenAICompatibleLocal", model: &Model{Engine: "openai", Name: "Qwen2.5-VL-7B-Instruct"}, want: false},
		{name: "OllamaGPTName", model: &Model{Engine: "ollama", Model: "gpt-oss:20b"}, want: false},
		{name: "CloudEndpointWithoutTag", model: &Model{Engine: "ollama", Model: "qwen3-vl:235b-instruct",
			Service: Service{Uri: "https://ollama.com/api/generate", Method: "POST"}}, want: true},
		{name: "LocalEndpoint", model: &Model{Engine: "ollama", Model: "gemma4:latest",
			Service: Service{Uri: "http://192.0.2.10:11434/api/generate", Method: "POST"}}, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.model.IsCloud())
		})
	}
}

func TestModel_MigrationFaceModel(t *testing.T) {
	t.Run("IgnoresTheBlock", func(t *testing.T) {
		// A migration writes every vector in its own target's space, so the gate against
		// mixing spaces would only stop the work that resolves the mismatch.
		t.Cleanup(face.UnblockEmbeddings)

		embedder := &stubEmbedder{dims: 128}
		prev := face.UseEmbedder(embedder)
		t.Cleanup(func() { face.UseEmbedder(prev) })

		face.BlockEmbeddings("12 marker(s) use sface, but this instance is configured for facenet")

		m := &Model{Name: "facenet", Type: ModelTypeFace}

		assert.Nil(t, m.FaceModel())
		assert.Equal(t, embedder, m.MigrationFaceModel())
	})
	t.Run("NilModel", func(t *testing.T) {
		assert.Nil(t, (*Model)(nil).MigrationFaceModel())
	})
}
