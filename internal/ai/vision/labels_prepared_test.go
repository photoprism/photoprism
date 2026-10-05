package vision

import (
	"errors"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestUsesPreparedLabels verifies that only local S2 models with the default input size opt in.
func TestUsesPreparedLabels(t *testing.T) {
	for _, tc := range []struct {
		name  string
		model *Model
		want  bool
	}{
		{"S2", NewLabelModel(classify.ModelEfficientFormerV2S2), true},
		{"OtherRegistered", NewLabelModel(classify.ModelRepViTM10), false},
		{"Custom", &Model{Type: ModelTypeLabels, Name: "custom", Resolution: 224}, false},
		{"OwnResolution", &Model{Type: ModelTypeLabels, Name: string(classify.ModelEfficientFormerV2S2), Resolution: 384}, false},
		{"Remote", &Model{Type: ModelTypeLabels, Name: string(classify.ModelEfficientFormerV2S2), Service: Service{Uri: "http://localhost:5000"}}, false},
		{"Engine", &Model{Type: ModelTypeLabels, Name: string(classify.ModelEfficientFormerV2S2), Engine: "ollama"}, false},
		{"Unresolved", &Model{Type: ModelTypeLabels, Name: string(classify.ModelEfficientFormerV2S2), Service: Service{Uri: "${LABEL_TEST_MISSING_URI}"}}, false},
		{"Nil", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) { assert.Equal(t, tc.want, tc.model.UsesPreparedLabels()) })
	}
}

// TestPreparedLabels verifies call counts, configured floors, and the existing union semantics.
func TestPreparedLabels(t *testing.T) {
	input := classify.Input{Image: image.NewNRGBA(image.Rect(0, 0, 224, 224)), Prepared: true}
	for _, tc := range []struct {
		name             string
		count, threshold int
	}{
		{"Square", 1, 20}, {"NonSquare", 2, 20}, {"ConfiguredFloor", 2, 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			inputs := make([]classify.Input, tc.count)
			for i := range inputs {
				inputs[i] = input
			}
			labels, err := preparedLabels(inputs, entity.SrcManual, tc.threshold, func(got classify.Input, floor int) (classify.Labels, error) {
				calls++
				assert.Same(t, input.Image, got.Image)
				assert.True(t, got.Prepared)
				assert.Equal(t, tc.threshold, floor)
				return classify.Labels{{Name: "dog", Uncertainty: 30 - calls, Priority: calls, Categories: []string{"animal"}}}, nil
			})
			require.NoError(t, err)
			assert.Equal(t, tc.count, calls)
			require.Len(t, labels, 1)
			assert.Equal(t, 30-tc.count, labels[0].Uncertainty)
			assert.Equal(t, tc.count, labels[0].Priority)
			assert.Equal(t, entity.SrcManual, labels[0].Source)
			assert.Equal(t, []string{"animal"}, labels[0].Categories)
		})
	}
	t.Run("Error", func(t *testing.T) {
		calls := 0
		_, err := preparedLabels([]classify.Input{input, input}, entity.SrcImage, 20, func(classify.Input, int) (classify.Labels, error) {
			calls++
			return nil, errors.New("prediction failed")
		})
		require.ErrorContains(t, err, "prediction failed")
		assert.Equal(t, 1, calls)
	})
	t.Run("InvalidCount", func(t *testing.T) {
		_, err := preparedLabels(nil, entity.SrcImage, 20, nil)
		require.Error(t, err)
	})
	t.Run("Unprepared", func(t *testing.T) {
		_, err := preparedLabels([]classify.Input{{}}, entity.SrcImage, 20, nil)
		require.Error(t, err)
	})
}

// TestGeneratePreparedLabelsConfiguration verifies missing and incompatible configurations.
func TestGeneratePreparedLabelsConfiguration(t *testing.T) {
	original := Config
	t.Cleanup(func() { Config = original })
	for _, cfg := range []*ConfigValues{nil, {}, {Models: Models{{Type: ModelTypeLabels, Name: "custom"}}}} {
		Config = cfg
		_, err := GeneratePreparedLabels(nil, entity.SrcAuto)
		require.Error(t, err)
	}
	assert.Equal(t, 20, DefaultThresholds.Confidence)
}

// TestLabelConfidenceDefault verifies config loading preserves explicit confidence floors.
func TestLabelConfidenceDefault(t *testing.T) {
	for _, tc := range []struct {
		name, yaml string
		want       int
	}{
		{"Default", "Models: []\n", 20},
		{"Explicit", "Thresholds:\n  Confidence: 10\n", 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "vision.yml")
			require.NoError(t, os.WriteFile(path, []byte(tc.yaml), fs.ModeConfigFile))
			cfg := NewConfig()
			require.NoError(t, cfg.Load(path))
			assert.Equal(t, tc.want, cfg.Thresholds.Confidence)
			assert.Equal(t, DefaultNSFWThreshold, cfg.Thresholds.NSFW)
		})
	}
}

// TestGeneratePreparedLabelsThreshold verifies the configured floor reaches real inference.
func TestGeneratePreparedLabelsThreshold(t *testing.T) {
	description := classify.FindModel(classify.ModelEfficientFormerV2S2)
	if !description.Installed(GetModelsPath()) {
		t.Skip("S2 model is not installed")
	}
	previous := Config
	model := NewLabelModel(classify.ModelEfficientFormerV2S2)
	Config = &ConfigValues{Models: Models{model}, Thresholds: Thresholds{Confidence: 20}}
	t.Cleanup(func() {
		Config = previous
		if classifier := model.ClassifyModel(); classifier != nil {
			require.NoError(t, classifier.Close())
		}
	})
	img, _, err := fs.DecodeImageFile(filepath.Join(samplesPath, "cat_224.jpeg"))
	require.NoError(t, err)
	input := classify.Input{Image: img, Prepared: true}
	labels, err := GeneratePreparedLabels([]classify.Input{input}, entity.SrcManual)
	require.NoError(t, err)
	require.NotEmpty(t, labels)
	assert.Equal(t, entity.SrcManual, labels[0].Source)
	for _, label := range labels {
		require.Greater(t, label.Uncertainty, 0, "fixture must distinguish floor 20 from 100")
	}
	Config.Thresholds.Confidence = 100
	labels, err = GeneratePreparedLabels([]classify.Input{input}, entity.SrcManual)
	require.NoError(t, err)
	assert.Empty(t, labels)
}
