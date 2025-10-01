package vision

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestModelGetOptionsDefaultsOllamaLabels(t *testing.T) {
	ollamaModel := "redule26/huihui_ai_qwen2.5-vl-7b-abliterated:latest"

	model := &Model{
		Type:   ModelTypeLabels,
		Name:   ollamaModel,
		Engine: ollama.EngineName,
	}

	model.ApplyEngineDefaults()

	m, n, v := model.Model()

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

func TestModel_IsDefault(t *testing.T) {
	nasnetCopy := *NasnetModel
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
