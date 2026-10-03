package vision

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/ai/vision/openai"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

// TestModelCloneExportedFields verifies every exported configuration field is copied.
func TestModelCloneExportedFields(t *testing.T) {
	unsafeIndex, neutralIndex := 1, 2
	mean, stdDev, logits := float32(0.5), float32(0.25), true
	source := &Model{
		Type: ModelTypeNsfw, Default: true, Model: "model", Name: "name", Version: "version",
		Engine: EngineONNX, Run: RunAlways, System: "system", Prompt: "prompt", Format: "json",
		Normalize: NormalizePhrase, Schema: "schema", SchemaFile: "schema.json", Resolution: 224,
		TensorFlow: &tensorflow.ModelInfo{
			TFVersion: "2", Tags: []string{"serve"},
			Input: &tensorflow.PhotoInput{
				Name: "input", Intervals: []tensorflow.Interval{{Start: 0, End: 1, Mean: &mean, StdDev: &stdDev}},
				Height: 224, Width: 224, Shape: tensorflow.DefaultPhotoInputShape(),
			},
			Output: &tensorflow.ModelOutput{Name: "output", NumOutputs: 10, OutputsLogits: true},
		},
		ONNX: &onnx.ModelInfo{
			File: "model.onnx", Input: &onnx.Input{Name: "input", Width: 224, Height: 224},
			Output: &onnx.Output{Name: "output", Width: 10, Logits: &logits},
		},
		LabelFile: "labels.txt", CanonicalOrder: true, Reduction: nsfw.ReductionSoftmaxUnsafe,
		UnsafeClassIndex: &unsafeIndex, NeutralClassIndex: &neutralIndex, DefaultThreshold: 0.9,
		Options: &ModelOptions{Temperature: 0.1, Stop: []string{"stop"}}, Service: Service{Uri: "https://example.com"},
		Path: "models/name", Disabled: true,
	}
	clone := source.Clone()
	require.NotNil(t, clone)

	sourceValue := reflect.ValueOf(source).Elem()
	cloneValue := reflect.ValueOf(clone).Elem()
	modelType := sourceValue.Type()
	for i := range modelType.NumField() {
		field := modelType.Field(i)
		if !field.IsExported() || field.Tag.Get("yaml") == "-" {
			continue
		}

		assert.False(t, sourceValue.Field(i).IsZero(), "fixture must populate Model.%s", field.Name)
		assert.Equal(t, sourceValue.Field(i).Interface(), cloneValue.Field(i).Interface(), field.Name)
	}

	require.NotSame(t, source.UnsafeClassIndex, clone.UnsafeClassIndex)
	require.NotSame(t, source.NeutralClassIndex, clone.NeutralClassIndex)
	require.NotSame(t, source.TensorFlow, clone.TensorFlow)
	require.NotSame(t, source.TensorFlow.Input, clone.TensorFlow.Input)
	require.NotSame(t, source.TensorFlow.Output, clone.TensorFlow.Output)
	require.NotSame(t, source.TensorFlow.Input.Intervals[0].Mean, clone.TensorFlow.Input.Intervals[0].Mean)
	require.NotSame(t, source.TensorFlow.Input.Intervals[0].StdDev, clone.TensorFlow.Input.Intervals[0].StdDev)
	require.NotSame(t, source.ONNX, clone.ONNX)
	require.NotSame(t, source.ONNX.Input, clone.ONNX.Input)
	require.NotSame(t, source.ONNX.Output, clone.ONNX.Output)
	require.NotSame(t, source.ONNX.Output.Logits, clone.ONNX.Output.Logits)
	require.NotSame(t, source.Options, clone.Options)

	clone.TensorFlow.Tags[0] = "changed"
	clone.TensorFlow.Input.Intervals[0].Start = -1
	clone.TensorFlow.Input.Shape[0] = tensorflow.ShapeColor
	clone.ONNX.Input.Width = 512
	clone.Options.Stop[0] = "changed"
	assert.Equal(t, "serve", source.TensorFlow.Tags[0])
	assert.Zero(t, source.TensorFlow.Input.Intervals[0].Start)
	assert.Equal(t, tensorflow.ShapeBatch, source.TensorFlow.Input.Shape[0])
	assert.Equal(t, 224, source.ONNX.Input.Width)
	assert.Equal(t, "stop", source.Options.Stop[0])
}

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
		{
			name: "OpenAIKeepsQuantSuffix",
			model: &Model{
				Model:  "unsloth/Qwen3.5-9B-GGUF:Q4_K_M",
				Engine: openai.EngineName,
			},
			wantModel:   "unsloth/Qwen3.5-9B-GGUF:Q4_K_M",
			wantName:    "unsloth/Qwen3.5-9B-GGUF:Q4_K_M",
			wantVersion: "",
		},
		{
			name: "OpenAIKeepsFineTuneId",
			model: &Model{
				Engine:  openai.EngineName,
				Service: Service{Model: "ft:gpt-4o-mini-2024-07-18:acme::A1b2C3d4"},
			},
			wantModel:   "ft:gpt-4o-mini-2024-07-18:acme::A1b2C3d4",
			wantName:    "ft:gpt-4o-mini-2024-07-18:acme::A1b2C3d4",
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

func TestModel_GetModelRequestEngine(t *testing.T) {
	t.Run("EngineLessOpenAI", func(t *testing.T) {
		m := &Model{Model: "unsloth/Qwen3.5-9B-GGUF:Q4_K_M", Service: Service{RequestFormat: ApiFormatOpenAI}}
		model, name, version := m.GetModel()
		assert.Equal(t, "unsloth/Qwen3.5-9B-GGUF:Q4_K_M", model)
		assert.Equal(t, "unsloth/Qwen3.5-9B-GGUF:Q4_K_M", name)
		assert.Equal(t, "", version)
	})
	t.Run("EngineLessOpenAIServiceModel", func(t *testing.T) {
		t.Setenv("VISION_TEST_FT_MODEL", "ft:gpt-4o-mini-2024-07-18:acme::A1b2C3d4")
		m := &Model{Name: "ignored", Service: Service{Model: "${VISION_TEST_FT_MODEL}", RequestFormat: ApiFormatOpenAI}}
		model, name, version := m.GetModel()
		assert.Equal(t, "ft:gpt-4o-mini-2024-07-18:acme::A1b2C3d4", model)
		assert.Equal(t, "ft:gpt-4o-mini-2024-07-18:acme::A1b2C3d4", name)
		assert.Equal(t, "", version)
	})
	t.Run("MixedCaseEngine", func(t *testing.T) {
		model, _, version := (&Model{Model: "qwen3-vl:8b", Engine: "OpenAI"}).GetModel()
		assert.Equal(t, "qwen3-vl:8b", model)
		assert.Equal(t, "", version)
	})
	t.Run("EngineWinsOverFormat", func(t *testing.T) {
		m := &Model{Name: "gemma3:27b", Engine: ollama.EngineName, Service: Service{RequestFormat: ApiFormatOpenAI}}
		model, name, version := m.GetModel()
		assert.Equal(t, "gemma3:27b", model)
		assert.Equal(t, "gemma3", name)
		assert.Equal(t, "27b", version)
	})
	t.Run("DisabledServiceFormatIgnored", func(t *testing.T) {
		m := &Model{Name: "gemma3:27b", Service: Service{RequestFormat: ApiFormatOpenAI, Disabled: true}}
		model, name, version := m.GetModel()
		assert.Equal(t, "gemma3", model)
		assert.Equal(t, "gemma3", name)
		assert.Equal(t, "27b", version)
	})
	t.Run("EngineLessOllama", func(t *testing.T) {
		m := &Model{Name: "gemma3:27b", Service: Service{RequestFormat: ApiFormatOllama}}
		model, name, version := m.GetModel()
		assert.Equal(t, "gemma3:27b", model)
		assert.Equal(t, "gemma3", name)
		assert.Equal(t, "27b", version)
	})
	t.Run("EngineLessOllamaAddsLatest", func(t *testing.T) {
		m := &Model{Name: "gemma3", Service: Service{RequestFormat: ApiFormatOllama}}
		model, name, version := m.GetModel()
		assert.Equal(t, "gemma3:latest", model)
		assert.Equal(t, "gemma3", name)
		assert.Equal(t, "latest", version)
	})
	t.Run("MixedCaseOllamaEngine", func(t *testing.T) {
		model, _, _ := (&Model{Name: "gemma3:27b", Engine: "Ollama"}).GetModel()
		assert.Equal(t, "gemma3:27b", model)
	})
	t.Run("VisionSplitsVersion", func(t *testing.T) {
		for _, m := range []*Model{
			{Name: "custom:v2", Engine: EngineVision},
			{Name: "custom:v2", Service: Service{RequestFormat: ApiFormatVision}},
		} {
			model, name, version := m.GetModel()
			assert.Equal(t, "custom", model)
			assert.Equal(t, "custom", name)
			assert.Equal(t, "v2", version)
		}
	})
	t.Run("OllamaHuggingFaceId", func(t *testing.T) {
		m := &Model{Name: "hf.co/unsloth/Qwen3.5-9B-GGUF:Q4_K_M", Engine: ollama.EngineName}
		model, name, version := m.GetModel()
		assert.Equal(t, "hf.co/unsloth/Qwen3.5-9B-GGUF:Q4_K_M", model)
		assert.Equal(t, "hf.co/unsloth/Qwen3.5-9B-GGUF", name)
		assert.Equal(t, "Q4_K_M", version)
	})
	t.Run("EngineDefaults", func(t *testing.T) {
		for engine, want := range map[string]string{openai.EngineName: "gpt-5-mini", ollama.EngineName: "gemma4:latest"} {
			m := &Model{Engine: engine}
			m.ApplyEngineDefaults()
			model, _, _ := m.GetModel()
			assert.Equal(t, want, model)
		}
	})
}

func TestModel_RequestEngine(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", (*Model)(nil).requestEngine())
	})
	t.Run("Engine", func(t *testing.T) {
		assert.Equal(t, openai.EngineName, (&Model{Engine: " OpenAI "}).requestEngine())
		assert.Equal(t, ollama.EngineName, (&Model{Engine: ollama.EngineName, Service: Service{RequestFormat: ApiFormatOpenAI}}).requestEngine())
		assert.Equal(t, EngineVision, (&Model{Engine: EngineVision}).requestEngine())
	})
	t.Run("RequestFormat", func(t *testing.T) {
		assert.Equal(t, openai.EngineName, (&Model{Service: Service{RequestFormat: ApiFormatOpenAI}}).requestEngine())
		assert.Equal(t, ollama.EngineName, (&Model{Service: Service{RequestFormat: ApiFormatOllama}}).requestEngine())
	})
	t.Run("NoEndpointResolution", func(t *testing.T) {
		resetUnresolvedUriWarnings(t)
		logHook, _ := captureLogs(t)
		m := &Model{Type: ModelTypeLabels, Name: "custom", Service: Service{Uri: "${VISION_TEST_MISSING_URI}", RequestFormat: ApiFormatOpenAI}}
		assert.Equal(t, openai.EngineName, m.requestEngine())
		m.GetModel()
		assert.Empty(t, logHook.AllEntries())
		m.EngineName()
		assert.NotEmpty(t, logHook.AllEntries())
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, "", (&Model{}).requestEngine())
		assert.Equal(t, "", (&Model{Service: Service{RequestFormat: ApiFormatVision}}).requestEngine())
		assert.Equal(t, "", (&Model{Service: Service{RequestFormat: "custom"}}).requestEngine())
		assert.Equal(t, "", (&Model{Service: Service{RequestFormat: ApiFormatOpenAI, Disabled: true}}).requestEngine())
	})
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
		useSharedService(t, "https://vision.example.com/api/v1/vision", "${GLOBAL_KEY}")
		t.Setenv("GLOBAL_KEY", "global-secret")

		model := &Model{Type: ModelTypeCaption}
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

// useSharedService sets the shared service URI and key for the duration of the test.
func useSharedService(t *testing.T, uri, key string) {
	t.Helper()

	prevUri, prevKey := ServiceUri, ServiceKey
	t.Cleanup(func() { ServiceUri, ServiceKey = prevUri, prevKey })
	ServiceUri, ServiceKey = uri, key
}

// clearEngineKeys unsets the OpenAI and Ollama key variables for the duration of the test.
func clearEngineKeys(t *testing.T) {
	t.Helper()

	t.Cleanup(func() { ensureEnvOnce = sync.Once{} })
	t.Setenv(openai.APIKeyEnv, "")
	t.Setenv(openai.APIKeyFileEnv, "")
	t.Setenv(ollama.APIKeyEnv, "")
	t.Setenv(ollama.APIKeyFileEnv, "")
	ensureEnvOnce = sync.Once{}
}

// TestModelEndpointKey checks that the shared key is only returned for models that use the shared service.
func TestModelEndpointKey(t *testing.T) {
	const sharedUri = "https://vision.example.com/api/v1/vision"
	const sharedKey = "shared-vision-key"
	const ownUri = "https://models.example.com/api/generate"

	cases := []struct {
		name    string
		model   *Model
		wantUri string
		wantKey string
	}{
		{name: "SharedService", model: &Model{Type: ModelTypeLabels}, wantUri: sharedUri + "/labels", wantKey: sharedKey},
		{name: "SharedServiceOwnKey", model: &Model{Type: ModelTypeLabels, Service: Service{Key: "own-key"}}, wantUri: sharedUri + "/labels", wantKey: "own-key"},
		{name: "OwnEndpoint", model: &Model{Type: ModelTypeLabels, Service: Service{Uri: ownUri}}, wantUri: ownUri, wantKey: ""},
		{name: "OwnEndpointOwnKey", model: &Model{Type: ModelTypeLabels, Service: Service{Uri: ownUri, Key: "own-key"}}, wantUri: ownUri, wantKey: "own-key"},
		{name: "OwnEndpointUnresolvedKey", model: &Model{Type: ModelTypeLabels, Service: Service{Uri: ownUri, Key: "${VISION_TEST_MISSING_KEY}"}}, wantUri: ownUri, wantKey: ""},
		{name: "UnresolvedEndpoint", model: &Model{Type: ModelTypeLabels, Service: Service{Uri: "${VISION_TEST_MISSING_URI}"}}, wantUri: "", wantKey: ""},
		{name: "UnresolvedEndpointBasicAuth", model: &Model{Type: ModelTypeLabels, Service: Service{Uri: "${VISION_TEST_MISSING_URI}", Username: "user", Password: "secret"}}, wantUri: "", wantKey: ""},
		{name: "UnresolvedEndpointOwnKey", model: &Model{Type: ModelTypeLabels, Service: Service{Uri: "${VISION_TEST_MISSING_URI}", Key: "own-key"}}, wantUri: "", wantKey: ""},
		{name: "UnresolvedEngineEndpoint", model: &Model{Type: ModelTypeCaption, Engine: openai.EngineName, Service: Service{Uri: "${VISION_TEST_MISSING_URI}", Key: "own-key"}}, wantUri: "", wantKey: ""},
		{name: "PartlyUnresolvedEndpoint", model: &Model{Type: ModelTypeLabels, Service: Service{Uri: "${VISION_TEST_MISSING_URI}/api/generate"}}, wantUri: "/api/generate", wantKey: ""},
		{name: "WhitespaceUriOwnKey", model: &Model{Type: ModelTypeLabels, Service: Service{Uri: "   ", Key: "own-key"}}, wantUri: sharedUri + "/labels", wantKey: "own-key"},
		{name: "DisabledServiceOwnKey", model: &Model{Type: ModelTypeLabels, Service: Service{Key: "own-key", Disabled: true}}, wantUri: sharedUri + "/labels", wantKey: sharedKey},
		{name: "DisabledService", model: &Model{Type: ModelTypeLabels, Service: Service{Uri: ownUri, Key: "own-key", Disabled: true}}, wantUri: sharedUri + "/labels", wantKey: sharedKey},
		{name: "OllamaEngine", model: &Model{Type: ModelTypeCaption, Engine: ollama.EngineName, Service: Service{Uri: ownUri}}, wantUri: ownUri, wantKey: ""},
		{name: "OpenAIEngine", model: &Model{Type: ModelTypeCaption, Engine: openai.EngineName}, wantUri: "https://api.openai.com/v1/responses", wantKey: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			useSharedService(t, sharedUri, sharedKey)
			clearEngineKeys(t)

			if tc.model.Engine != "" {
				tc.model.ApplyEngineDefaults()
			}

			uri, _ := tc.model.Endpoint()
			assert.Equal(t, tc.wantUri, uri)
			assert.Equal(t, tc.wantKey, tc.model.EndpointKey())
		})
	}
	t.Run("NoSharedService", func(t *testing.T) {
		useSharedService(t, "", sharedKey)

		model := &Model{Type: ModelTypeLabels}
		uri, _ := model.Endpoint()
		assert.Empty(t, uri)
		assert.Empty(t, model.EndpointKey())
	})
	t.Run("NoType", func(t *testing.T) {
		useSharedService(t, sharedUri, sharedKey)

		model := &Model{}
		uri, _ := model.Endpoint()
		assert.Empty(t, uri)
		assert.Empty(t, model.EndpointKey())
	})
	t.Run("NilModel", func(t *testing.T) {
		useSharedService(t, sharedUri, sharedKey)

		var model *Model
		assert.Empty(t, model.EndpointKey())
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
	t.Run("EngineLessOpenAIHeaders", func(t *testing.T) {
		req := &ApiRequest{}
		model := &Model{Service: Service{RequestFormat: ApiFormatOpenAI, Org: "org-123", Project: "proj-abc", Tier: "flex"}}

		model.ApplyService(req)

		assert.Equal(t, "org-123", req.Org)
		assert.Equal(t, "proj-abc", req.Project)
		assert.Equal(t, "flex", req.Tier)
	})
	t.Run("EngineWinsOverFormat", func(t *testing.T) {
		req := &ApiRequest{}
		model := &Model{Engine: ollama.EngineName, Service: Service{RequestFormat: ApiFormatOpenAI, Org: "org-123", Project: "proj-abc", Tier: "flex"}}

		model.ApplyService(req)

		assert.Equal(t, "", req.Org)
		assert.Equal(t, "", req.Project)
		assert.Equal(t, "", req.Tier)
	})
	t.Run("EngineLessOllamaIgnoresOpenAIHeaders", func(t *testing.T) {
		req := &ApiRequest{}
		model := &Model{Service: Service{RequestFormat: ApiFormatOllama, Org: "org-123", Project: "proj-abc", Tier: "flex"}}

		model.ApplyService(req)

		assert.Equal(t, "", req.Org)
		assert.Equal(t, "", req.Project)
		assert.Equal(t, "", req.Tier)
	})
}

func TestModel_IsDefault(t *testing.T) {
	defaultCopy := DefaultLabelModel.Clone() //nolint:govet // copy for test inspection only
	defaultCopy.Default = false

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
			model: defaultCopy,
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

// TestModel_ClassifyModelMissingRegisteredCachesError verifies failed initialization preserves operator disablement.
func TestModel_ClassifyModelMissingRegisteredCachesError(t *testing.T) {
	previousModelsPath := ModelsPath
	ModelsPath = t.TempDir()
	t.Cleanup(func() { ModelsPath = previousModelsPath })

	model := NewLabelModel(classify.ModelRepViTM10)
	require.NotNil(t, model)
	assert.Nil(t, model.ClassifyModel())
	assert.False(t, model.Disabled)
	require.Error(t, model.classifyErr)
	assert.Nil(t, model.ClassifyModel())
	model.DisabledByMode = true
	clone := model.Clone()
	assert.Nil(t, clone.classifyErr)
	assert.False(t, clone.DisabledByMode)
}

// TestModelOnnxProvider verifies registered and custom local models use the global provider.
func TestModelOnnxProvider(t *testing.T) {
	previousProvider := OnnxProvider
	OnnxProvider = onnx.ProviderCUDA
	t.Cleanup(func() { OnnxProvider = previousProvider })

	registeredLabels := NewLabelModel(classify.DefaultModelName())
	require.NotNil(t, registeredLabels)
	registeredLabels.Disabled = true
	require.NotNil(t, registeredLabels.ClassifyModel())
	assert.Equal(t, onnx.ProviderCUDA, registeredLabels.ClassifyModel().Provider())

	customLabels := &Model{Type: ModelTypeLabels, Name: "custom-labels", Path: "custom-labels.onnx", ONNX: &onnx.ModelInfo{}, Disabled: true}
	require.NotNil(t, customLabels.ClassifyModel())
	assert.Equal(t, onnx.ProviderCUDA, customLabels.ClassifyModel().Provider())

	registeredNSFW := NewNsfwModel(nsfw.DefaultModelName())
	require.NotNil(t, registeredNSFW)
	registeredNSFW.Disabled = true
	require.NotNil(t, registeredNSFW.NsfwModel())
	assert.Equal(t, onnx.ProviderCUDA, registeredNSFW.NsfwModel().Provider())

	customNSFW := &Model{Type: ModelTypeNsfw, Name: "custom-nsfw", Path: "custom-nsfw.onnx", ONNX: &onnx.ModelInfo{}, Disabled: true}
	require.NotNil(t, customNSFW.NsfwModel())
	assert.Equal(t, onnx.ProviderCUDA, customNSFW.NsfwModel().Provider())
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
		assert.Contains(t, buffer.String(), "PHOTOPRISM_FACE_MODEL")
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

// TestModel_EndpointUnresolved checks that a model whose own service URI does not resolve has no endpoint.
func TestModel_EndpointUnresolved(t *testing.T) {
	t.Run("WarnsOncePerModel", func(t *testing.T) {
		useSharedService(t, "https://vision.example.com/api/v1/vision", "shared-vision-key")
		logHook, systemHook := captureLogs(t)
		resetUnresolvedUriWarnings(t)

		labels := &Model{Type: ModelTypeLabels, Name: "custom", Service: Service{Uri: "${VISION_TEST_MISSING_URI}", Username: "user", Password: "pass"}}
		caption := &Model{Type: ModelTypeCaption, Model: "gemma3:4b", Service: Service{Uri: "${VISION_TEST_MISSING_URI}"}}

		for range 3 {
			for _, model := range []*Model{labels, caption} {
				uri, method := model.Endpoint()
				assert.Empty(t, uri)
				assert.Empty(t, method)
			}
		}

		require.Len(t, logHook.AllEntries(), 2)
		assert.Equal(t, logrus.WarnLevel, logHook.AllEntries()[0].Level)
		assert.Equal(t, "vision: service uri of labels model does not resolve, so no service is used (details in system log)", logHook.AllEntries()[0].Message)
		assert.Equal(t, "vision: service uri of caption model does not resolve, so no service is used (details in system log)", logHook.AllEntries()[1].Message)
		require.Len(t, systemHook.AllEntries(), 2)
		assert.Equal(t, logrus.WarnLevel, systemHook.AllEntries()[0].Level)
		assert.Equal(t, "vision: service uri of labels model custom does not resolve", systemHook.AllEntries()[0].Message)
		assert.Equal(t, "vision: service uri of caption model gemma3 does not resolve", systemHook.AllEntries()[1].Message)
		assert.NotContains(t, systemHook.AllEntries()[0].Message, "pass")
	})
	t.Run("ResolvedOrBlank", func(t *testing.T) {
		useSharedService(t, "https://vision.example.com/api/v1/vision", "shared-vision-key")
		logHook, _ := captureLogs(t)
		resetUnresolvedUriWarnings(t)

		uri, _ := (&Model{Type: ModelTypeLabels}).Endpoint()
		assert.Equal(t, "https://vision.example.com/api/v1/vision/labels", uri)
		uri, _ = (&Model{Type: ModelTypeLabels, Service: Service{Uri: "https://models.example.com/api"}}).Endpoint()
		assert.Equal(t, "https://models.example.com/api", uri)
		uri, _ = (&Model{Type: ModelTypeLabels, Service: Service{Uri: "${VISION_TEST_MISSING_URI}", Disabled: true}}).Endpoint()
		assert.Equal(t, "https://vision.example.com/api/v1/vision/labels", uri)
		assert.Empty(t, logHook.AllEntries())
	})
}

// TestModel_UnresolvedUriErr checks the error for a model whose own service URI does not resolve.
func TestModel_UnresolvedUriErr(t *testing.T) {
	t.Run("Unresolved", func(t *testing.T) {
		logHook, systemHook := captureLogs(t)
		resetUnresolvedUriWarnings(t)

		model := &Model{Type: ModelTypeNsfw, Name: "qwen3-vl:4b", Engine: ollama.EngineName, Service: Service{Uri: "${VISION_TEST_MISSING_URI}"}}
		assert.EqualError(t, model.unresolvedUriErr(), "service uri of nsfw model does not resolve")
		assert.EqualError(t, model.unresolvedUriErr(), "service uri of nsfw model does not resolve")
		require.Len(t, logHook.AllEntries(), 1)
		assert.Equal(t, "vision: service uri of nsfw model does not resolve, so no service is used (details in system log)", logHook.LastEntry().Message)
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, "vision: service uri of nsfw model qwen3-vl:4b does not resolve", systemHook.LastEntry().Message)
	})
	t.Run("ServiceModelFromEnv", func(t *testing.T) {
		logHook, systemHook := captureLogs(t)
		resetUnresolvedUriWarnings(t)
		t.Setenv("VISION_TEST_SERVICE_MODEL", "private/vision-model:q4")

		model := &Model{Type: ModelTypeCaption, Engine: ollama.EngineName, Service: Service{Uri: "${VISION_TEST_MISSING_URI}", Model: "${VISION_TEST_SERVICE_MODEL}"}}
		assert.EqualError(t, model.unresolvedUriErr(), "service uri of caption model does not resolve")
		require.Len(t, logHook.AllEntries(), 1)
		for _, entry := range logHook.AllEntries() {
			assert.NotContains(t, entry.Message, "private/vision-model")
		}
		require.Len(t, systemHook.AllEntries(), 1)
		assert.Contains(t, systemHook.LastEntry().Message, "private/vision-model:q4")
	})
	t.Run("Resolved", func(t *testing.T) {
		assert.NoError(t, (&Model{Type: ModelTypeLabels, Service: Service{Uri: "https://models.example.com/api"}}).unresolvedUriErr())
		assert.NoError(t, (&Model{Type: ModelTypeLabels}).unresolvedUriErr())
		var model *Model
		assert.NoError(t, model.unresolvedUriErr())
	})
}

// TestModel_WarnUnresolvedUri checks that the warning for an unresolved service URI is logged once per model.
func TestModel_WarnUnresolvedUri(t *testing.T) {
	logHook, systemHook := captureLogs(t)
	resetUnresolvedUriWarnings(t)

	model := &Model{Type: ModelTypeLabels, Name: "custom", Service: Service{Uri: "${VISION_TEST_MISSING_URI}"}}
	model.warnUnresolvedUri()
	model.warnUnresolvedUri()
	(&Model{Type: ModelTypeLabels, Name: "other", Service: Service{Uri: "${VISION_TEST_MISSING_URI}"}}).warnUnresolvedUri()

	require.Len(t, logHook.AllEntries(), 2)
	require.Len(t, systemHook.AllEntries(), 2)
}

// TestService_UriUnresolved checks which service URIs count as unresolved.
func TestService_UriUnresolved(t *testing.T) {
	t.Run("Unresolved", func(t *testing.T) {
		assert.True(t, (&Service{Uri: "${VISION_TEST_MISSING_URI}"}).UriUnresolved())
		assert.True(t, (&Service{Uri: " ${VISION_TEST_MISSING_URI} "}).UriUnresolved())
	})
	t.Run("NotUnresolved", func(t *testing.T) {
		assert.False(t, (&Service{}).UriUnresolved())
		assert.False(t, (&Service{Uri: "   "}).UriUnresolved())
		assert.False(t, (&Service{Uri: "https://models.example.com/api"}).UriUnresolved())
		assert.False(t, (&Service{Uri: "${VISION_TEST_MISSING_URI}", Disabled: true}).UriUnresolved())
		var service *Service
		assert.False(t, service.UriUnresolved())
	})
}

// resetUnresolvedUriWarnings clears the models warned about before and after a test.
func resetUnresolvedUriWarnings(t *testing.T) {
	t.Helper()
	unresolvedUriWarned.Clear()
	t.Cleanup(unresolvedUriWarned.Clear)
}

// TestCustomClassifyInitializationError verifies custom failures preserve saved disablement.
func TestCustomClassifyInitializationError(t *testing.T) {
	previous := ModelsPath
	ModelsPath = t.TempDir()
	t.Cleanup(func() { ModelsPath = previous })
	model := &Model{Type: ModelTypeLabels, Name: "custom", Path: "custom/model.onnx", ONNX: &onnx.ModelInfo{}}
	hook := captureVisionLog(t)
	assert.Nil(t, model.ClassifyModel())
	require.Error(t, model.classifyErr)
	assert.False(t, model.Disabled)
	cached := model.classifyErr
	assert.Nil(t, model.ClassifyModel())
	assert.Same(t, cached, model.classifyErr)
	assert.Len(t, initWarnings(hook.AllEntries()), 1)
}
